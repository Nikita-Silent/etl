package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/go-frontol-loader/pkg/parser"
)

func TestE2E_ParseResponseFile(t *testing.T) {
	root := findRepoRoot(t)
	filePath := filepath.Join(root, "data", "response.txt")

	transactions, header, err := parser.ParseFile(filePath, "P13")
	if err != nil {
		t.Fatalf("ParseFile() unexpected error: %v", err)
	}
	if header == nil {
		t.Fatal("ParseFile() returned nil header")
	}
	if header.DBID == "" {
		t.Fatalf("ParseFile() header.DBID is empty")
	}
	if header.ReportNum == "" {
		t.Fatalf("ParseFile() header.ReportNum is empty")
	}
	if len(transactions) == 0 {
		t.Fatalf("ParseFile() returned no transactions")
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "data", "response.txt")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Fatalf("repo root not found from %s", wd)
	return ""
}
