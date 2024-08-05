package configs

import (
	"fmt"
	"github.com/spf13/viper"
	"reflect"
)

type ConfigLoader interface {
	BindEnvVars(prefix string) error
	LoadConfig() error
	ValidateConfig() error
}

//type ServerConfig struct {
//	Host string `mapstructure:"host" env:"APP_HOST" validate:"required"`
//	Port int    `mapstructure:"port" env:"APP_PORT" validate:"required"`
//}

func BindEnvVars(cfg interface{}, prefix string) error {
	v := reflect.ValueOf(cfg)
	t := reflect.TypeOf(cfg)

	if v.Kind() != reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		typeField := t.Field(i)
		mapStructTag := typeField.Tag.Get("map_struct")
		envTag := typeField.Tag.Get("env")
		requiredTag := typeField.Tag.Get("required")

		if requiredTag == "required" && isZero(field) {
			fieldName := typeField.Name
			if prefix != "" {
				fieldName = prefix + "." + fieldName
			}
			return fmt.Errorf("%s is required", fieldName)
		}

		keyPath := mapStructTag
		if prefix != "" {
			keyPath = fmt.Sprintf("%s.%s", prefix, mapStructTag)
		}

		if envTag != "" {
			if err := viper.BindEnv(keyPath, envTag); err != nil {
				return err
			}
		}

		if field.Kind() == reflect.Struct {
			if err := BindEnvVars(field.Addr().Interface(), keyPath); err != nil {
				return err
			}
		}

	}
	return nil
}

func (cfg *AppConfig) LoadConfig() error {
	viper.AutomaticEnv()
	if err := BindEnvVars(cfg, ""); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func isZero(v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}
