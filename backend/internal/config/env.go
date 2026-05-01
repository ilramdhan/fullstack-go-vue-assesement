package config

import (
	"os"
	"strings"
)

var (
	JwtSecret          = []byte(getEnv("JWT_SECRET", "dev-secret-replace-me"))
	JwtExpired         = getEnv("JWT_EXPIRED", "24h")
	HttpAddress        = getEnv("HTTP_ADDR", ":8080")
	DatabasePath       = getEnv("DATABASE_PATH", "data/dashboard.db")
	CorsAllowedOrigins = splitCSV(getEnv("CORS_ALLOWED_ORIGINS",
		"http://localhost:5173,http://localhost:4173,http://localhost:3000"))
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
