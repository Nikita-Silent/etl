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
	loader := newEnvLoader()

	dbPort, err := loader.getEnvAsIntStrict("DB_PORT", 5432)
	if err != nil {
		return nil, err
	}
	dbConnectTimeoutSeconds, err := loader.getEnvAsIntStrict("DB_CONNECT_TIMEOUT_SECONDS", int(models.DefaultDBConnectTimeout/time.Second))
	if err != nil {
		return nil, err
	}
	ftpPort, err := loader.getEnvAsIntStrict("FTP_PORT", 21)
	if err != nil {
		return nil, err
	}
	ftpConnectTimeoutSeconds, err := loader.getEnvAsIntStrict("FTP_CONNECT_TIMEOUT_SECONDS", int(models.DefaultFTPConnectTimeout/time.Second))
	if err != nil {
		return nil, err
	}
	ftpPoolSize, err := loader.getEnvAsIntStrict("FTP_POOL_SIZE", 5)
	if err != nil {
		return nil, err
	}
	batchSize, err := loader.getEnvAsIntStrict("BATCH_SIZE", 1000)
	if err != nil {
		return nil, err
	}
	maxRetries, err := loader.getEnvAsIntStrict("MAX_RETRIES", 3)
	if err != nil {
		return nil, err
	}
	retryDelaySeconds, err := loader.getEnvAsIntStrict("RETRY_DELAY_SECONDS", 5)
	if err != nil {
		return nil, err
	}
	waitDelayMinutes, err := loader.getEnvAsIntStrict("WAIT_DELAY_MINUTES", 1)
	if err != nil {
		return nil, err
	}
	pipelineLoadTimeoutMinutes, err := loader.getEnvAsIntStrict("PIPELINE_LOAD_TIMEOUT_MINUTES", int(models.DefaultPipelineLoadTimeout/time.Minute))
	if err != nil {
		return nil, err
	}
	cliRunTimeoutMinutes, err := loader.getEnvAsIntStrict("CLI_RUN_TIMEOUT_MINUTES", int(models.DefaultCLIRunTimeout/time.Minute))
	if err != nil {
		return nil, err
	}
	workerPoolSize, err := loader.getEnvAsIntStrict("WORKER_POOL_SIZE", 10)
	if err != nil {
		return nil, err
	}
	serverPort, err := loader.getEnvAsIntStrict("SERVER_PORT", 8080)
	if err != nil {
		return nil, err
	}
	webhookTimeoutMinutes, err := loader.getEnvAsIntStrict("WEBHOOK_TIMEOUT_MINUTES", 0)
	if err != nil {
		return nil, err
	}
	webhookReportHTTPTimeoutSeconds, err := loader.getEnvAsIntStrict("WEBHOOK_REPORT_HTTP_TIMEOUT_SECONDS", int(models.DefaultWebhookReportHTTPTimeout/time.Second))
	if err != nil {
		return nil, err
	}
	webhookReportResultWaitSeconds, err := loader.getEnvAsIntStrict("WEBHOOK_REPORT_RESULT_WAIT_SECONDS", int(models.DefaultWebhookReportResultWaitTimeout/time.Second))
	if err != nil {
		return nil, err
	}
	httpReadHeaderTimeoutSeconds, err := loader.getEnvAsIntStrict("HTTP_READ_HEADER_TIMEOUT_SECONDS", int(models.DefaultHTTPReadHeaderTimeout/time.Second))
	if err != nil {
		return nil, err
	}
	httpReadTimeoutSeconds, err := loader.getEnvAsIntStrict("HTTP_READ_TIMEOUT_SECONDS", int(models.DefaultHTTPReadTimeout/time.Second))
	if err != nil {
		return nil, err
	}
	httpWriteTimeoutSeconds, err := loader.getEnvAsIntStrict("HTTP_WRITE_TIMEOUT_SECONDS", int(models.DefaultHTTPWriteTimeout/time.Second))
	if err != nil {
		return nil, err
	}
	httpIdleTimeoutSeconds, err := loader.getEnvAsIntStrict("HTTP_IDLE_TIMEOUT_SECONDS", int(models.DefaultHTTPIdleTimeout/time.Second))
	if err != nil {
		return nil, err
	}
	shutdownTimeoutSeconds, err := loader.getEnvAsIntStrict("SHUTDOWN_TIMEOUT_SECONDS", 30)
	if err != nil {
		return nil, err
	}
	kassaStructure, err := parseKassaStructure(loader.getEnv("KASSA_STRUCTURE", ""))
	if err != nil {
		return nil, err
	}

	config := &models.Config{
		// Database settings
		DBHost:           loader.getEnv("DB_HOST", "localhost"),
		DBPort:           dbPort,
		DBUser:           loader.getEnv("DB_USER", "postgres"),
		DBPassword:       loader.getEnv("DB_PASSWORD", ""),
		DBName:           loader.getEnv("DB_NAME", "kassa_db"),
		DBSSLMode:        loader.getEnv("DB_SSLMODE", "disable"),
		DBConnectTimeout: time.Duration(dbConnectTimeoutSeconds) * time.Second,

		// FTP settings
		FTPHost:           loader.getEnv("FTP_HOST", "localhost"),
		FTPPort:           ftpPort,
		FTPUser:           loader.getEnv("FTP_USER", ""),
		FTPPassword:       loader.getEnv("FTP_PASSWORD", ""),
		FTPRequestDir:     loader.getEnv("FTP_REQUEST_DIR", "/request"),
		FTPResponseDir:    loader.getEnv("FTP_RESPONSE_DIR", "/response"),
		FTPPoolSize:       ftpPoolSize,
		FTPConnectTimeout: time.Duration(ftpConnectTimeoutSeconds) * time.Second,
		KassaStructure:    kassaStructure,

		// Application settings
		LocalDir:            loader.getEnv("LOCAL_DIR", "/tmp/frontol"),
		BatchSize:           batchSize,
		MaxRetries:          maxRetries,
		RetryDelay:          time.Duration(retryDelaySeconds) * time.Second,
		WaitDelayMinutes:    time.Duration(waitDelayMinutes) * time.Minute,
		PipelineLoadTimeout: time.Duration(pipelineLoadTimeoutMinutes) * time.Minute,
		CLIRunTimeout:       time.Duration(cliRunTimeoutMinutes) * time.Minute,
		WorkerPoolSize:      workerPoolSize,
		LogLevel:            loader.getEnv("LOG_LEVEL", "info"),
		LogFormat:           loader.getEnv("LOG_FORMAT", "json"),
		LogBackend:          loader.getEnv("LOG_BACKEND", "zerolog"),

		// Webhook server settings
		ServerPort:                     serverPort,
		WebhookReportURL:               loader.getEnv("WEBHOOK_REPORT_URL", ""),
		WebhookTimeoutMinutes:          webhookTimeoutMinutes, // 0 = no timeout, send only on completion
		WebhookReportHTTPTimeout:       time.Duration(webhookReportHTTPTimeoutSeconds) * time.Second,
		WebhookReportResultWaitTimeout: time.Duration(webhookReportResultWaitSeconds) * time.Second,
		WebhookBearerToken:             loader.getEnv("WEBHOOK_BEARER_TOKEN", ""),
		HTTPReadHeaderTimeout:          time.Duration(httpReadHeaderTimeoutSeconds) * time.Second,
		HTTPReadTimeout:                time.Duration(httpReadTimeoutSeconds) * time.Second,
		HTTPWriteTimeout:               time.Duration(httpWriteTimeoutSeconds) * time.Second,
		HTTPIdleTimeout:                time.Duration(httpIdleTimeoutSeconds) * time.Second,
		ShutdownTimeout:                time.Duration(shutdownTimeoutSeconds) * time.Second,
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
	if cfg.DBConnectTimeout <= 0 {
		return fmt.Errorf("DB_CONNECT_TIMEOUT_SECONDS must be greater than 0, got %v", cfg.DBConnectTimeout)
	}
	if cfg.FTPConnectTimeout <= 0 {
		return fmt.Errorf("FTP_CONNECT_TIMEOUT_SECONDS must be greater than 0, got %v", cfg.FTPConnectTimeout)
	}
	if cfg.PipelineLoadTimeout <= 0 {
		return fmt.Errorf("PIPELINE_LOAD_TIMEOUT_MINUTES must be greater than 0, got %v", cfg.PipelineLoadTimeout)
	}
	if cfg.CLIRunTimeout <= 0 {
		return fmt.Errorf("CLI_RUN_TIMEOUT_MINUTES must be greater than 0, got %v", cfg.CLIRunTimeout)
	}
	if cfg.WebhookReportHTTPTimeout <= 0 {
		return fmt.Errorf("WEBHOOK_REPORT_HTTP_TIMEOUT_SECONDS must be greater than 0, got %v", cfg.WebhookReportHTTPTimeout)
	}
	if cfg.WebhookReportResultWaitTimeout <= 0 {
		return fmt.Errorf("WEBHOOK_REPORT_RESULT_WAIT_SECONDS must be greater than 0, got %v", cfg.WebhookReportResultWaitTimeout)
	}
	if cfg.HTTPReadHeaderTimeout <= 0 {
		return fmt.Errorf("HTTP_READ_HEADER_TIMEOUT_SECONDS must be greater than 0, got %v", cfg.HTTPReadHeaderTimeout)
	}
	if cfg.HTTPReadTimeout <= 0 {
		return fmt.Errorf("HTTP_READ_TIMEOUT_SECONDS must be greater than 0, got %v", cfg.HTTPReadTimeout)
	}
	if cfg.HTTPWriteTimeout <= 0 {
		return fmt.Errorf("HTTP_WRITE_TIMEOUT_SECONDS must be greater than 0, got %v", cfg.HTTPWriteTimeout)
	}
	if cfg.HTTPIdleTimeout <= 0 {
		return fmt.Errorf("HTTP_IDLE_TIMEOUT_SECONDS must be greater than 0, got %v", cfg.HTTPIdleTimeout)
	}
	if cfg.ShutdownTimeout <= 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT_SECONDS must be greater than 0, got %v", cfg.ShutdownTimeout)
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

	validLogFormats := map[string]bool{
		"json":    true,
		"text":    true,
		"console": true,
	}
	if !validLogFormats[strings.ToLower(cfg.LogFormat)] {
		return fmt.Errorf("LOG_FORMAT must be one of: json, text, console; got %s", cfg.LogFormat)
	}

	backend := strings.ToLower(cfg.LogBackend)
	if backend == "" {
		backend = "zerolog"
		cfg.LogBackend = backend
	}

	validBackends := map[string]bool{
		"zerolog": true,
		"slog":    true,
	}
	if !validBackends[backend] {
		return fmt.Errorf("LOG_BACKEND must be one of: zerolog, slog; got %s", cfg.LogBackend)
	}

	return nil
}

