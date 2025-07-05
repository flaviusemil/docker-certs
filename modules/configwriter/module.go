package configwriter

import (
	"docker-certs/core/eventbus"
	"docker-certs/core/types"
	"log"
)

type Module struct {
	listeners []eventbus.ListenerHandle
}

func (m *Module) Init() error {
	log.Println("[config-writer] Initializing...")
	if err := createDynamicYAMLIfNotExists(); err != nil {
		return err
	}

	return nil
}

func (m *Module) RegisterEventHandlers() {
	log.Println("[config-writer] Registering listeners...")

	eventbus.On(types.CertCreated, updateTraefikConfig)
}
