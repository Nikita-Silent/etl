package config

import (
	"os"
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
			name:     "empty string returns default",
			input:    "",
			wantLen:  2, // default has 2 kassas
			wantKeys: []string{"001", "002"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseKassaStructure(tt.input)

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
	defer func() {
		os.Setenv("DB_PASSWORD", origDBPass)
		os.Setenv("FTP_USER", origFTPUser)
		os.Setenv("FTP_PASSWORD", origFTPPass)
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

			_, err := LoadConfig()

			if tt.wantErr {
				if err == nil {
					t.Error("LoadConfig() expected error, got nil")
					return
				}
				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			envBackup := make(map[string]string)
			envKeys := []string{
				"DB_PASSWORD", "FTP_USER", "FTP_PASSWORD", "DB_PORT", "BATCH_SIZE", "MAX_RETRIES", "LOG_LEVEL",
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
				"DB_PASSWORD":  "pass",
				"FTP_USER":     "user",
				"FTP_PASSWORD": "pass",
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
				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
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