// LoadDBConfig loads only database configuration (for migrations)
func LoadDBConfig() (*models.Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load() // .env file is optional, continue with environment variables
	loader := newEnvLoader()
	dbPort, err := loader.getEnvAsIntStrict("DB_PORT", 5432)
	if err != nil {
		return nil, err
	}
	dbConnectTimeoutSeconds, err := loader.getEnvAsIntStrict("DB_CONNECT_TIMEOUT_SECONDS", int(models.DefaultDBConnectTimeout/time.Second))
	if err != nil {
		return nil, err
	}

	config := &models.Config{
		// Database settings only
		DBHost:           loader.getEnv("DB_HOST", "localhost"),
		DBPort:           dbPort,
		DBUser:           loader.getEnv("DB_USER", "postgres"),
		DBPassword:       loader.getEnv("DB_PASSWORD", ""),
		DBName:           loader.getEnv("DB_NAME", "kassa_db"),
		DBSSLMode:        loader.getEnv("DB_SSLMODE", "disable"),
		DBConnectTimeout: time.Duration(dbConnectTimeoutSeconds) * time.Second,
	}

	// Validate required DB fields only
	if config.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	if config.DBPort < 1 || config.DBPort > 65535 {
		return nil, fmt.Errorf("DB_PORT must be between 1 and 65535, got %d", config.DBPort)
	}
	if config.DBConnectTimeout <= 0 {
		return nil, fmt.Errorf("DB_CONNECT_TIMEOUT_SECONDS must be greater than 0, got %v", config.DBConnectTimeout)
	}

	return config, nil
}

