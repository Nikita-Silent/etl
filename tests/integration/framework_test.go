//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/user/go-frontol-loader/pkg/models"
	"github.com/user/go-frontol-loader/tests/integration/framework"
)

func TestFramework_BasicSetup(t *testing.T) {
	env := framework.SetupTestEnvironment(t)

	// Verify PostgreSQL connection
	ctx := env.GetContext()
	if err := env.Postgres.Pool.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Verify FTP connection
	if env.FTP.Client == nil {
		t.Fatal("FTP client is nil")
	}

	t.Logf("PostgreSQL DSN: %s", env.Postgres.GetDSN())
	t.Logf("FTP connection: %s", env.FTP.GetConnectionString())
}

func TestFramework_DataBuilder(t *testing.T) {
	env := framework.SetupTestEnvironment(t)
	env.Reset(t)

	ctx := env.GetContext()

	// Create test transaction
	tr := &models.TxItemRegistration1_11{
		TransactionIDUnique:        123,
		SourceFolder:               "001/folder1",
		TransactionDate:            time.Date(2024, 12, 18, 0, 0, 0, 0, time.UTC),
		TransactionTime:            time.Date(2000, 1, 1, 10, 30, 0, 0, time.UTC),
		TransactionType:            1,
		CashRegisterCode:           1,
		DocumentNumber:             1,
		CashierCode:                101,
		ItemIdentifier:             "TEST001",
		DimensionValueCodes:        "GROUP1",
		PriceWithoutDiscounts:      100.50,
		Quantity:                   2,
		PositionAmountWithRounding: 100.50,
		OperationType:              1,
		ShiftNumber:                1,
	}

	if err := env.Builder.CreateTransactionRegistration(ctx, tr); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Verify transaction was created
	count, err := env.Builder.CountTransactions(ctx, "tx_item_registration_1_11")
	if err != nil {
		t.Fatalf("Failed to count transactions: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 transaction, got %d", count)
	}

	// Retrieve and verify transaction
	retrieved, err := env.Builder.GetTransaction(ctx, 123, "001/folder1")
	if err != nil {
		t.Fatalf("Failed to get transaction: %v", err)
	}

	if retrieved.ItemIdentifier != "TEST001" {
		t.Errorf("Expected item code TEST001, got %s", retrieved.ItemIdentifier)
	}
	if retrieved.PriceWithoutDiscounts != 100.50 {
		t.Errorf("Expected amount 100.50, got %f", retrieved.PriceWithoutDiscounts)
	}
}

func TestFramework_SeedData(t *testing.T) {
	env := framework.SetupTestEnvironment(t)
	env.Reset(t)
	env.LoadBasicData(t)

	ctx := env.GetContext()

	// Verify basic data was loaded
	count, err := env.Builder.CountTransactions(ctx, "tx_item_registration_1_11")
	if err != nil {
		t.Fatalf("Failed to count transactions: %v", err)
	}

	if count < 1 {
		t.Errorf("Expected at least 1 transaction from basic data, got %d", count)
	}
}

func TestFramework_Reset(t *testing.T) {
	env := framework.SetupTestEnvironment(t)
	ctx := env.GetContext()

	// Create some data
	tr := &models.TxItemRegistration1_11{
		TransactionIDUnique: 999,
		SourceFolder:        "001/folder1",
		TransactionDate:     time.Date(2024, 12, 18, 0, 0, 0, 0, time.UTC),
		TransactionTime:     time.Date(2000, 1, 1, 10, 30, 0, 0, time.UTC),
		TransactionType:     1,
		CashRegisterCode:    1,
		DocumentNumber:      1,
		CashierCode:         101,
	}

	if err := env.Builder.CreateTransactionRegistration(ctx, tr); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Verify data exists
	count, err := env.Builder.CountTransactions(ctx, "tx_item_registration_1_11")
	if err != nil {
		t.Fatalf("Failed to count transactions: %v", err)
	}
	if count == 0 {
		t.Fatal("Expected at least 1 transaction before reset")
	}

	// Reset environment
	env.Reset(t)

	// Verify data was cleared
	count, err = env.Builder.CountTransactions(ctx, "tx_item_registration_1_11")
	if err != nil {
		t.Fatalf("Failed to count transactions after reset: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 transactions after reset, got %d", count)
	}
}

func TestFramework_FTPFileCreation(t *testing.T) {
	env := framework.SetupTestEnvironment(t)
	env.Reset(t)

	ctx := env.GetContext()

	// Create test file
	content := `#
DB001
REPORT001
1;2024-12-18;10:30:00;1;1;1;101;ITEM001;GROUP1;100.50;2;100.50;1;1;`

	err := env.Builder.CreateFTPFile(ctx, "001", "folder1", "test.txt", content)
	if err != nil {
		t.Fatalf("Failed to create FTP file: %v", err)
	}

	// Verify file exists on FTP
	files, err := env.FTP.Client.ListFiles("/response/001/folder1")
	if err != nil {
		t.Fatalf("Failed to list FTP files: %v", err)
	}

	found := false
	for _, file := range files {
		if file.Name == "test.txt" {
			found = true
			break
		}
	}

	if !found {
		t.Error("test.txt not found on FTP server")
	}
}

func TestFramework_MultipleTransactionTypes(t *testing.T) {
	env := framework.SetupTestEnvironment(t)
	env.Reset(t)

	ctx := env.GetContext()

	// Create multiple transaction types
	transactions := []models.TxItemRegistration1_11{
		{
			TransactionIDUnique: 1,
			SourceFolder:        "001/folder1",
			TransactionType:     1,
			CashRegisterCode:    1,
			DocumentNumber:      1,
			CashierCode:         101,
			ItemIdentifier:      "ITEM001",
		},
		{
			TransactionIDUnique: 2,
			SourceFolder:        "001/folder1",
			TransactionType:     1,
			CashRegisterCode:    1,
			DocumentNumber:      2,
			CashierCode:         102,
			ItemIdentifier:      "ITEM002",
		},
		{
			TransactionIDUnique: 3,
			SourceFolder:        "002/folder1",
			TransactionType:     1,
			CashRegisterCode:    2,
			DocumentNumber:      1,
			CashierCode:         201,
			ItemIdentifier:      "ITEM003",
		},
	}

	for _, tr := range transactions {
		if err := env.Builder.CreateTransactionRegistration(ctx, &tr); err != nil {
			t.Fatalf("Failed to create transaction %d: %v", tr.TransactionIDUnique, err)
		}
	}

	// Verify count
	count, err := env.Builder.CountTransactions(ctx, "tx_item_registration_1_11")
	if err != nil {
		t.Fatalf("Failed to count transactions: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 transactions, got %d", count)
	}
}
