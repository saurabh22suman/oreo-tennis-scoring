package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server configuration
	Port        int
	FrontendURL string
	CORSOrigin  string

	// Database
	DatabaseURL string

	// Admin credentials (NEVER stored in DB)
	AdminUsername     string
	AdminPasswordHash string

	// JWT
	JWTSecret []byte
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{}

	// Server
	port := getEnvOrDefault("PORT", "8080")
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}
	cfg.Port = portNum

	cfg.FrontendURL = getEnvOrDefault("FRONTEND_URL", "http://localhost:5173")
	cfg.CORSOrigin = getEnvOrDefault("CORS_ORIGIN", cfg.FrontendURL)

	// Database
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// Admin credentials
	cfg.AdminUsername = os.Getenv("ADMIN_USERNAME")
	if cfg.AdminUsername == "" {
		return nil, fmt.Errorf("ADMIN_USERNAME environment variable is required")
	}

	cfg.AdminPasswordHash = os.Getenv("ADMIN_PASSWORD_HASH")
	if cfg.AdminPasswordHash == "" {
		return nil, fmt.Errorf("ADMIN_PASSWORD_HASH environment variable is required")
	}

	// JWT Secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	cfg.JWTSecret = []byte(jwtSecret)

	return cfg, nil
}

// GetAllowedOrigins returns a list of allowed CORS origins.
func (c *Config) GetAllowedOrigins() []string {
	origins := []string{c.FrontendURL}
	if c.CORSOrigin != "" && c.CORSOrigin != c.FrontendURL {
		origins = append(origins, strings.Split(c.CORSOrigin, ",")...)
	}
	return origins
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