// parseKassaStructure parses kassa structure from environment variable
func parseKassaStructure(kassaStr string) (map[string][]string, error) {
	if kassaStr == "" {
		return nil, fmt.Errorf("KASSA_STRUCTURE is required")
	}

	// Parse format: "001:folder1,folder2;002:folder1,folder2"
	structure := make(map[string][]string)

	// Split by semicolon to get kassa groups
	kassaGroups := strings.Split(kassaStr, ";")
	for _, group := range kassaGroups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		parts := strings.Split(group, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid KASSA_STRUCTURE group %q", group)
		}
		kassaCode := strings.TrimSpace(parts[0])
		if kassaCode == "" {
			return nil, fmt.Errorf("empty kassa code in KASSA_STRUCTURE")
		}
		folders := strings.Split(parts[1], ",")
		cleanFolders := make([]string, 0, len(folders))
		for _, folder := range folders {
			folder = strings.TrimSpace(folder)
			if folder == "" {
				return nil, fmt.Errorf("empty folder for kassa %s in KASSA_STRUCTURE", kassaCode)
			}
			cleanFolders = append(cleanFolders, folder)
		}
		structure[kassaCode] = cleanFolders
	}

	if len(structure) == 0 {
		return nil, fmt.Errorf("KASSA_STRUCTURE cannot be empty")
	}

	return structure, nil
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

