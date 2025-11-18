package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env            string
	Port           int
	DBPath         string
	JWTSecret      string
	CORSOrigins    []string
	LogLevel       string
	RateLimitRPS   float64
	RateLimitBurst int
}

func Load() Config {
	return Config{
		Env:            getEnv("APP_ENV", "dev"),
		Port:           getEnvInt("APP_PORT", 8080),
		DBPath:         getEnv("DB_PATH", "app.db"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-me"),
		CORSOrigins:    splitAndTrim(getEnv("CORS_ORIGINS", "*")),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		RateLimitRPS:   getEnvFloat("RATE_LIMIT_RPS", 10),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 20),
	}
}

func (c Config) Address() string {
	return fmt.Sprintf(":%d", c.Port)
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvFloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return def
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return []string{}
	}
	return out
}


