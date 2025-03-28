package event

import (
	"sync"
)

// 事件管理器
type EventBus[T any] struct {
	subscribers map[string][]chan T // 事件类型 -> 订阅者列表
	lock        sync.RWMutex        // 保护并发访问
}

var Bus = NewEventBus[any]() //全局单例

// 创建新的事件总线
func NewEventBus[T any]() *EventBus[T] {
	return &EventBus[T]{
		subscribers: make(map[string][]chan T),
	}
}

// 订阅事件
func (eb *EventBus[T]) Subscribe(eventType string) <-chan T {
	eb.lock.Lock()
	defer eb.lock.Unlock()
	ch := make(chan T, 10) // 创建一个新的事件通道
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

// 取消订阅事件
func (eb *EventBus[T]) Unsubscribe(eventType string, ch <-chan T) {
	eb.lock.Lock()
	defer eb.lock.Unlock()

	// 找到并移除指定的通道
	if subscribers, found := eb.subscribers[eventType]; found {
		for i, sub := range subscribers {
			if sub == ch {
				// 关闭通道
				close(sub)
				// 移除该通道
				eb.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
	}
}

// 发布事件
func (eb *EventBus[T]) Publish(eventType string, message T) {
	eb.lock.RLock()
	defer eb.lock.RUnlock()
	if subscribers, found := eb.subscribers[eventType]; found {
		for _, ch := range subscribers {
			ch <- message // 发送消息到所有订阅者
		}
	}
}

type EventSubscription struct {
	Name    string     // 事件名称
	Channel <-chan any // 事件通道
}
