package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Addr      string
	DBURL     string
	Platform  string
	JWTSecret string
	PolkaKey  string
	StaticDir string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:      getEnv("ADDR", ":8080"),
		DBURL:     os.Getenv("DB_URL"),
		Platform:  os.Getenv("PLATFORM"),
		JWTSecret: os.Getenv("SECRET"),
		PolkaKey:  os.Getenv("POLKA_KEY"),
		StaticDir: getEnv("STATIC_DIR", "./public"),
	}

	required := []struct{ key, value string }{
		{"DB_URL", cfg.DBURL},
		{"SECRET", cfg.JWTSecret},
		{"POLKA_KEY", cfg.PolkaKey},
	}

	var errs []error
	for _, r := range required {
		if r.value == "" {
			errs = append(errs, fmt.Errorf("%s must be set", r.key))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
