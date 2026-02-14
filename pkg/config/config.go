package config

import (
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	MongoURI string
	APIKeys  []string
	Port     string
	DBName   string
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		MongoURI: getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		Port:     getEnv("PORT", "8080"),
		DBName:   getEnv("DB_NAME", "services_db"),
	}

	// Parse comma-separated API keys
	apiKeysStr := getEnv("API_KEYS", "")
	if apiKeysStr != "" {
		cfg.APIKeys = strings.Split(apiKeysStr, ",")
		// Trim whitespace from each key
		for i, key := range cfg.APIKeys {
			cfg.APIKeys[i] = strings.TrimSpace(key)
		}
	}

	return cfg
}

// getEnv returns the environment variable value or a default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// HasAPIKeys returns true if API keys are configured
func (c *Config) HasAPIKeys() bool {
	return len(c.APIKeys) > 0
}

// IsValidAPIKey checks if the provided key is in the configured keys
func (c *Config) IsValidAPIKey(key string) bool {
	for _, k := range c.APIKeys {
		if k == key {
			return true
		}
	}
	return false
}
