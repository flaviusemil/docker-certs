package configs

type AppConfig struct {
	Name           string `map_struct:"name" validate:"required"`
	Debug          bool   `map_struct:"debug" env:"APP_DEBUG"`
	MDNSPublishing bool   `map_struct:"mDNSPublishing" env:"APP_MDNS_PUBLISHING" default:"true"`
	CertsDir       string `map_struct:"certsDir" env:"APP_CERTS_DIR" default:"certs"`
}
