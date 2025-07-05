package types

type EventType string

const (
	ContainerStarted EventType = "docker.container.started"
	ContainerStopped EventType = "docker.container.stopped"
	CertCreated      EventType = "cert.created"
)

type Event[T any] struct {
	Type    EventType
	Payload T
}
