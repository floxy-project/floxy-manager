package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port     int
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

func Load() *Config {
	cfg := &Config{
		Port: getEnvInt("PORT", 3001),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5435),
			Name:     getEnv("DB_NAME", "floxy"),
			User:     getEnv("DB_USER", "floxy"),
			Password: getEnv("DB_PASSWORD", "password"),
		},
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
