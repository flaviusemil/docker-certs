package module

type Initializable interface {
	Init() error
}

type EventAware interface {
	RegisterEventHandlers()
}

type Closable interface {
	Close() error
}
