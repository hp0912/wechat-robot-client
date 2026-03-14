package vars

import (
	"log"
	"sync"
	"wechat-robot-client/model"
)

// SettingsChangeCallback 全局配置变更回调函数
type SettingsChangeCallback func(settings *model.GlobalSettings) error

// GlobalSettingsObserver 全局配置变更观察者，用于在全局配置发生变化时通知所有注册的回调
type GlobalSettingsObserver struct {
	mu        sync.RWMutex
	callbacks []namedCallback
}

type namedCallback struct {
	name     string
	callback SettingsChangeCallback
}

// NewGlobalSettingsObserver 创建新的全局配置观察者
func NewGlobalSettingsObserver() *GlobalSettingsObserver {
	return &GlobalSettingsObserver{}
}

// Register 注册一个配置变更回调
func (o *GlobalSettingsObserver) Register(name string, callback SettingsChangeCallback) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.callbacks = append(o.callbacks, namedCallback{name: name, callback: callback})
}

// NotifyAll 通知所有注册的回调，配置已变更
func (o *GlobalSettingsObserver) NotifyAll(settings *model.GlobalSettings) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	for _, cb := range o.callbacks {
		log.Printf("[GlobalSettings] 通知配置变更: %s", cb.name)
		if err := cb.callback(settings); err != nil {
			log.Printf("[GlobalSettings] %s 处理配置变更失败: %v", cb.name, err)
		}
	}
}

// SettingsObserver 全局配置变更观察者单例
var SettingsObserver = NewGlobalSettingsObserver()
