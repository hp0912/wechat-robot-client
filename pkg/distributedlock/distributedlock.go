package distributedlock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrLockNotAcquired = errors.New("lock not acquired")
	ErrLockNotHeld     = errors.New("lock not held by this instance")
)

// DistributedLock Redis分布式锁
type DistributedLock struct {
	client      redis.Cmdable
	key         string
	value       string
	expiration  time.Duration
	retryDelay  time.Duration
	maxRetries  int
	autoRenewal bool
	renewalCtx  context.Context
	renewalStop context.CancelFunc
}

// LockOption 锁配置选项
type LockOption func(*DistributedLock)

// WithExpiration 设置锁过期时间
func WithExpiration(expiration time.Duration) LockOption {
	return func(dl *DistributedLock) {
		dl.expiration = expiration
	}
}

// WithRetryDelay 设置重试延迟
func WithRetryDelay(delay time.Duration) LockOption {
	return func(dl *DistributedLock) {
		dl.retryDelay = delay
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) LockOption {
	return func(dl *DistributedLock) {
		dl.maxRetries = maxRetries
	}
}

// WithAutoRenewal 启用自动续期
func WithAutoRenewal() LockOption {
	return func(dl *DistributedLock) {
		dl.autoRenewal = true
	}
}

// NewDistributedLock 创建新的分布式锁实例
func NewDistributedLock(client redis.Cmdable, key string, opts ...LockOption) *DistributedLock {
	dl := &DistributedLock{
		client:      client,
		key:         fmt.Sprintf("lock:%s", key),
		value:       fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix()),
		expiration:  30 * time.Second,
		retryDelay:  100 * time.Millisecond,
		maxRetries:  3,
		autoRenewal: false,
	}

	for _, opt := range opts {
		opt(dl)
	}

	return dl
}

// Lock 获取锁
func (dl *DistributedLock) Lock(ctx context.Context) error {
	for i := 0; i <= dl.maxRetries; i++ {
		acquired, err := dl.tryLock(ctx)
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if acquired {
			if dl.autoRenewal {
				dl.startAutoRenewal()
			}
			return nil
		}

		if i < dl.maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(dl.retryDelay):
				continue
			}
		}
	}

	return ErrLockNotAcquired
}

// TryLock 尝试获取锁（不重试）
func (dl *DistributedLock) TryLock(ctx context.Context) error {
	acquired, err := dl.tryLock(ctx)
	if err != nil {
		return fmt.Errorf("failed to try lock: %w", err)
	}

	if !acquired {
		return ErrLockNotAcquired
	}

	if dl.autoRenewal {
		dl.startAutoRenewal()
	}

	return nil
}

// Unlock 释放锁
func (dl *DistributedLock) Unlock(ctx context.Context) error {
	// 停止自动续期
	if dl.renewalStop != nil {
		dl.renewalStop()
		dl.renewalStop = nil
	}

	// Lua脚本确保只有持有锁的实例才能释放锁
	script := `
        if redis.call("GET", KEYS[1]) == ARGV[1] then
            return redis.call("DEL", KEYS[1])
        else
            return 0
        end
    `

	result, err := dl.client.Eval(ctx, script, []string{dl.key}, dl.value).Result()
	if err != nil {
		return fmt.Errorf("failed to unlock: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotHeld
	}

	return nil
}

// IsLocked 检查锁是否被当前实例持有
func (dl *DistributedLock) IsLocked(ctx context.Context) (bool, error) {
	value, err := dl.client.Get(ctx, dl.key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}

	return value == dl.value, nil
}

// Extend 延长锁的过期时间
func (dl *DistributedLock) Extend(ctx context.Context, expiration time.Duration) error {
	// Lua脚本确保只有持有锁的实例才能延长过期时间
	script := `
        if redis.call("GET", KEYS[1]) == ARGV[1] then
            return redis.call("EXPIRE", KEYS[1], ARGV[2])
        else
            return 0
        end
    `

	result, err := dl.client.Eval(ctx, script, []string{dl.key}, dl.value, int(expiration.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotHeld
	}

	return nil
}

// tryLock 尝试获取锁的内部实现
func (dl *DistributedLock) tryLock(ctx context.Context) (bool, error) {
	result, err := dl.client.SetNX(ctx, dl.key, dl.value, dl.expiration).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// startAutoRenewal 启动自动续期
func (dl *DistributedLock) startAutoRenewal() {
	dl.renewalCtx, dl.renewalStop = context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(dl.expiration / 3) // 每1/3过期时间续期一次
		defer ticker.Stop()

		for {
			select {
			case <-dl.renewalCtx.Done():
				return
			case <-ticker.C:
				if err := dl.Extend(dl.renewalCtx, dl.expiration); err != nil {
					// 续期失败，可能锁已被其他实例获取
					return
				}
			}
		}
	}()
}

// WithLock 使用分布式锁执行函数的便捷方法
func WithLock(ctx context.Context, client redis.Cmdable, key string, fn func() error, opts ...LockOption) error {
	lock := NewDistributedLock(client, key, opts...)

	if err := lock.Lock(ctx); err != nil {
		return err
	}

	defer func() {
		if unlockErr := lock.Unlock(ctx); unlockErr != nil {
			// 记录解锁错误，但不影响原函数的执行结果
			fmt.Printf("Failed to unlock: %v\n", unlockErr)
		}
	}()

	return fn()
}
