package configs

type AppConfig struct {
	Name  string `map_struct:"name" validate:"required"`
	Debug bool   `map_struct:"debug" env:"APP_DEBUG"`
}
