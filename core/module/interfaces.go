package module

import "docker-certs/core/eventbus"

type Initializable interface {
	Init() error
}

type EventAware interface {
	RegisterListeners(bus *eventbus.EventBus)
}

type Closable interface {
	Close() error
}
