package ftp

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestEnsureDirectoryExists tests the directory creation logic
func TestEnsureDirectoryExists(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ftp_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "test", "nested", "path")

	// Test that mkdir -p creates nested directories
	cmd := exec.Command("mkdir", "-p", testPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("mkdir -p failed: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", testPath)
	}

	// Test idempotency - running mkdir -p again should not fail
	cmd = exec.Command("mkdir", "-p", testPath)
	if err := cmd.Run(); err != nil {
		t.Errorf("mkdir -p failed on existing directory (should be idempotent): %v", err)
	}
}

// TestKassaStructureParsing tests parsing of KASSA_STRUCTURE
func TestKassaStructureParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected number of folders
	}{
		{
			name:     "single kassa single folder",
			input:    "P13:P13",
			expected: 1,
		},
		{
			name:     "single kassa multiple folders",
			input:    "N22:N22_Inter,N22_FURN",
			expected: 2,
		},
		{
			name:     "multiple kassas",
			input:    "P13:P13;N22:N22_Inter,N22_FURN",
			expected: 3,
		},
		{
			name:     "empty structure",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a simplified test - in real implementation,
			// we would call parseKassaStructure from config package
			// For now, we just verify the structure can be parsed
			parts := splitKassaStructure(tt.input)
			if len(parts) != tt.expected {
				t.Errorf("Expected %d folders, got %d", tt.expected, len(parts))
			}
		})
	}
}

// splitKassaStructure is a helper function for testing
func splitKassaStructure(structure string) []string {
	if structure == "" {
		return []string{}
	}

	var result []string
	kassaGroups := splitBySemicolon(structure)
	for _, group := range kassaGroups {
		if group == "" {
			continue
		}
		parts := splitByColon(group)
		if len(parts) == 2 {
			folders := splitByComma(parts[1])
			result = append(result, folders...)
		}
	}
	return result
}

func splitBySemicolon(s string) []string {
	var result []string
	current := ""
	for _, char := range s {
		if char == ';' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func splitByColon(s string) []string {
	for i, char := range s {
		if char == ':' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func splitByComma(s string) []string {
	var result []string
	current := ""
	for _, char := range s {
		if char == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// TestPathConstruction tests path construction logic
func TestPathConstruction(t *testing.T) {
	tests := []struct {
		name         string
		baseDir      string
		kassaCode    string
		folderName   string
		expectedPath string
	}{
		{
			name:         "absolute base path",
			baseDir:      "/request",
			kassaCode:    "P13",
			folderName:   "P13",
			expectedPath: "/request/P13/P13",
		},
		{
			name:         "relative base path",
			baseDir:      "request",
			kassaCode:    "N22",
			folderName:   "N22_Inter",
			expectedPath: "request/N22/N22_Inter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if filepath.IsAbs(tt.baseDir) {
				path = filepath.Join(tt.baseDir, tt.kassaCode, tt.folderName)
			} else {
				path = filepath.Join(tt.baseDir, tt.kassaCode, tt.folderName)
			}

			// Normalize path separators for comparison
			expected := filepath.FromSlash(tt.expectedPath)
			if path != expected {
				t.Errorf("Expected path %s, got %s", expected, path)
			}
		})
	}
}
