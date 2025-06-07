package shutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ShutdownHandler 优雅退出处理器
type ShutdownHandler interface {
	Shutdown(ctx context.Context) error
	Name() string
}

// ShutdownManager 优雅退出管理器
type ShutdownManager struct {
	handlers []ShutdownHandler
	timeout  time.Duration
	mu       sync.RWMutex
}

// NewShutdownManager 创建优雅退出管理器
func NewShutdownManager(timeout time.Duration) *ShutdownManager {
	return &ShutdownManager{
		handlers: make([]ShutdownHandler, 0),
		timeout:  timeout,
	}
}

// Register 注册需要优雅退出的组件
func (m *ShutdownManager) Register(handler ShutdownHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
	log.Printf("注册优雅退出处理函数: %s", handler.Name())
}

// Start 开始监听程序终止信号
func (m *ShutdownManager) Start() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-quit
		log.Printf("接收到了程序终止信号: %v", sig)
		m.shutdown()
		os.Exit(0)
	}()
}

// shutdown 执行所有组件的优雅退出
func (m *ShutdownManager) shutdown() {
	m.mu.RLock()
	handlers := make([]ShutdownHandler, len(m.handlers))
	copy(handlers, m.handlers)
	m.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	// 并发执行所有组件的停止操作
	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h ShutdownHandler) {
			defer wg.Done()

			log.Printf("正在终止: %s", h.Name())
			if err := h.Shutdown(ctx); err != nil {
				log.Printf("异常终止 %s: %v", h.Name(), err)
			} else {
				log.Printf("正常终止: %s", h.Name())
			}
		}(handler)
	}

	// 等待所有组件停止完成或超时
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("所有组件都已经优雅退出...")
	case <-ctx.Done():
		log.Println("程序优雅退出超时，强制退出...")
	}
}
