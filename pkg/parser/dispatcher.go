package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/user/go-frontol-loader/pkg/models"
)

// TransactionType represents a transaction type with its associated parser
type TransactionType struct {
	Type        int
	Description string
	Parser      func([]string, string) (interface{}, error)
}

// GetTransactionType returns the transaction type for a given type code
// Based on Frontol 6 Integration documentation (frontol_6_integration.md)
func GetTransactionType(typeCode int) (*TransactionType, error) {
	switch typeCode {
	// Регистрация товара (стр. 266-269)
	case 1, 11:
		return &TransactionType{
			Type:        typeCode,
			Description: "Регистрация товара",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				return wrapParsed[models.TxItemRegistration1_11](fields, sourceFolder, "tx_item_registration_1_11")
			},
		}, nil
	case 2, 12:
		return &TransactionType{
			Type:        typeCode,
			Description: "Сторно товара",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				return wrapParsed[models.TxItemStorno2_12](fields, sourceFolder, "tx_item_storno_2_12")
			},
		}, nil
	case 4, 14:
		return &TransactionType{
			Type:        typeCode,
			Description: "Налог на товар",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				return wrapParsed[models.TxItemTax4_14](fields, sourceFolder, "tx_item_tax_4_14")
			},
		}, nil
	case 6, 16:
		return &TransactionType{
			Type:        typeCode,
			Description: "ККТ регистрация",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				return wrapParsed[models.TxItemKKT6_16](fields, sourceFolder, "tx_item_kkt_6_16")
			},
		}, nil

	// Установка спеццены/цены из прайс-листа (стр. 271-272)
	case 3:
		return &TransactionType{
			Type:        typeCode,
			Description: "Установка спеццены",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				return wrapParsed[models.TxSpecialPrice3](fields, sourceFolder, "tx_special_price_3")
			},
		}, nil

	// Начисление и возврат бонуса (стр. 273)
	case 9, 10:
		return &TransactionType{
			Type:        typeCode,
			Description: "Начисление и возврат бонуса",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_bonus_accrual_9"
				if typeCode == 10 {
					table = "tx_bonus_refund_10"
				}
				return wrapParsed[models.TxBonusAccrual9](fields, sourceFolder, table)
			},
		}, nil

	// Скидки на позицию (стр. 275)
	case 15, 17:
		return &TransactionType{
			Type:        typeCode,
			Description: "Скидки на позицию",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_position_discount_15"
				if typeCode == 17 {
					table = "tx_position_discount_17"
				}
				return wrapParsed[models.TxPositionDiscount15](fields, sourceFolder, table)
			},
		}, nil

	// Регистрация купюр (стр. 277)
	case 21, 22, 23, 24:
		return &TransactionType{
			Type:        typeCode,
			Description: "Регистрация купюр",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_bill_registration_21_23"
				if typeCode == 22 || typeCode == 24 {
					table = "tx_bill_storno_22_24"
				}
				return wrapParsed[models.TxBillRegistration21_23](fields, sourceFolder, table)
			},
		}, nil

	// Регистрация сотрудников в документе редактирования списка сотрудников (стр. 278)
	case 25:
		return &TransactionType{
			Type:        typeCode,
			Description: "Регистрация сотрудников",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				return wrapParsed[models.TxEmployeeRegistration25](fields, sourceFolder, "tx_employee_registration_25")
			},
		}, nil

	// Учет сотрудников по документу/позиции (стр. 279)
	case 26, 29:
		return &TransactionType{
			Type:        typeCode,
			Description: "Учет сотрудников",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_employee_accounting_doc_26"
				if typeCode == 29 {
					table = "tx_employee_accounting_pos_29"
				}
				return wrapParsed[models.TxEmployeeAccountingDoc26](fields, sourceFolder, table)
			},
		}, nil

	// Изменение статуса карты (стр. 281)
	case 27:
		return &TransactionType{
			Type:        typeCode,
			Description: "Изменение статуса карты",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				return wrapParsed[models.TxCardStatusChange27](fields, sourceFolder, "tx_card_status_change_27")
			},
		}, nil

	// Регистрация/сторнирование модификаторов (стр. 283)
	case 30, 31:
		return &TransactionType{
			Type:        typeCode,
			Description: "Модификаторы",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_modifier_registration_30"
				if typeCode == 31 {
					table = "tx_modifier_storno_31"
				}
				return wrapParsed[models.TxModifierRegistration30](fields, sourceFolder, table)
			},
		}, nil

	// Оплата и возврат оплаты бонусом (стр. 284)
	case 32, 33, 82, 83:
		return &TransactionType{
			Type:        typeCode,
			Description: "Оплата бонусом",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_bonus_payment_32"
				switch typeCode {
				case 33:
					table = "tx_bonus_payment_33"
				case 82:
					table = "tx_bonus_payment_82"
				case 83:
					table = "tx_bonus_payment_83"
				}
				return wrapParsed[models.TxBonusPayment32](fields, sourceFolder, table)
			},
		}, nil

	// Предоплата документом (стр. 286)
	case 34, 84:
		return &TransactionType{
			Type:        typeCode,
			Description: "Предоплата документом",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_prepayment_34"
				if typeCode == 84 {
					table = "tx_prepayment_84"
				}
				return wrapParsed[models.TxPrepayment34](fields, sourceFolder, table)
			},
		}, nil

	// Скидки на документ (стр. 288)
	case 35, 37, 38, 85, 87:
		return &TransactionType{
			Type:        typeCode,
			Description: "Скидки на документ",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_document_discount_35"
				switch typeCode {
				case 37:
					table = "tx_document_discount_37"
				case 38:
					table = "tx_document_rounding_38"
				case 85:
					table = "tx_document_discount_85"
				case 87:
					table = "tx_document_discount_87"
				}
				return wrapParsed[models.TxDocumentDiscount35](fields, sourceFolder, table)
			},
		}, nil

	// Нефискальная оплата (стр. 290)
	case 36, 86:
		return &TransactionType{
			Type:        typeCode,
			Description: "Нефискальная оплата",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_non_fiscal_payment_36"
				if typeCode == 86 {
					table = "tx_non_fiscal_payment_86"
				}
				return wrapParsed[models.TxNonFiscalPayment36](fields, sourceFolder, table)
			},
		}, nil

	// Фискальная оплата (стр. 292)
	case 40, 43:
		return &TransactionType{
			Type:        typeCode,
			Description: "Фискальная оплата",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_fiscal_payment_40"
				if typeCode == 43 {
					table = "tx_fiscal_payment_43"
				}
				return wrapParsed[models.TxFiscalPayment40](fields, sourceFolder, table)
			},
		}, nil

	// Открытие/закрытие документа (стр. 293)
	case 42, 45, 49, 55, 56, 58, 65, 120:
		return &TransactionType{
			Type:        typeCode,
			Description: "Открытие/закрытие документа",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_document_open_42"
				switch typeCode {
				case 45:
					table = "tx_document_close_kkt_45"
				case 49:
					table = "tx_document_close_gp_49"
				case 55:
					table = "tx_document_close_55"
				case 56:
					table = "tx_document_cancel_56"
				case 58:
					table = "tx_document_non_fin_close_58"
				case 65:
					table = "tx_document_clients_65"
				case 120:
					table = "tx_document_egais_120"
				}
				return wrapParsed[models.TxDocumentOpen42](fields, sourceFolder, table)
			},
		}, nil

	// НДС по чеку из ККТ (стр. 299)
	case 88:
		return &TransactionType{
			Type:        typeCode,
			Description: "НДС по чеку из ККТ",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				return wrapParsed[models.TxVATKKT88](fields, sourceFolder, "tx_vat_kkt_88")
			},
		}, nil

	// Дополнительные транзакции (Внесение/Выплата) (стр. 300)
	case 50, 51:
		return &TransactionType{
			Type:        typeCode,
			Description: "Внесение/Выплата",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_cash_in_50"
				if typeCode == 51 {
					table = "tx_cash_out_51"
				}
				return wrapParsed[models.TxCashIn50](fields, sourceFolder, table)
			},
		}, nil

	// Изменение счетчика (стр. 302)
	case 57:
		return &TransactionType{
			Type:        typeCode,
			Description: "Изменение счетчика",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				return wrapParsed[models.TxCounterChange57](fields, sourceFolder, "tx_counter_change_57")
			},
		}, nil

	// Отчеты (стр. 304)
	case 60, 61, 62, 63, 64:
		return &TransactionType{
			Type:        typeCode,
			Description: "Отчеты",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				table := "tx_report_zless_60"
				switch typeCode {
				case 61:
					table = "tx_shift_close_61"
				case 62:
					table = "tx_shift_open_62"
				case 63:
					table = "tx_report_z_63"
				case 64:
					table = "tx_shift_open_doc_64"
				}
				return wrapParsed[models.TxReportZless60](fields, sourceFolder, table)
			},
		}, nil

	// Отправка данных во Frontol Mark Unit (стр. 306)
	case 121:
		return &TransactionType{
			Type:        typeCode,
			Description: "Frontol Mark Unit",
			Parser: func(fields []string, sourceFolder string) (interface{}, error) {
				return wrapParsed[models.TxMarkUnit121](fields, sourceFolder, "tx_mark_unit_121")
			},
		}, nil

	default:
		return nil, fmt.Errorf("unhandled transaction type: %d", typeCode)
	}
}

