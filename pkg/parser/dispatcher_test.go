package parser

import (
	"fmt"
	"math"
	"testing"

	"github.com/user/go-frontol-loader/pkg/models"
)

func TestGetTransactionTypeSupported(t *testing.T) {
	supported := GetSupportedTransactionTypes()
	for _, tt := range supported {
		t.Run(fmt.Sprintf("type_%d", tt.Type), func(t *testing.T) {
			txType, err := GetTransactionType(tt.Type)
			if err != nil {
				t.Fatalf("GetTransactionType(%d) unexpected error: %v", tt.Type, err)
			}
			if txType == nil {
				t.Fatalf("GetTransactionType(%d) returned nil", tt.Type)
			}
			if txType.Type != tt.Type {
				t.Fatalf("GetTransactionType(%d).Type = %d", tt.Type, txType.Type)
			}
			if txType.Parser == nil {
				t.Fatalf("GetTransactionType(%d) parser is nil", tt.Type)
			}
			if _, err := txType.Parser([]string{}, "source"); err != nil {
				t.Fatalf("Parser(%d) unexpected error: %v", tt.Type, err)
			}
		})
	}
}

func TestGetTransactionTypeUnsupported(t *testing.T) {
	invalidTypes := []int{-1, 0, 5, 7, 8, 9999, math.MaxInt}
	for _, typeCode := range invalidTypes {
		t.Run(fmt.Sprintf("type_%d", typeCode), func(t *testing.T) {
			if _, err := GetTransactionType(typeCode); err == nil {
				t.Fatalf("GetTransactionType(%d) expected error, got nil", typeCode)
			}
		})
	}
}

func TestSupportedTypesCount(t *testing.T) {
	supported := GetSupportedTransactionTypes()
	if len(supported) < len(models.TxSchemas) {
		t.Fatalf("supported types = %d, want >= %d (TxSchemas)", len(supported), len(models.TxSchemas))
	}
}
