package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/user/go-frontol-loader/pkg/models"
)

// DatabasePool defines the interface for database operations
// This allows for easier testing with mocks
type DatabasePool interface {
	Close()
	BeginTx(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row

	// Load methods
	LoadData(ctx context.Context, tx pgx.Tx, tableName string, columns []string, rows [][]interface{}) error
	LoadTransactionRegistrations(ctx context.Context, tx pgx.Tx, transactions []models.TransactionRegistration) error
	LoadSpecialPrices(ctx context.Context, tx pgx.Tx, prices []models.SpecialPrice) error
	LoadBonusTransactions(ctx context.Context, tx pgx.Tx, transactions []models.BonusTransaction) error
	LoadDiscountTransactions(ctx context.Context, tx pgx.Tx, transactions []models.DiscountTransaction) error
	LoadBillRegistrations(ctx context.Context, tx pgx.Tx, bills []models.BillRegistration) error
	LoadEmployeeEdits(ctx context.Context, tx pgx.Tx, edits []models.EmployeeEdit) error
	LoadEmployeeAccounting(ctx context.Context, tx pgx.Tx, accounting []models.EmployeeAccounting) error
	LoadVatKKTTransactions(ctx context.Context, tx pgx.Tx, transactions []models.VatKKTTransaction) error
	LoadAdditionalTransactions(ctx context.Context, tx pgx.Tx, transactions []models.AdditionalTransaction) error
	LoadAstuExchangeTransactions(ctx context.Context, tx pgx.Tx, transactions []models.AstuExchangeTransaction) error
	LoadCounterChangeTransactions(ctx context.Context, tx pgx.Tx, transactions []models.CounterChangeTransaction) error
	LoadKKTShiftReports(ctx context.Context, tx pgx.Tx, reports []models.KKTShiftReport) error
	LoadFrontolMarkUnitTransactions(ctx context.Context, tx pgx.Tx, transactions []models.FrontolMarkUnitTransaction) error
	LoadBonusPayments(ctx context.Context, tx pgx.Tx, payments []models.BonusPayment) error
	LoadDocumentOperations(ctx context.Context, tx pgx.Tx, operations []models.DocumentOperation) error
	LoadDocumentDiscounts(ctx context.Context, tx pgx.Tx, discounts []models.DocumentDiscount) error
	LoadFiscalPayments(ctx context.Context, tx pgx.Tx, payments []models.FiscalPayment) error
	LoadCardStatusChanges(ctx context.Context, tx pgx.Tx, changes []models.CardStatusChange) error
	LoadModifierTransactions(ctx context.Context, tx pgx.Tx, transactions []models.ModifierTransaction) error
	LoadPrepaymentTransactions(ctx context.Context, tx pgx.Tx, transactions []models.PrepaymentTransaction) error
	LoadNonFiscalPayments(ctx context.Context, tx pgx.Tx, payments []models.NonFiscalPayment) error
}

// Ensure Pool implements DatabasePool interface
var _ DatabasePool = (*Pool)(nil)
