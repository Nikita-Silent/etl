//go:build integration
// +build integration

package framework

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/go-frontol-loader/pkg/models"
)

// TestDataBuilder provides methods for building test data
type TestDataBuilder struct {
	postgres *PostgresContainer
	ftp      *FTPContainer
}

// NewTestDataBuilder creates a new test data builder
func NewTestDataBuilder(postgres *PostgresContainer, ftp *FTPContainer) *TestDataBuilder {
	return &TestDataBuilder{
		postgres: postgres,
		ftp:      ftp,
	}
}

// CreateTransactionRegistration inserts a test tx_item_registration_1_11 row.
func (b *TestDataBuilder) CreateTransactionRegistration(ctx context.Context, tr *models.TxItemRegistration1_11) error {
	query := `
		INSERT INTO tx_item_registration_1_11 (
			transaction_id_unique, source_folder, transaction_date, transaction_time,
			transaction_type, cash_register_code, document_number, cashier_code,
			item_identifier, dimension_value_codes, price_without_discounts, quantity, position_amount_with_rounding,
			operation_type, shift_number
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := b.postgres.Pool.Exec(ctx, query,
		tr.TransactionIDUnique, tr.SourceFolder, tr.TransactionDate, tr.TransactionTime,
		tr.TransactionType, tr.CashRegisterCode, tr.DocumentNumber, tr.CashierCode,
		tr.ItemIdentifier, tr.DimensionValueCodes, tr.PriceWithoutDiscounts, tr.Quantity, tr.PositionAmountWithRounding,
		tr.OperationType, tr.ShiftNumber,
	)

	return err
}

// CreateFTPFile creates a test file on the FTP server
func (b *TestDataBuilder) CreateFTPFile(ctx context.Context, kassaCode, folderName, filename, content string) error {
	// Create local temporary file
	tmpDir := "/tmp/frontol_test_fixtures"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	tmpFile := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Upload to FTP
	remotePath := fmt.Sprintf("/response/%s/%s/%s", kassaCode, folderName, filename)
	if err := b.ftp.Client.UploadFile(tmpFile, remotePath); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// SeedDataSet represents a set of seed data
type SeedDataSet struct {
	Name         string
	Transactions []models.TxItemRegistration1_11
	Files        []TestFile
}

// TestFile represents a test file to create
type TestFile struct {
	KassaCode  string
	FolderName string
	Filename   string
	Content    string
}

// LoadSeedData loads a predefined seed data set
func (b *TestDataBuilder) LoadSeedData(ctx context.Context, dataset SeedDataSet) error {
	// Insert transactions
	for _, tr := range dataset.Transactions {
		if err := b.CreateTransactionRegistration(ctx, &tr); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}
	}

	// Create FTP files
	for _, file := range dataset.Files {
		if err := b.CreateFTPFile(ctx, file.KassaCode, file.FolderName, file.Filename, file.Content); err != nil {
			return fmt.Errorf("failed to create FTP file: %w", err)
		}
	}

	return nil
}

// GetBasicDataSet returns a basic test data set
func GetBasicDataSet() SeedDataSet {
	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	timeStr := time.Date(2000, 1, 1, now.Hour(), now.Minute(), now.Second(), 0, time.UTC)

	return SeedDataSet{
		Name: "basic",
		Transactions: []models.TxItemRegistration1_11{
			{
				TransactionIDUnique:        1,
				SourceFolder:               "001/folder1",
				TransactionDate:            date,
				TransactionTime:            timeStr,
				TransactionType:            1,
				CashRegisterCode:           1,
				DocumentNumber:             1,
				CashierCode:                101,
				ItemIdentifier:             "ITEM001",
				DimensionValueCodes:        "GROUP1",
				PriceWithoutDiscounts:      100.50,
				Quantity:                   2,
				PositionAmountWithRounding: 100.50,
				OperationType:              1,
				ShiftNumber:                1,
			},
			{
				TransactionIDUnique:        2,
				SourceFolder:               "001/folder1",
				TransactionDate:            date,
				TransactionTime:            timeStr,
				TransactionType:            1,
				CashRegisterCode:           1,
				DocumentNumber:             2,
				CashierCode:                101,
				ItemIdentifier:             "ITEM002",
				DimensionValueCodes:        "GROUP1",
				PriceWithoutDiscounts:      50.25,
				Quantity:                   1,
				PositionAmountWithRounding: 50.25,
				OperationType:              1,
				ShiftNumber:                1,
			},
		},
		Files: []TestFile{
			{
				KassaCode:  "001",
				FolderName: "folder1",
				Filename:   "test_data.txt",
				Content: `#
DB001
REPORT001
1;18.12.2024;10:30:00;1;1;1;101;ITEM001;GROUP1;100.50;2;100.50;1;1;`,
			},
		},
	}
}

// CountTransactions returns the number of transactions in the database
func (b *TestDataBuilder) CountTransactions(ctx context.Context, table string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	var count int
	err := b.postgres.Pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// GetTransaction retrieves a transaction by ID
func (b *TestDataBuilder) GetTransaction(ctx context.Context, id int64, sourceFolder string) (*models.TxItemRegistration1_11, error) {
	query := `
		SELECT transaction_id_unique, source_folder, transaction_date, transaction_time,
		       transaction_type, cash_register_code, document_number, cashier_code,
		       item_identifier, dimension_value_codes, price_without_discounts, quantity, position_amount_with_rounding,
		       operation_type, shift_number
		FROM tx_item_registration_1_11
		WHERE transaction_id_unique = $1 AND source_folder = $2
	`

	tr := &models.TxItemRegistration1_11{}
	err := b.postgres.Pool.QueryRow(ctx, query, id, sourceFolder).Scan(
		&tr.TransactionIDUnique, &tr.SourceFolder, &tr.TransactionDate, &tr.TransactionTime,
		&tr.TransactionType, &tr.CashRegisterCode, &tr.DocumentNumber, &tr.CashierCode,
		&tr.ItemIdentifier, &tr.DimensionValueCodes, &tr.PriceWithoutDiscounts, &tr.Quantity, &tr.PositionAmountWithRounding,
		&tr.OperationType, &tr.ShiftNumber,
	)

	return tr, err
}
