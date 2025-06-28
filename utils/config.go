package utils

import (
	"os"
	"reflect"
)

type Config struct {
	AppName    string
	DbHost     string
	DbUsername string
	DbPassword string
}

func CfgLoad(app string) *Config {
	GlobalLogger().Info("Loading config for %s", app)
	defer GlobalLogger().Info("Loading config for %s done", app)
	return &Config{
		AppName:    app,
		DbHost:     getEnv("DB_HOST", ""),
		DbUsername: getEnv("DB_USERNAME", ""),
		DbPassword: getEnv("DB_PASSWORD", ""),
	}
}

func (cfg Config) DumpAll() {
	log := GlobalLogger()
	val := reflect.ValueOf(cfg)
	typ := val.Type()
	log.Debug("----DUMP CFG BGN-----")
	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldValue := val.Field(i)
		log.Debug("%s = %v", fieldType.Name, fieldValue.Interface())
	}
	log.Debug("----DUMP CFG END-----")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		GlobalLogger().Debug("%s set value: %s", key, value)
		return value
	}
	GlobalLogger().Debug("%s set default: %s", key, defaultValue)
	return defaultValue
}