type envLoader struct{}

func newEnvLoader() envLoader {
	return envLoader{}
}

func (envLoader) getEnv(key, defaultValue string) string {
	return getEnv(key, defaultValue)
}

func (envLoader) getEnvAsIntStrict(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer, got %q", key, value)
	}
	return parsed, nil
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
	loader := newEnvLoader()

	ftpUser := loader.getEnv("FTP_USER", "frontol")

	// Auto-generate FTP_ROOT_PATH from FTP_USER if not explicitly set
	ftpRootPath := loader.getEnv("FTP_ROOT_PATH", "")
	if ftpRootPath == "" {
		ftpRootPath = fmt.Sprintf("/home/ftp/%s", ftpUser)
	}
	ftpPort, err := loader.getEnvAsIntStrict("FTP_PORT", 21)
	if err != nil {
		return nil, err
	}
	passiveMinPort, err := loader.getEnvAsIntStrict("PASV_MIN_PORT", 30000)
	if err != nil {
		return nil, err
	}
	passiveMaxPort, err := loader.getEnvAsIntStrict("PASV_MAX_PORT", 30009)
	if err != nil {
		return nil, err
	}

	config := &FTPConfig{
		FTPPort:        ftpPort,
		FTPUser:        ftpUser,
		FTPPassword:    loader.getEnv("FTP_PASSWORD", "frontol123"),
		FTPRootPath:    ftpRootPath,
		PublicHost:     loader.getEnv("PUBLICHOST", ""),
		PassiveMinPort: passiveMinPort,
		PassiveMaxPort: passiveMaxPort,
		LogLevel:       loader.getEnv("LOG_LEVEL", "info"),
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
