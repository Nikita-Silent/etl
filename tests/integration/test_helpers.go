//go:build integration
// +build integration

package integration

import (
	"fmt"
	"os"

	"github.com/user/go-frontol-loader/pkg/models"
)

// getTestDBConfig returns database config for tests
// Uses Docker container if available, otherwise falls back to environment variables
func getTestDBConfig() *models.Config {
	// Default values for Docker test container
	cfg := &models.Config{
		DBHost:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		DBPort:     getEnvOrDefaultInt("TEST_DB_PORT", 5433), // Port 5433 for test container
		DBUser:     getEnvOrDefault("TEST_DB_USER", "frontol_user"),
		DBPassword: getEnvOrDefault("TEST_DB_PASSWORD", "test_password"),
		DBName:     getEnvOrDefault("TEST_DB_NAME", "kassa_db_test"),
		DBSSLMode:  "disable",
	}

	// If TEST_DB_HOST is not set, try to use standard DB_* variables
	if cfg.DBHost == "localhost" && os.Getenv("DB_HOST") != "" {
		cfg.DBHost = os.Getenv("DB_HOST")
		cfg.DBPort = getEnvOrDefaultInt("DB_PORT", 5432)
		cfg.DBUser = getEnvOrDefault("DB_USER", "frontol_user")
		cfg.DBPassword = getEnvOrDefault("DB_PASSWORD", "")
		cfg.DBName = getEnvOrDefault("DB_NAME", "kassa_db")
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
