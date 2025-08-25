package config

import (
	"github.com/lpernett/godotenv"
	"os"
)

type Config struct {
	Port   string
	DBPath string
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()
	return Config{
		Port:   getEnv("PORT", "8080"),
		DBPath: getEnv("DB_PATH", "./internal/database/database.sql"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
