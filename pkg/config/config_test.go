package config

import (
	"os"
	"strings"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		want         string
	}{
		{
			name:         "env set returns env value",
			key:          "TEST_VAR_1",
			defaultValue: "default",
			envValue:     "custom",
			setEnv:       true,
			want:         "custom",
		},
		{
			name:         "env not set returns default",
			key:          "TEST_VAR_2",
			defaultValue: "default",
			envValue:     "",
			setEnv:       false,
			want:         "default",
		},
		{
			name:         "empty env value returns default",
			key:          "TEST_VAR_3",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnv(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		setEnv       bool
		want         int
	}{
		{
			name:         "valid int",
			key:          "TEST_INT_1",
			defaultValue: 10,
			envValue:     "42",
			setEnv:       true,
			want:         42,
		},
		{
			name:         "invalid int returns default",
			key:          "TEST_INT_2",
			defaultValue: 10,
			envValue:     "not_a_number",
			setEnv:       true,
			want:         10,
		},
		{
			name:         "empty returns default",
			key:          "TEST_INT_3",
			defaultValue: 10,
			envValue:     "",
			setEnv:       false,
			want:         10,
		},
		{
			name:         "negative int",
			key:          "TEST_INT_4",
			defaultValue: 10,
			envValue:     "-5",
			setEnv:       true,
			want:         -5,
		},
		{
			name:         "zero",
			key:          "TEST_INT_5",
			defaultValue: 10,
			envValue:     "0",
			setEnv:       true,
			want:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsInt(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvAsInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseKassaStructure(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		errSub   string
		wantLen  int
		wantKeys []string
	}{
		{
			name:     "valid single kassa",
			input:    "001:folder1,folder2",
			wantLen:  1,
			wantKeys: []string{"001"},
		},
		{
			name:     "valid multiple kassas",
			input:    "001:folder1,folder2;002:folder3,folder4",
			wantLen:  2,
			wantKeys: []string{"001", "002"},
		},
		{
			name:    "empty string returns error",
			input:   "",
			wantErr: true,
			errSub:  "KASSA_STRUCTURE is required",
		},
		{
			name:     "single folder per kassa",
			input:    "P13:P13;N22:N22_Inter",
			wantLen:  2,
			wantKeys: []string{"P13", "N22"},
		},
		{
			name:     "with spaces",
			input:    "001: folder1, folder2 ; 002: folder3",
			wantLen:  2,
			wantKeys: []string{"001", "002"},
		},
		{
			name:    "invalid group format",
			input:   "001-folder1",
			wantErr: true,
			errSub:  "invalid KASSA_STRUCTURE group",
		},
		{
			name:    "empty folder rejected",
			input:   "001:folder1,",
			wantErr: true,
			errSub:  "empty folder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKassaStructure(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("parseKassaStructure() expected error, got nil")
				}
				if tt.errSub != "" && !strings.Contains(err.Error(), tt.errSub) {
					t.Fatalf("parseKassaStructure() error = %v, want substring %q", err, tt.errSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseKassaStructure() unexpected error: %v", err)
			}

			if len(got) != tt.wantLen {
				t.Errorf("parseKassaStructure() len = %v, want %v", len(got), tt.wantLen)
			}

			for _, key := range tt.wantKeys {
				if _, ok := got[key]; !ok {
					t.Errorf("parseKassaStructure() missing key %v", key)
				}
			}
		})
	}
}

func TestLoadConfig_Validation(t *testing.T) {
	// Save original env vars
	origDBPass := os.Getenv("DB_PASSWORD")
	origFTPUser := os.Getenv("FTP_USER")
	origFTPPass := os.Getenv("FTP_PASSWORD")
	origKassaStructure := os.Getenv("KASSA_STRUCTURE")
	defer func() {
		os.Setenv("DB_PASSWORD", origDBPass)
		os.Setenv("FTP_USER", origFTPUser)
		os.Setenv("FTP_PASSWORD", origFTPPass)
		if origKassaStructure == "" {
			os.Unsetenv("KASSA_STRUCTURE")
		} else {
			os.Setenv("KASSA_STRUCTURE", origKassaStructure)
		}
	}()

	tests := []struct {
		name      string
		dbPass    string
		ftpUser   string
		ftpPass   string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "missing DB_PASSWORD",
			dbPass:    "",
			ftpUser:   "user",
			ftpPass:   "pass",
			wantErr:   true,
			errSubstr: "DB_PASSWORD",
		},
		{
			name:      "missing FTP_USER",
			dbPass:    "pass",
			ftpUser:   "",
			ftpPass:   "pass",
			wantErr:   true,
			errSubstr: "FTP_USER",
		},
		{
			name:      "missing FTP_PASSWORD",
			dbPass:    "pass",
			ftpUser:   "user",
			ftpPass:   "",
			wantErr:   true,
			errSubstr: "FTP_PASSWORD",
		},
		{
			name:    "all required present",
			dbPass:  "pass",
			ftpUser: "user",
			ftpPass: "pass",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("DB_PASSWORD", tt.dbPass)
			os.Setenv("FTP_USER", tt.ftpUser)
			os.Setenv("FTP_PASSWORD", tt.ftpPass)
			os.Setenv("KASSA_STRUCTURE", "P13:P13")

			_, err := LoadConfig()

			if tt.wantErr {
				if err == nil {
					t.Error("LoadConfig() expected error, got nil")
					return
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("LoadConfig() error = %v, want to contain %v", err, tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("LoadConfig() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		modifyFn  func(*testing.T) map[string]string
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid config",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":  "pass",
					"FTP_USER":     "user",
					"FTP_PASSWORD": "pass",
				}
			},
			wantErr: false,
		},
		{
			name: "invalid DB_PORT (too low)",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":  "pass",
					"FTP_USER":     "user",
					"FTP_PASSWORD": "pass",
					"DB_PORT":      "0",
				}
			},
			wantErr:   true,
			errSubstr: "DB_PORT must be between 1 and 65535",
		},
		{
			name: "invalid DB_PORT format",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":     "pass",
					"FTP_USER":        "user",
					"FTP_PASSWORD":    "pass",
					"KASSA_STRUCTURE": "P13:P13",
					"DB_PORT":         "invalid",
				}
			},
			wantErr:   true,
			errSubstr: "DB_PORT must be a valid integer",
		},
		{
			name: "invalid DB_PORT (too high)",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":  "pass",
					"FTP_USER":     "user",
					"FTP_PASSWORD": "pass",
					"DB_PORT":      "70000",
				}
			},
			wantErr:   true,
			errSubstr: "DB_PORT must be between 1 and 65535",
		},
		{
			name: "invalid BATCH_SIZE (zero)",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":  "pass",
					"FTP_USER":     "user",
					"FTP_PASSWORD": "pass",
					"BATCH_SIZE":   "0",
				}
			},
			wantErr:   true,
			errSubstr: "BATCH_SIZE must be greater than 0",
		},
		{
			name: "invalid BATCH_SIZE (too large)",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":  "pass",
					"FTP_USER":     "user",
					"FTP_PASSWORD": "pass",
					"BATCH_SIZE":   "200000",
				}
			},
			wantErr:   true,
			errSubstr: "BATCH_SIZE too large",
		},
		{
			name: "invalid MAX_RETRIES (negative)",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":  "pass",
					"FTP_USER":     "user",
					"FTP_PASSWORD": "pass",
					"MAX_RETRIES":  "-1",
				}
			},
			wantErr:   true,
			errSubstr: "MAX_RETRIES must be non-negative",
		},
		{
			name: "invalid MAX_RETRIES (too large)",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":  "pass",
					"FTP_USER":     "user",
					"FTP_PASSWORD": "pass",
					"MAX_RETRIES":  "20",
				}
			},
			wantErr:   true,
			errSubstr: "MAX_RETRIES too large",
		},
		{
			name: "invalid LOG_LEVEL",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":  "pass",
					"FTP_USER":     "user",
					"FTP_PASSWORD": "pass",
					"LOG_LEVEL":    "invalid",
				}
			},
			wantErr:   true,
			errSubstr: "LOG_LEVEL must be one of",
		},
		{
			name: "invalid LOG_FORMAT",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":     "pass",
					"FTP_USER":        "user",
					"FTP_PASSWORD":    "pass",
					"KASSA_STRUCTURE": "P13:P13",
					"LOG_FORMAT":      "yaml",
				}
			},
			wantErr:   true,
			errSubstr: "LOG_FORMAT must be one of",
		},
		{
			name: "invalid DB connect timeout",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":                "pass",
					"FTP_USER":                   "user",
					"FTP_PASSWORD":               "pass",
					"KASSA_STRUCTURE":            "P13:P13",
					"DB_CONNECT_TIMEOUT_SECONDS": "0",
				}
			},
			wantErr:   true,
			errSubstr: "DB_CONNECT_TIMEOUT_SECONDS must be greater than 0",
		},
		{
			name: "invalid pipeline load timeout",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":                   "pass",
					"FTP_USER":                      "user",
					"FTP_PASSWORD":                  "pass",
					"KASSA_STRUCTURE":               "P13:P13",
					"PIPELINE_LOAD_TIMEOUT_MINUTES": "0",
				}
			},
			wantErr:   true,
			errSubstr: "PIPELINE_LOAD_TIMEOUT_MINUTES must be greater than 0",
		},
		{
			name: "missing KASSA_STRUCTURE",
			modifyFn: func(t *testing.T) map[string]string {
				return map[string]string{
					"DB_PASSWORD":     "pass",
					"FTP_USER":        "user",
					"FTP_PASSWORD":    "pass",
					"KASSA_STRUCTURE": "",
				}
			},
			wantErr:   true,
			errSubstr: "KASSA_STRUCTURE is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			envBackup := make(map[string]string)
			envKeys := []string{
				"DB_PASSWORD", "FTP_USER", "FTP_PASSWORD", "DB_PORT", "BATCH_SIZE", "MAX_RETRIES", "LOG_LEVEL",
				"LOG_FORMAT", "KASSA_STRUCTURE", "DB_CONNECT_TIMEOUT_SECONDS", "FTP_CONNECT_TIMEOUT_SECONDS",
				"PIPELINE_LOAD_TIMEOUT_MINUTES", "CLI_RUN_TIMEOUT_MINUTES", "WEBHOOK_REPORT_HTTP_TIMEOUT_SECONDS",
				"WEBHOOK_REPORT_RESULT_WAIT_SECONDS", "HTTP_READ_HEADER_TIMEOUT_SECONDS", "HTTP_READ_TIMEOUT_SECONDS",
				"HTTP_WRITE_TIMEOUT_SECONDS", "HTTP_IDLE_TIMEOUT_SECONDS", "SHUTDOWN_TIMEOUT_SECONDS",
			}
			for _, key := range envKeys {
				envBackup[key] = os.Getenv(key)
				os.Unsetenv(key)
			}
			defer func() {
				for _, key := range envKeys {
					if val, ok := envBackup[key]; ok && val != "" {
						os.Setenv(key, val)
					} else {
						os.Unsetenv(key)
					}
				}
			}()

			// Base required envs
			baseEnv := map[string]string{
				"DB_PASSWORD":     "pass",
				"FTP_USER":        "user",
				"FTP_PASSWORD":    "pass",
				"KASSA_STRUCTURE": "P13:P13",
			}
			for k, v := range baseEnv {
				os.Setenv(k, v)
			}

			// Set test env vars
			envVars := tt.modifyFn(t)
			for key, val := range envVars {
				os.Setenv(key, val)
			}

			// Load and validate config
			_, err := LoadConfig()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("Error = %v, want to contain %v", err, tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLoadFTPConfig_InvalidPorts(t *testing.T) {
	keys := []string{"FTP_PORT", "PASV_MIN_PORT", "PASV_MAX_PORT"}
	backup := make(map[string]string, len(keys))
	for _, key := range keys {
		backup[key] = os.Getenv(key)
		defer func(k string) {
			if backup[k] == "" {
				os.Unsetenv(k)
				return
			}
			os.Setenv(k, backup[k])
		}(key)
	}

	os.Setenv("FTP_PORT", "nope")
	if _, err := LoadFTPConfig(); err == nil || !strings.Contains(err.Error(), "FTP_PORT must be a valid integer") {
		t.Fatalf("LoadFTPConfig() error = %v, want FTP_PORT validation error", err)
	}
}
