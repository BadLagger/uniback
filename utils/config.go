package utils

import (
	"os"
	"reflect"
	"strconv"
)

type Config struct {
	AppName         string
	DbHost          string
	DbPort          string
	DbUsername      string
	DbPassword      string
	DbName          string
	DbCtxTimeoutSec int
}

func CfgLoad(app string) *Config {
	GlobalLogger().Info("Loading config for %s", app)
	defer GlobalLogger().Info("Loading config for %s done", app)
	return &Config{
		AppName:         app,
		DbHost:          getEnv("DB_HOST", "localhost"),
		DbPort:          getEnv("DB_PORT", "8080"),
		DbUsername:      getEnv("DB_USERNAME", "postgres"),
		DbPassword:      getEnv("DB_PASSWORD", ""),
		DbName:          getEnv("DB_NAME", ""),
		DbCtxTimeoutSec: getEnvInt("DB_CTX_TOUT_SEC", 3),
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

func getEnvInt(key string, defaultValue int) int {
	if strValue, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(strValue)
		if err == nil {
			GlobalLogger().Debug("%s set value: %d", key, intValue)
			return intValue
		}
		GlobalLogger().Error("failed to parse %s : %w", key, err)
	}
	GlobalLogger().Debug("%s set default: %d", key, defaultValue)
	return defaultValue
}
