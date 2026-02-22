// Package config loads server configuration from environment variables.
// Use a .env file locally (loaded by godotenv); in Docker pass env vars directly.
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all server configuration.
type Config struct {
	// Network
	LoginAddr string // e.g. "0.0.0.0:2000"
	GameAddr  string // e.g. "0.0.0.0:5000"

	// Advertised game server address (what the login server tells clients)
	GamePublicIP   string
	GamePublicPort uint16

	// PostgreSQL
	DBDSN string // postgres://user:pass@host:port/dbname?sslmode=disable

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Session token TTL (seconds); default 300
	SessionTTL int
}

// Load reads configuration from environment, loading .env if present.
func Load() *Config {
	// Ignore error — .env is optional (not present in Docker)
	_ = godotenv.Load()

	return &Config{
		LoginAddr:      getEnv("LOGIN_ADDR", "0.0.0.0:2000"),
		GameAddr:       getEnv("GAME_ADDR", "0.0.0.0:5000"),
		GamePublicIP:   getEnv("GAME_PUBLIC_IP", "127.0.0.1"),
		GamePublicPort: uint16(getEnvInt("GAME_PUBLIC_PORT", 5000)),
		DBDSN:          getEnv("DB_DSN", "postgres://r2:r2password@localhost:5432/r2online?sslmode=disable"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvInt("REDIS_DB", 0),
		SessionTTL:     getEnvInt("SESSION_TTL", 300),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Printf("config: invalid %s=%q, using default %d", key, v, fallback)
			return fallback
		}
		return n
	}
	return fallback
}
