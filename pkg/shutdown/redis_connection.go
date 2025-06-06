package shutdown

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisConnection struct {
	Client *redis.Client
}

func (r *RedisConnection) Name() string {
	return "redis 数据库连接"
}

func (r *RedisConnection) Shutdown(ctx context.Context) error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}
