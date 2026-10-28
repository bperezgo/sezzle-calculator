package config

import (
	"fmt"
	"os"
	"strconv"
)

const defaultPort = "8080"

type Config struct {
	Port    string
	GinMode string
}

func Load() (Config, error) {
	cfg := Config{
		Port:    getenv("PORT", defaultPort),
		GinMode: getenv("GIN_MODE", "release"),
	}

	if n, err := strconv.Atoi(cfg.Port); err != nil || n < 1 || n > 65535 {
		return Config{}, fmt.Errorf("config: PORT must be a number between 1 and 65535, got %q", cfg.Port)
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
