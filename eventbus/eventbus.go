package eventbus

import (
	"docker-certs/types"
	"fmt"
	"github.com/docker/docker/api/types/events"
	"reflect"
	"sync"
)

var (
	bus  *EventBus
	once sync.Once
)

type EventListener func(event types.Event)

type EventBus struct {
	listeners map[events.Action][]EventListener
	mu        sync.RWMutex
}

func GetEventBus() *EventBus {
	once.Do(func() {
		bus = New()
	})
	return bus
}

func New() *EventBus {
	return &EventBus{
		listeners: make(map[events.Action][]EventListener),
	}
}

func (bus *EventBus) AutoRegisterListeners(pkgPath string, action events.Action) {
	pkg := reflect.ValueOf(pkgPath)
	pkgType := pkg.Type()

	fmt.Println(pkgType.NumMethod())

	for i := 0; i < pkgType.NumMethod(); i++ {
		method := pkgType.Method(i)
		if isValidListener(method.Func) {
			bus.RegisterListener(action, func(event types.Event) {
				method.Func.Call([]reflect.Value{reflect.ValueOf(pkg.Interface()), reflect.ValueOf(event)})
			})
		}
	}
}

func isValidListener(fn reflect.Value) bool {
	fnType := fn.Type()
	return fnType.Kind() == reflect.Func &&
		fnType.NumIn() == 1 &&
		fnType.In(0).Kind() == reflect.Struct
}

func (bus *EventBus) RegisterListener(action events.Action, listener EventListener) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if _, ok := bus.listeners[action]; !ok {
		bus.listeners[action] = []EventListener{}
	}

	bus.listeners[action] = append(bus.listeners[action], listener)
}

func (bus *EventBus) Publish(event types.Event) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	if listeners, ok := bus.listeners[event.Action]; ok {
		for _, listener := range listeners {
			listener(event)
		}
	}
}

func (bus *EventBus) Close() {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.listeners = make(map[events.Action][]EventListener)
}
