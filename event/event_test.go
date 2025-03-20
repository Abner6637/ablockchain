package event

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试 EventBus 是否能正确订阅和接收消息
func TestEventBus_StringMessage(t *testing.T) {
	eb := NewEventBus[string]() // 创建支持 string 类型的 EventBus

	ch := eb.Subscribe("test_event") // 订阅事件

	go eb.Publish("test_event", "Hello, World!")

	// 设定超时时间，避免 goroutine 卡住
	select {
	case msg := <-ch:
		assert.Equal(t, "Hello, World!", msg, "接收到的消息应与发送的匹配")
	case <-time.After(time.Second):
		t.Fatal("测试超时，未能接收到消息")
	}
}

// 测试多个订阅者能否收到同一条消息
func TestEventBus_MultipleSubscribers(t *testing.T) {
	eb := NewEventBus[string]()

	ch1 := eb.Subscribe("multi_event")
	ch2 := eb.Subscribe("multi_event")

	go eb.Publish("multi_event", "Broadcast Message")

	// 接收消息
	msg1 := <-ch1
	msg2 := <-ch2

	assert.Equal(t, "Broadcast Message", msg1)
	assert.Equal(t, "Broadcast Message", msg2)
}

// 测试不同类型的数据
func TestEventBus_IntMessage(t *testing.T) {
	eb := NewEventBus[int]()

	ch := eb.Subscribe("int_event")

	go eb.Publish("int_event", 42)

	select {
	case msg := <-ch:
		assert.Equal(t, 42, msg, "接收到的消息应为 42")
	case <-time.After(time.Second):
		t.Fatal("测试超时，未能接收到消息")
	}
}

// 测试并发发布事件
func TestEventBus_ConcurrentPublish(t *testing.T) {
	eb := NewEventBus[string]()

	ch := eb.Subscribe("concurrent_event")

	var wg sync.WaitGroup
	wg.Add(2)

	// 并发发布两次事件
	go func() {
		defer wg.Done()
		eb.Publish("concurrent_event", "Message 1")
	}()
	go func() {
		defer wg.Done()
		eb.Publish("concurrent_event", "Message 2")
	}()

	wg.Wait()

	// 验证是否收到至少一个消息
	select {
	case msg := <-ch:
		assert.Contains(t, []string{"Message 1", "Message 2"}, msg, "应接收到 Message 1 或 Message 2")
		fmt.Println(msg)
	case <-time.After(time.Second):
		t.Fatal("测试超时，未能接收到消息")
	}
}
