package config

import (
	"os"
	"time"
)

type Config struct {
	ListenAddr           string
	DatabaseURL          string
	GeoIPProviderURL     string
	GeoIPProviderTimeout time.Duration
}

func Load() Config {
	return Config{
		ListenAddr:           getEnv("LISTEN_ADDR", ":8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://geoip:geoip@localhost:5432/geoip?sslmode=disable"),
		GeoIPProviderURL:     getEnv("GEOIP_PROVIDER_URL", "https://ipapi.co"),
		GeoIPProviderTimeout: getDurationEnv("GEOIP_PROVIDER_TIMEOUT", 5*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
