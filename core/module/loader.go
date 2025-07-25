package module

import (
	"log"
)

var loadedModules []Closable

func LoadModules(modules []interface{}) {
	for _, m := range modules {
		if initable, ok := m.(Initializable); ok {
			if err := initable.Init(); err != nil {
				log.Fatalf("Module init failed: %v", err)
			}
		}
		if aware, ok := m.(EventAware); ok {
			aware.RegisterEventHandlers()
		}
		if closable, ok := m.(Closable); ok {
			loadedModules = append(loadedModules, closable)
		}
	}
}

func CloseModules() {
	for _, m := range loadedModules {
		if err := m.Close(); err != nil {
			log.Printf("Module close error: %v", err)
		}
	}
}
