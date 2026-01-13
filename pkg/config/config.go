package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/user/go-frontol-loader/pkg/models"
)

// LoadConfig loads configuration from .env file and environment variables
func LoadConfig() (*models.Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load() // .env file is optional, continue with environment variables

	config := &models.Config{
		// Database settings
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "kassa_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// FTP settings
		FTPHost:        getEnv("FTP_HOST", "localhost"),
		FTPPort:        getEnvAsInt("FTP_PORT", 21),
		FTPUser:        getEnv("FTP_USER", ""),
		FTPPassword:    getEnv("FTP_PASSWORD", ""),
		FTPRequestDir:  getEnv("FTP_REQUEST_DIR", "/request"),
		FTPResponseDir: getEnv("FTP_RESPONSE_DIR", "/response"),
		FTPPoolSize:    getEnvAsInt("FTP_POOL_SIZE", 5),
		KassaStructure: parseKassaStructure(getEnv("KASSA_STRUCTURE", "")),

		// Application settings
		LocalDir:         getEnv("LOCAL_DIR", "/tmp/frontol"),
		BatchSize:        getEnvAsInt("BATCH_SIZE", 1000),
		MaxRetries:       getEnvAsInt("MAX_RETRIES", 3),
		RetryDelay:       time.Duration(getEnvAsInt("RETRY_DELAY_SECONDS", 5)) * time.Second,
		WaitDelayMinutes: time.Duration(getEnvAsInt("WAIT_DELAY_MINUTES", 1)) * time.Minute,
		WorkerPoolSize:   getEnvAsInt("WORKER_POOL_SIZE", 10),
		LogLevel:         getEnv("LOG_LEVEL", "info"),

		// Webhook server settings
		ServerPort:            getEnvAsInt("SERVER_PORT", 8080),
		WebhookReportURL:      getEnv("WEBHOOK_REPORT_URL", ""),
		WebhookTimeoutMinutes: getEnvAsInt("WEBHOOK_TIMEOUT_MINUTES", 0), // 0 = no timeout, send only on completion
		WebhookBearerToken:    getEnv("WEBHOOK_BEARER_TOKEN", ""),        // Bearer token for webhook authorization
		ShutdownTimeout:       time.Duration(getEnvAsInt("SHUTDOWN_TIMEOUT_SECONDS", 30)) * time.Second,
	}

	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfig performs comprehensive validation of configuration
func ValidateConfig(cfg *models.Config) error {
	// Validate required fields
	if cfg.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if cfg.FTPUser == "" {
		return fmt.Errorf("FTP_USER is required")
	}
	if cfg.FTPPassword == "" {
		return fmt.Errorf("FTP_PASSWORD is required")
	}

	// Validate port ranges
	if cfg.DBPort < 1 || cfg.DBPort > 65535 {
		return fmt.Errorf("DB_PORT must be between 1 and 65535, got %d", cfg.DBPort)
	}
	if cfg.FTPPort < 1 || cfg.FTPPort > 65535 {
		return fmt.Errorf("FTP_PORT must be between 1 and 65535, got %d", cfg.FTPPort)
	}
	if cfg.ServerPort < 1 || cfg.ServerPort > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535, got %d", cfg.ServerPort)
	}

	// Validate FTP pool size
	if cfg.FTPPoolSize < 1 {
		return fmt.Errorf("FTP_POOL_SIZE must be at least 1, got %d", cfg.FTPPoolSize)
	}
	if cfg.FTPPoolSize > 50 {
		return fmt.Errorf("FTP_POOL_SIZE too large (max 50), got %d", cfg.FTPPoolSize)
	}

	// Validate worker pool size
	if cfg.WorkerPoolSize < 1 {
		return fmt.Errorf("WORKER_POOL_SIZE must be at least 1, got %d", cfg.WorkerPoolSize)
	}
	if cfg.WorkerPoolSize > 100 {
		return fmt.Errorf("WORKER_POOL_SIZE too large (max 100), got %d", cfg.WorkerPoolSize)
	}

	// Validate batch size
	if cfg.BatchSize <= 0 {
		return fmt.Errorf("BATCH_SIZE must be greater than 0, got %d", cfg.BatchSize)
	}
	if cfg.BatchSize > 100000 {
		return fmt.Errorf("BATCH_SIZE too large (max 100000), got %d", cfg.BatchSize)
	}

	// Validate retry settings
	if cfg.MaxRetries < 0 {
		return fmt.Errorf("MAX_RETRIES must be non-negative, got %d", cfg.MaxRetries)
	}
	if cfg.MaxRetries > 10 {
		return fmt.Errorf("MAX_RETRIES too large (max 10), got %d", cfg.MaxRetries)
	}
	if cfg.RetryDelay < 0 {
		return fmt.Errorf("RETRY_DELAY_SECONDS must be non-negative, got %v", cfg.RetryDelay)
	}

	// Validate timeout settings
	if cfg.WaitDelayMinutes < 0 {
		return fmt.Errorf("WAIT_DELAY_MINUTES must be non-negative, got %v", cfg.WaitDelayMinutes)
	}
	if cfg.WebhookTimeoutMinutes < 0 {
		return fmt.Errorf("WEBHOOK_TIMEOUT_MINUTES must be non-negative, got %d", cfg.WebhookTimeoutMinutes)
	}

	// Validate kassa structure is not empty
	if len(cfg.KassaStructure) == 0 {
		return fmt.Errorf("KASSA_STRUCTURE cannot be empty")
	}

	// Validate each kassa has at least one folder
	for kassaCode, folders := range cfg.KassaStructure {
		if len(folders) == 0 {
			return fmt.Errorf("kassa %s has no folders configured", kassaCode)
		}
		if kassaCode == "" {
			return fmt.Errorf("empty kassa code in KASSA_STRUCTURE")
		}
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[strings.ToLower(cfg.LogLevel)] {
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error; got %s", cfg.LogLevel)
	}

	return nil
}

// LoadDBConfig loads only database configuration (for migrations)
func LoadDBConfig() (*models.Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load() // .env file is optional, continue with environment variables

	config := &models.Config{
		// Database settings only
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "kassa_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Validate required DB fields only
	if config.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	return config, nil
}

// parseKassaStructure parses kassa structure from environment variable
func parseKassaStructure(kassaStr string) map[string][]string {
	if kassaStr == "" {
		// Default structure if not provided
		return map[string][]string{
			"001": {"folder1", "folder2"},
			"002": {"folder1", "folder2"},
		}
	}

	// Parse format: "001:folder1,folder2;002:folder1,folder2"
	structure := make(map[string][]string)

	// Split by semicolon to get kassa groups
	kassaGroups := strings.Split(kassaStr, ";")
	for _, group := range kassaGroups {
		parts := strings.Split(group, ":")
		if len(parts) == 2 {
			kassaCode := strings.TrimSpace(parts[0])
			folders := strings.Split(parts[1], ",")
			var cleanFolders []string
			for _, folder := range folders {
				cleanFolders = append(cleanFolders, strings.TrimSpace(folder))
			}
			structure[kassaCode] = cleanFolders
		}
	}

	return structure
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// FTPConfig represents FTP server configuration
type FTPConfig struct {
	FTPPort        int
	FTPUser        string
	FTPPassword    string
	FTPRootPath    string
	PublicHost     string
	PassiveMinPort int
	PassiveMaxPort int
	LogLevel       string
}

// LoadFTPConfig loads FTP server configuration from environment variables
func LoadFTPConfig() (*FTPConfig, error) {
	// Load .env file if it exists
	_ = godotenv.Load() // .env file is optional, continue with environment variables

	ftpUser := getEnv("FTP_USER", "frontol")

	// Auto-generate FTP_ROOT_PATH from FTP_USER if not explicitly set
	ftpRootPath := getEnv("FTP_ROOT_PATH", "")
	if ftpRootPath == "" {
		ftpRootPath = fmt.Sprintf("/home/ftp/%s", ftpUser)
	}

	config := &FTPConfig{
		FTPPort:        getEnvAsInt("FTP_PORT", 21),
		FTPUser:        ftpUser,
		FTPPassword:    getEnv("FTP_PASSWORD", "frontol123"),
		FTPRootPath:    ftpRootPath,
		PublicHost:     getEnv("PUBLICHOST", ""),
		PassiveMinPort: getEnvAsInt("PASV_MIN_PORT", 30000),
		PassiveMaxPort: getEnvAsInt("PASV_MAX_PORT", 30009),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	// Validate required fields
	if config.FTPUser == "" {
		return nil, fmt.Errorf("FTP_USER is required")
	}
	if config.FTPPassword == "" {
		return nil, fmt.Errorf("FTP_PASSWORD is required")
	}

	return config, nil
}
