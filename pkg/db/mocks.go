package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/user/go-frontol-loader/pkg/models"
)

// MockPool is a mock implementation of DatabasePool for testing
type MockPool struct {
	BeginTxFunc                         func(ctx context.Context) (pgx.Tx, error)
	QueryFunc                           func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRowFunc                        func(ctx context.Context, sql string, args ...interface{}) pgx.Row
	LoadDataFunc                        func(ctx context.Context, tx pgx.Tx, tableName string, columns []string, rows [][]interface{}) error
	LoadTransactionRegistrationsFunc    func(ctx context.Context, tx pgx.Tx, transactions []models.TransactionRegistration) error
	LoadSpecialPricesFunc               func(ctx context.Context, tx pgx.Tx, prices []models.SpecialPrice) error
	LoadBonusTransactionsFunc           func(ctx context.Context, tx pgx.Tx, transactions []models.BonusTransaction) error
	LoadDiscountTransactionsFunc        func(ctx context.Context, tx pgx.Tx, transactions []models.DiscountTransaction) error
	LoadBillRegistrationsFunc           func(ctx context.Context, tx pgx.Tx, bills []models.BillRegistration) error
	LoadEmployeeEditsFunc               func(ctx context.Context, tx pgx.Tx, edits []models.EmployeeEdit) error
	LoadEmployeeAccountingFunc          func(ctx context.Context, tx pgx.Tx, accounting []models.EmployeeAccounting) error
	LoadVatKKTTransactionsFunc          func(ctx context.Context, tx pgx.Tx, transactions []models.VatKKTTransaction) error
	LoadAdditionalTransactionsFunc      func(ctx context.Context, tx pgx.Tx, transactions []models.AdditionalTransaction) error
	LoadAstuExchangeTransactionsFunc    func(ctx context.Context, tx pgx.Tx, transactions []models.AstuExchangeTransaction) error
	LoadCounterChangeTransactionsFunc   func(ctx context.Context, tx pgx.Tx, transactions []models.CounterChangeTransaction) error
	LoadKKTShiftReportsFunc             func(ctx context.Context, tx pgx.Tx, reports []models.KKTShiftReport) error
	LoadFrontolMarkUnitTransactionsFunc func(ctx context.Context, tx pgx.Tx, transactions []models.FrontolMarkUnitTransaction) error
	LoadBonusPaymentsFunc               func(ctx context.Context, tx pgx.Tx, payments []models.BonusPayment) error
	LoadDocumentOperationsFunc          func(ctx context.Context, tx pgx.Tx, operations []models.DocumentOperation) error
	LoadDocumentDiscountsFunc           func(ctx context.Context, tx pgx.Tx, discounts []models.DocumentDiscount) error
	LoadFiscalPaymentsFunc              func(ctx context.Context, tx pgx.Tx, payments []models.FiscalPayment) error
	LoadCardStatusChangesFunc           func(ctx context.Context, tx pgx.Tx, changes []models.CardStatusChange) error
	LoadModifierTransactionsFunc        func(ctx context.Context, tx pgx.Tx, transactions []models.ModifierTransaction) error
	LoadPrepaymentTransactionsFunc      func(ctx context.Context, tx pgx.Tx, transactions []models.PrepaymentTransaction) error
	LoadNonFiscalPaymentsFunc           func(ctx context.Context, tx pgx.Tx, payments []models.NonFiscalPayment) error
}

func (m *MockPool) Close() {}

func (m *MockPool) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx)
	}
	return nil, nil
}

func (m *MockPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func (m *MockPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, sql, args...)
	}
	return nil
}

func (m *MockPool) LoadData(ctx context.Context, tx pgx.Tx, tableName string, columns []string, rows [][]interface{}) error {
	if m.LoadDataFunc != nil {
		return m.LoadDataFunc(ctx, tx, tableName, columns, rows)
	}
	return nil
}

func (m *MockPool) LoadTransactionRegistrations(ctx context.Context, tx pgx.Tx, transactions []models.TransactionRegistration) error {
	if m.LoadTransactionRegistrationsFunc != nil {
		return m.LoadTransactionRegistrationsFunc(ctx, tx, transactions)
	}
	return nil
}

func (m *MockPool) LoadSpecialPrices(ctx context.Context, tx pgx.Tx, prices []models.SpecialPrice) error {
	if m.LoadSpecialPricesFunc != nil {
		return m.LoadSpecialPricesFunc(ctx, tx, prices)
	}
	return nil
}

func (m *MockPool) LoadBonusTransactions(ctx context.Context, tx pgx.Tx, transactions []models.BonusTransaction) error {
	if m.LoadBonusTransactionsFunc != nil {
		return m.LoadBonusTransactionsFunc(ctx, tx, transactions)
	}
	return nil
}

func (m *MockPool) LoadDiscountTransactions(ctx context.Context, tx pgx.Tx, transactions []models.DiscountTransaction) error {
	if m.LoadDiscountTransactionsFunc != nil {
		return m.LoadDiscountTransactionsFunc(ctx, tx, transactions)
	}
	return nil
}

