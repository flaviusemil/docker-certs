package configs

import (
	"github.com/spf13/viper"
	"log"
	"sync"
)

type AppConfig struct {
	Name           string `map_struct:"name" validate:"required"`
	Debug          bool   `map_struct:"debug" env:"APP_DEBUG" default:"false"`
	MDNSPublishing bool   `map_struct:"mDNSPublishing" env:"APP_MDNS_PUBLISHING" default:"true"`
	CertsDir       string `map_struct:"certsDir" env:"APP_CERTS_DIR" default:"certs"`
}

var (
	instance *AppConfig
	once     sync.Once
)

func GetConfig() *AppConfig {
	once.Do(func() {
		instance = &AppConfig{}
		viper.SetEnvPrefix("app")
		viper.AutomaticEnv()

		if err := bindEnvVarsWithDefaults(instance, ""); err != nil {
			log.Fatalf("[cmd] Failed to bind env vars: %v", err)
		}

		if err := viper.Unmarshal(instance); err != nil {
			log.Fatalf("[cmd] Unable to decode config into struct, %v", err)
		}

		log.Println("[cmd] Configs loaded.")
	})

	return instance
}