// ParseTransactionLine parses a single transaction line using the dispatcher
func ParseTransactionLine(line string, sourceFolder string) (interface{}, error) {
	fields := strings.Split(line, ";")
	if len(fields) < 4 {
		return nil, fmt.Errorf("insufficient fields in line")
	}

	// Get transaction type from field 4 (index 3)
	transactionType, err := strconv.Atoi(fields[3])
	if err != nil {
		return nil, fmt.Errorf("invalid transaction type: %s", fields[3])
	}

	// Get transaction type definition
	txType, err := GetTransactionType(transactionType)
	if err != nil {
		return nil, err
	}

	// Parse using the appropriate parser
	return txType.Parser(fields, sourceFolder)
}

// GetSupportedTransactionTypes returns a list of all supported transaction types
func GetSupportedTransactionTypes() []TransactionType {
	return []TransactionType{
		{Type: 1, Description: "Регистрация товара"},
		{Type: 2, Description: "Регистрация товара"},
		{Type: 3, Description: "Установка спеццены"},
		{Type: 4, Description: "Регистрация товара"},
		{Type: 6, Description: "Регистрация товара"},
		{Type: 9, Description: "Начисление и возврат бонуса"},
		{Type: 10, Description: "Начисление и возврат бонуса"},
		{Type: 11, Description: "Регистрация товара"},
		{Type: 12, Description: "Регистрация товара"},
		{Type: 14, Description: "Регистрация товара"},
		{Type: 15, Description: "Скидки на позицию"},
		{Type: 16, Description: "Регистрация товара"},
		{Type: 17, Description: "Скидки на позицию"},
		{Type: 21, Description: "Регистрация купюр"},
		{Type: 22, Description: "Регистрация купюр"},
		{Type: 23, Description: "Регистрация купюр"},
		{Type: 24, Description: "Регистрация купюр"},
		{Type: 25, Description: "Регистрация сотрудников"},
		{Type: 26, Description: "Учет сотрудников"},
		{Type: 27, Description: "Изменение статуса карты"},
		{Type: 29, Description: "Учет сотрудников"},
		{Type: 30, Description: "Модификаторы"},
		{Type: 31, Description: "Модификаторы"},
		{Type: 32, Description: "Оплата бонусом"},
		{Type: 33, Description: "Оплата бонусом"},
		{Type: 34, Description: "Предоплата документом"},
		{Type: 35, Description: "Скидки на документ"},
		{Type: 36, Description: "Нефискальная оплата"},
		{Type: 37, Description: "Скидки на документ"},
		{Type: 38, Description: "Скидки на документ"},
		{Type: 40, Description: "Фискальная оплата"},
		{Type: 42, Description: "Открытие/закрытие документа"},
		{Type: 43, Description: "Фискальная оплата"},
		{Type: 45, Description: "Открытие/закрытие документа"},
		{Type: 49, Description: "Открытие/закрытие документа"},
		{Type: 50, Description: "Внесение/Выплата"},
		{Type: 51, Description: "Внесение/Выплата"},
		{Type: 55, Description: "Открытие/закрытие документа"},
		{Type: 56, Description: "Открытие/закрытие документа"},
		{Type: 57, Description: "Изменение счетчика"},
		{Type: 58, Description: "Открытие/закрытие документа"},
		{Type: 60, Description: "Отчеты"},
		{Type: 61, Description: "Отчеты"},
		{Type: 62, Description: "Отчеты"},
		{Type: 63, Description: "Отчеты"},
		{Type: 64, Description: "Отчеты"},
		{Type: 65, Description: "Открытие/закрытие документа"},
		{Type: 82, Description: "Оплата бонусом"},
		{Type: 83, Description: "Оплата бонусом"},
		{Type: 84, Description: "Предоплата документом"},
		{Type: 85, Description: "Скидки на документ"},
		{Type: 86, Description: "Нефискальная оплата"},
		{Type: 87, Description: "Скидки на документ"},
		{Type: 88, Description: "НДС по чеку из ККТ"},
		{Type: 120, Description: "Открытие/закрытие документа"},
		{Type: 121, Description: "Frontol Mark Unit"},
	}
}

// ValidateTransactionType checks if a transaction type is supported
func ValidateTransactionType(typeCode int) bool {
	_, err := GetTransactionType(typeCode)
	return err == nil
}

// GetTransactionTypeDescription returns the description for a transaction type
func GetTransactionTypeDescription(typeCode int) string {
	txType, err := GetTransactionType(typeCode)
	if err != nil {
		return "Неизвестный тип транзакции"
	}
	return txType.Description
}
