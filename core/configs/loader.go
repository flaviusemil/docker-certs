package configs

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"reflect"
	"strconv"
)

func BindEnvVarsWithDefaults(cfg interface{}) error {
	return bindEnvVarsWithDefaults(cfg, "")
}

func bindEnvVarsWithDefaults(cfg interface{}, parentKey string) error {
	val := reflect.ValueOf(cfg)
	typ := reflect.TypeOf(cfg)

	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("[configs] cfg must be a pointer to a struct")
	}

	val = val.Elem()
	typ = typ.Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		key := fieldType.Tag.Get("map_struct")
		envTag := fieldType.Tag.Get("env")
		defaultTag := fieldType.Tag.Get("default")

		if key == "" {
			key = fieldType.Name
		}

		if parentKey != "" {
			key = parentKey + "." + key
		}

		if defaultTag != "" {
			log.Printf("[configs] Setting default for %s = %s", key, defaultTag)
			switch field.Kind() {
			case reflect.String:
				viper.SetDefault(key, defaultTag)
			case reflect.Bool:
				v, err := strconv.ParseBool(defaultTag)
				if err == nil {
					viper.SetDefault(key, v)
				}
			case reflect.Int:
				v, err := strconv.Atoi(defaultTag)
				if err == nil {
					viper.SetDefault(key, v)
				}
			default:
				panic("[configs] unhandled default type for field: " + key)
			}
		}

		if envTag != "" {
			if err := viper.BindEnv(key, envTag); err != nil {
				return fmt.Errorf("[configs] failed to bind env %s to key %s: %w", envTag, key, err)
			}
		}

		if field.Kind() == reflect.Struct {
			fieldPtr := field.Addr().Interface()
			if err := bindEnvVarsWithDefaults(fieldPtr, key); err != nil {
				return err
			}
		}
	}

	return nil
}