func (m *MockPool) LoadBillRegistrations(ctx context.Context, tx pgx.Tx, bills []models.BillRegistration) error {
	if m.LoadBillRegistrationsFunc != nil {
		return m.LoadBillRegistrationsFunc(ctx, tx, bills)
	}
	return nil
}

func (m *MockPool) LoadEmployeeEdits(ctx context.Context, tx pgx.Tx, edits []models.EmployeeEdit) error {
	if m.LoadEmployeeEditsFunc != nil {
		return m.LoadEmployeeEditsFunc(ctx, tx, edits)
	}
	return nil
}

func (m *MockPool) LoadEmployeeAccounting(ctx context.Context, tx pgx.Tx, accounting []models.EmployeeAccounting) error {
	if m.LoadEmployeeAccountingFunc != nil {
		return m.LoadEmployeeAccountingFunc(ctx, tx, accounting)
	}
	return nil
}

func (m *MockPool) LoadVatKKTTransactions(ctx context.Context, tx pgx.Tx, transactions []models.VatKKTTransaction) error {
	if m.LoadVatKKTTransactionsFunc != nil {
		return m.LoadVatKKTTransactionsFunc(ctx, tx, transactions)
	}
	return nil
}

func (m *MockPool) LoadAdditionalTransactions(ctx context.Context, tx pgx.Tx, transactions []models.AdditionalTransaction) error {
	if m.LoadAdditionalTransactionsFunc != nil {
		return m.LoadAdditionalTransactionsFunc(ctx, tx, transactions)
	}
	return nil
}

func (m *MockPool) LoadAstuExchangeTransactions(ctx context.Context, tx pgx.Tx, transactions []models.AstuExchangeTransaction) error {
	if m.LoadAstuExchangeTransactionsFunc != nil {
		return m.LoadAstuExchangeTransactionsFunc(ctx, tx, transactions)
	}
	return nil
}

func (m *MockPool) LoadCounterChangeTransactions(ctx context.Context, tx pgx.Tx, transactions []models.CounterChangeTransaction) error {
	if m.LoadCounterChangeTransactionsFunc != nil {
		return m.LoadCounterChangeTransactionsFunc(ctx, tx, transactions)
	}
	return nil
}

func (m *MockPool) LoadKKTShiftReports(ctx context.Context, tx pgx.Tx, reports []models.KKTShiftReport) error {
	if m.LoadKKTShiftReportsFunc != nil {
		return m.LoadKKTShiftReportsFunc(ctx, tx, reports)
	}
	return nil
}

func (m *MockPool) LoadFrontolMarkUnitTransactions(ctx context.Context, tx pgx.Tx, transactions []models.FrontolMarkUnitTransaction) error {
	if m.LoadFrontolMarkUnitTransactionsFunc != nil {
		return m.LoadFrontolMarkUnitTransactionsFunc(ctx, tx, transactions)
	}
	return nil
}

func (m *MockPool) LoadBonusPayments(ctx context.Context, tx pgx.Tx, payments []models.BonusPayment) error {
	if m.LoadBonusPaymentsFunc != nil {
		return m.LoadBonusPaymentsFunc(ctx, tx, payments)
	}
	return nil
}

func (m *MockPool) LoadDocumentOperations(ctx context.Context, tx pgx.Tx, operations []models.DocumentOperation) error {
	if m.LoadDocumentOperationsFunc != nil {
		return m.LoadDocumentOperationsFunc(ctx, tx, operations)
	}
	return nil
}

func (m *MockPool) LoadDocumentDiscounts(ctx context.Context, tx pgx.Tx, discounts []models.DocumentDiscount) error {
	if m.LoadDocumentDiscountsFunc != nil {
		return m.LoadDocumentDiscountsFunc(ctx, tx, discounts)
	}
	return nil
}

func (m *MockPool) LoadFiscalPayments(ctx context.Context, tx pgx.Tx, payments []models.FiscalPayment) error {
	if m.LoadFiscalPaymentsFunc != nil {
		return m.LoadFiscalPaymentsFunc(ctx, tx, payments)
	}
	return nil
}

func (m *MockPool) LoadCardStatusChanges(ctx context.Context, tx pgx.Tx, changes []models.CardStatusChange) error {
	if m.LoadCardStatusChangesFunc != nil {
		return m.LoadCardStatusChangesFunc(ctx, tx, changes)
	}
	return nil
}

func (m *MockPool) LoadModifierTransactions(ctx context.Context, tx pgx.Tx, transactions []models.ModifierTransaction) error {
	if m.LoadModifierTransactionsFunc != nil {
		return m.LoadModifierTransactionsFunc(ctx, tx, transactions)
	}
	return nil
}

func (m *MockPool) LoadPrepaymentTransactions(ctx context.Context, tx pgx.Tx, transactions []models.PrepaymentTransaction) error {
	if m.LoadPrepaymentTransactionsFunc != nil {
		return m.LoadPrepaymentTransactionsFunc(ctx, tx, transactions)
	}
	return nil
}

func (m *MockPool) LoadNonFiscalPayments(ctx context.Context, tx pgx.Tx, payments []models.NonFiscalPayment) error {
	if m.LoadNonFiscalPaymentsFunc != nil {
		return m.LoadNonFiscalPaymentsFunc(ctx, tx, payments)
	}
	return nil
}
