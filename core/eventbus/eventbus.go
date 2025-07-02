package eventbus

import (
	"docker-certs/core/types"
	"sync"
)

var (
	bus  *EventBus
	once sync.Once
)

type ListenerHandle any

type EventBus struct {
	mu        sync.RWMutex
	listeners map[types.EventType][]ListenerHandle
}

func Get() *EventBus {
	once.Do(func() {
		bus = &EventBus{listeners: make(map[types.EventType][]ListenerHandle)}
	})
	return bus
}

func On[T any](et types.EventType, fn func(e types.Event[T])) {
	bus := Get()
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.listeners[et] = append(bus.listeners[et], ListenerHandle(fn))
}

func Publish[T any](event types.Event[T]) {
	bus := Get()
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	if listeners, ok := bus.listeners[event.Type]; ok {
		for _, listener := range listeners {
			fn := listener.(func(event2 types.Event[T]))
			go fn(event)
		}
	}
}

func Close() {
	bus := Get()
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.listeners = make(map[types.EventType][]ListenerHandle)
}
