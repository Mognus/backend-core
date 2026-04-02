package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	CORS     CORSConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Host string
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	AllowOrigins string
	AllowMethods string
	AllowHeaders string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file (optional in production)
	godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: GetEnv("PORT", "8080"),
			Host: GetEnv("HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     GetEnv("DB_HOST", "localhost"),
			User:     GetEnv("DB_USER", "postgres"),
			Password: GetEnv("DB_PASSWORD", "postgres"),
			Name:     GetEnv("DB_NAME", "app_db"),
			Port:     GetEnv("DB_PORT", "5432"),
		},
		CORS: CORSConfig{
			AllowOrigins: GetEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000"),
			AllowMethods: GetEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowHeaders: GetEnv("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization"),
		},
	}

	return cfg, nil
}

// GetEnv gets environment variable with fallback (exported for use in other packages)
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Get is a convenience method to get config value with fallback
func (c *Config) Get(key, fallback string) string {
	return GetEnv(key, fallback)
}
