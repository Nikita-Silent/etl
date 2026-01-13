package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/user/go-frontol-loader/pkg/models"
)

// ParseFile parses a Frontol file and returns all transaction data grouped by type
func ParseFile(filePath string, sourceFolder string) (map[string]interface{}, *models.FileHeader, error) {
	// #nosec G304 -- filePath comes from configured input directories.
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	// Parse file header (first 3 lines)
	header, err := parseFileHeader(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse file header: %w", err)
	}

	// Reset file position to beginning and skip header
	_, _ = file.Seek(0, 0)
	scanner := bufio.NewScanner(file)

	// Skip first 3 lines (header)
	for i := 0; i < 3; i++ {
		if !scanner.Scan() {
			return nil, nil, fmt.Errorf("file has insufficient lines")
		}
	}

	// Parse transactions
	transactions, err := parseTransactions(scanner, sourceFolder)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse transactions: %w", err)
	}

	return transactions, header, nil
}

// parseFileHeader parses the first 3 lines of the file
func parseFileHeader(file *os.File) (*models.FileHeader, error) {
	scanner := bufio.NewScanner(file)

	// Read first line (processed flag)
	if !scanner.Scan() {
		return nil, fmt.Errorf("file is empty")
	}
	processedLine := strings.TrimSpace(scanner.Text())
	processed := processedLine == "1" || processedLine == "@"

	// Read second line (DB ID)
	if !scanner.Scan() {
		return nil, fmt.Errorf("file has only one line")
	}
	dbID := strings.TrimSpace(scanner.Text())

	// Read third line (report number)
	if !scanner.Scan() {
		return nil, fmt.Errorf("file has only two lines")
	}
	reportNum := strings.TrimSpace(scanner.Text())

	return &models.FileHeader{
		Processed: processed,
		DBID:      dbID,
		ReportNum: reportNum,
	}, nil
}

// parseTransactions parses all transaction lines in the file
func parseTransactions(scanner *bufio.Scanner, sourceFolder string) (map[string]interface{}, error) {

	// Initialize transaction collections (tx_* tables)
	txItemRegistrations := []models.TxItemRegistration1_11{}
	txItemStorno := []models.TxItemStorno2_12{}
	txItemTax := []models.TxItemTax4_14{}
	txItemKKT := []models.TxItemKKT6_16{}
	txSpecialPrices := []models.TxSpecialPrice3{}
	txBonusAccruals := []models.TxBonusAccrual9{}
	txBonusRefunds := []models.TxBonusRefund10{}
	txPositionDiscounts15 := []models.TxPositionDiscount15{}
	txPositionDiscounts17 := []models.TxPositionDiscount17{}
	txBillRegistrations := []models.TxBillRegistration21_23{}
	txBillStornos := []models.TxBillStorno22_24{}
	txEmployeeRegistrations := []models.TxEmployeeRegistration25{}
	txEmployeeAccountingDocs := []models.TxEmployeeAccountingDoc26{}
	txEmployeeAccountingPos := []models.TxEmployeeAccountingPos29{}
	txCardStatusChanges := []models.TxCardStatusChange27{}
	txModifierRegistrations := []models.TxModifierRegistration30{}
	txModifierStornos := []models.TxModifierStorno31{}
	txBonusPayments32 := []models.TxBonusPayment32{}
	txBonusPayments33 := []models.TxBonusPayment33{}
	txBonusPayments82 := []models.TxBonusPayment82{}
	txBonusPayments83 := []models.TxBonusPayment83{}
	txPrepayments34 := []models.TxPrepayment34{}
	txPrepayments84 := []models.TxPrepayment84{}
	txDocumentDiscounts35 := []models.TxDocumentDiscount35{}
	txDocumentDiscounts37 := []models.TxDocumentDiscount37{}
	txDocumentDiscounts85 := []models.TxDocumentDiscount85{}
	txDocumentDiscounts87 := []models.TxDocumentDiscount87{}
	txDocumentRoundings38 := []models.TxDocumentRounding38{}
	txNonFiscalPayments36 := []models.TxNonFiscalPayment36{}
	txNonFiscalPayments86 := []models.TxNonFiscalPayment86{}
	txFiscalPayments40 := []models.TxFiscalPayment40{}
	txFiscalPayments43 := []models.TxFiscalPayment43{}
	txDocumentOpens42 := []models.TxDocumentOpen42{}
	txDocumentCloseKKT45 := []models.TxDocumentCloseKKT45{}
	txDocumentCloseGp49 := []models.TxDocumentCloseGp49{}
	txDocumentCloses55 := []models.TxDocumentClose55{}
	txDocumentCancels56 := []models.TxDocumentCancel56{}
	txDocumentNonFinCloses58 := []models.TxDocumentNonFinClose58{}
	txDocumentClients65 := []models.TxDocumentClients65{}
	txDocumentEGAIS120 := []models.TxDocumentEGAIS120{}
	txVatKKT88 := []models.TxVATKKT88{}
	txCashIns50 := []models.TxCashIn50{}
	txCashOuts51 := []models.TxCashOut51{}
	txCounterChanges57 := []models.TxCounterChange57{}
	txReportZless60 := []models.TxReportZless60{}
	txReportZ63 := []models.TxReportZ63{}
	txShiftOpenDocs64 := []models.TxShiftOpenDoc64{}
	txShiftCloses61 := []models.TxShiftClose61{}
	txShiftOpens62 := []models.TxShiftOpen62{}
	txMarkUnits121 := []models.TxMarkUnit121{}

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse transaction line
		transaction, err := parseTransactionLine(line, sourceFolder)
		if err != nil {
			return nil, fmt.Errorf("error parsing line %d: %w", lineNumber, err)
		}

		parsed, ok := transaction.(ParsedTransaction)
		if !ok {
			return nil, fmt.Errorf("unknown transaction type: %T", transaction)
		}

		switch parsed.Table {
		case "tx_item_registration_1_11":
			if err := appendTx("tx_item_registration_1_11", parsed.Value, &txItemRegistrations); err != nil {
				return nil, err
			}
		case "tx_item_storno_2_12":
			if err := appendTx("tx_item_storno_2_12", parsed.Value, &txItemStorno); err != nil {
				return nil, err
			}
		case "tx_item_tax_4_14":
			if err := appendTx("tx_item_tax_4_14", parsed.Value, &txItemTax); err != nil {
				return nil, err
			}
		case "tx_item_kkt_6_16":
			if err := appendTx("tx_item_kkt_6_16", parsed.Value, &txItemKKT); err != nil {
				return nil, err
			}
		case "tx_special_price_3":
			if err := appendTx("tx_special_price_3", parsed.Value, &txSpecialPrices); err != nil {
				return nil, err
			}
		case "tx_bonus_accrual_9":
			if err := appendTx("tx_bonus_accrual_9", parsed.Value, &txBonusAccruals); err != nil {
				return nil, err
			}
		case "tx_bonus_refund_10":
			if err := appendTx("tx_bonus_refund_10", parsed.Value, &txBonusRefunds); err != nil {
				return nil, err
			}
		case "tx_position_discount_15":
			if err := appendTx("tx_position_discount_15", parsed.Value, &txPositionDiscounts15); err != nil {
				return nil, err
			}
		case "tx_position_discount_17":
			if err := appendTx("tx_position_discount_17", parsed.Value, &txPositionDiscounts17); err != nil {
				return nil, err
			}
		case "tx_bill_registration_21_23":
			if err := appendTx("tx_bill_registration_21_23", parsed.Value, &txBillRegistrations); err != nil {
				return nil, err
			}
		case "tx_bill_storno_22_24":
			if err := appendTx("tx_bill_storno_22_24", parsed.Value, &txBillStornos); err != nil {
				return nil, err
			}
		case "tx_employee_registration_25":
			if err := appendTx("tx_employee_registration_25", parsed.Value, &txEmployeeRegistrations); err != nil {
				return nil, err
			}
		case "tx_employee_accounting_doc_26":
			if err := appendTx("tx_employee_accounting_doc_26", parsed.Value, &txEmployeeAccountingDocs); err != nil {
				return nil, err
			}
		case "tx_employee_accounting_pos_29":
			if err := appendTx("tx_employee_accounting_pos_29", parsed.Value, &txEmployeeAccountingPos); err != nil {
				return nil, err
			}
		case "tx_card_status_change_27":
			if err := appendTx("tx_card_status_change_27", parsed.Value, &txCardStatusChanges); err != nil {
				return nil, err
			}
		case "tx_modifier_registration_30":
			if err := appendTx("tx_modifier_registration_30", parsed.Value, &txModifierRegistrations); err != nil {
				return nil, err
			}
		case "tx_modifier_storno_31":
			if err := appendTx("tx_modifier_storno_31", parsed.Value, &txModifierStornos); err != nil {
				return nil, err
			}
		case "tx_bonus_payment_32":
			if err := appendTx("tx_bonus_payment_32", parsed.Value, &txBonusPayments32); err != nil {
				return nil, err
			}
		case "tx_bonus_payment_33":
			if err := appendTx("tx_bonus_payment_33", parsed.Value, &txBonusPayments33); err != nil {
				return nil, err
			}
		case "tx_bonus_payment_82":
			if err := appendTx("tx_bonus_payment_82", parsed.Value, &txBonusPayments82); err != nil {
				return nil, err
			}
		case "tx_bonus_payment_83":
			if err := appendTx("tx_bonus_payment_83", parsed.Value, &txBonusPayments83); err != nil {
				return nil, err
			}
		case "tx_prepayment_34":
			if err := appendTx("tx_prepayment_34", parsed.Value, &txPrepayments34); err != nil {
				return nil, err
			}
		case "tx_prepayment_84":
			if err := appendTx("tx_prepayment_84", parsed.Value, &txPrepayments84); err != nil {
				return nil, err
			}
		case "tx_document_discount_35":
			if err := appendTx("tx_document_discount_35", parsed.Value, &txDocumentDiscounts35); err != nil {
				return nil, err
			}
		case "tx_document_discount_37":
			if err := appendTx("tx_document_discount_37", parsed.Value, &txDocumentDiscounts37); err != nil {
				return nil, err
			}
		case "tx_document_discount_85":
			if err := appendTx("tx_document_discount_85", parsed.Value, &txDocumentDiscounts85); err != nil {
				return nil, err
			}
		case "tx_document_discount_87":
			if err := appendTx("tx_document_discount_87", parsed.Value, &txDocumentDiscounts87); err != nil {
				return nil, err
			}
		case "tx_document_rounding_38":
			if err := appendTx("tx_document_rounding_38", parsed.Value, &txDocumentRoundings38); err != nil {
				return nil, err
			}
		case "tx_non_fiscal_payment_36":
			if err := appendTx("tx_non_fiscal_payment_36", parsed.Value, &txNonFiscalPayments36); err != nil {
				return nil, err
			}
		case "tx_non_fiscal_payment_86":
			if err := appendTx("tx_non_fiscal_payment_86", parsed.Value, &txNonFiscalPayments86); err != nil {
				return nil, err
			}
		case "tx_fiscal_payment_40":
			if err := appendTx("tx_fiscal_payment_40", parsed.Value, &txFiscalPayments40); err != nil {
				return nil, err
			}
		case "tx_fiscal_payment_43":
			if err := appendTx("tx_fiscal_payment_43", parsed.Value, &txFiscalPayments43); err != nil {
				return nil, err
			}
		case "tx_document_open_42":
			if err := appendTx("tx_document_open_42", parsed.Value, &txDocumentOpens42); err != nil {
				return nil, err
			}
		case "tx_document_close_kkt_45":
			if err := appendTx("tx_document_close_kkt_45", parsed.Value, &txDocumentCloseKKT45); err != nil {
				return nil, err
			}
		case "tx_document_close_gp_49":
			if err := appendTx("tx_document_close_gp_49", parsed.Value, &txDocumentCloseGp49); err != nil {
				return nil, err
			}
		case "tx_document_close_55":
			if err := appendTx("tx_document_close_55", parsed.Value, &txDocumentCloses55); err != nil {
				return nil, err
			}
		case "tx_document_cancel_56":
			if err := appendTx("tx_document_cancel_56", parsed.Value, &txDocumentCancels56); err != nil {
				return nil, err
			}
		case "tx_document_non_fin_close_58":
			if err := appendTx("tx_document_non_fin_close_58", parsed.Value, &txDocumentNonFinCloses58); err != nil {
				return nil, err
			}
		case "tx_document_clients_65":
			if err := appendTx("tx_document_clients_65", parsed.Value, &txDocumentClients65); err != nil {
				return nil, err
			}
		case "tx_document_egais_120":
			if err := appendTx("tx_document_egais_120", parsed.Value, &txDocumentEGAIS120); err != nil {
				return nil, err
			}
		case "tx_vat_kkt_88":
			if err := appendTx("tx_vat_kkt_88", parsed.Value, &txVatKKT88); err != nil {
				return nil, err
			}
		case "tx_cash_in_50":
			if err := appendTx("tx_cash_in_50", parsed.Value, &txCashIns50); err != nil {
				return nil, err
			}
		case "tx_cash_out_51":
			if err := appendTx("tx_cash_out_51", parsed.Value, &txCashOuts51); err != nil {
				return nil, err
			}
		case "tx_counter_change_57":
			if err := appendTx("tx_counter_change_57", parsed.Value, &txCounterChanges57); err != nil {
				return nil, err
			}
		case "tx_report_zless_60":
			if err := appendTx("tx_report_zless_60", parsed.Value, &txReportZless60); err != nil {
				return nil, err
			}
		case "tx_report_z_63":
			if err := appendTx("tx_report_z_63", parsed.Value, &txReportZ63); err != nil {
				return nil, err
			}
		case "tx_shift_open_doc_64":
			if err := appendTx("tx_shift_open_doc_64", parsed.Value, &txShiftOpenDocs64); err != nil {
				return nil, err
			}
		case "tx_shift_close_61":
			if err := appendTx("tx_shift_close_61", parsed.Value, &txShiftCloses61); err != nil {
				return nil, err
			}
		case "tx_shift_open_62":
			if err := appendTx("tx_shift_open_62", parsed.Value, &txShiftOpens62); err != nil {
				return nil, err
			}
		case "tx_mark_unit_121":
			if err := appendTx("tx_mark_unit_121", parsed.Value, &txMarkUnits121); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown transaction table: %s", parsed.Table)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Return grouped transactions (tx_* tables)
	result := make(map[string]interface{})
	if len(txItemRegistrations) > 0 {
		result["tx_item_registration_1_11"] = txItemRegistrations
	}
	if len(txItemStorno) > 0 {
		result["tx_item_storno_2_12"] = txItemStorno
	}
	if len(txItemTax) > 0 {
		result["tx_item_tax_4_14"] = txItemTax
	}
	if len(txItemKKT) > 0 {
		result["tx_item_kkt_6_16"] = txItemKKT
	}
	if len(txSpecialPrices) > 0 {
		result["tx_special_price_3"] = txSpecialPrices
	}
	if len(txBonusAccruals) > 0 {
		result["tx_bonus_accrual_9"] = txBonusAccruals
	}
	if len(txBonusRefunds) > 0 {
		result["tx_bonus_refund_10"] = txBonusRefunds
	}
	if len(txPositionDiscounts15) > 0 {
		result["tx_position_discount_15"] = txPositionDiscounts15
	}
	if len(txPositionDiscounts17) > 0 {
		result["tx_position_discount_17"] = txPositionDiscounts17
	}
	if len(txBillRegistrations) > 0 {
		result["tx_bill_registration_21_23"] = txBillRegistrations
	}
	if len(txBillStornos) > 0 {
		result["tx_bill_storno_22_24"] = txBillStornos
	}
	if len(txEmployeeRegistrations) > 0 {
		result["tx_employee_registration_25"] = txEmployeeRegistrations
	}
	if len(txEmployeeAccountingDocs) > 0 {
		result["tx_employee_accounting_doc_26"] = txEmployeeAccountingDocs
	}
	if len(txEmployeeAccountingPos) > 0 {
		result["tx_employee_accounting_pos_29"] = txEmployeeAccountingPos
	}
	if len(txCardStatusChanges) > 0 {
		result["tx_card_status_change_27"] = txCardStatusChanges
	}
	if len(txModifierRegistrations) > 0 {
		result["tx_modifier_registration_30"] = txModifierRegistrations
	}
	if len(txModifierStornos) > 0 {
		result["tx_modifier_storno_31"] = txModifierStornos
	}
	if len(txBonusPayments32) > 0 {
		result["tx_bonus_payment_32"] = txBonusPayments32
	}
	if len(txBonusPayments33) > 0 {
		result["tx_bonus_payment_33"] = txBonusPayments33
	}
	if len(txBonusPayments82) > 0 {
		result["tx_bonus_payment_82"] = txBonusPayments82
	}
	if len(txBonusPayments83) > 0 {
		result["tx_bonus_payment_83"] = txBonusPayments83
	}
	if len(txPrepayments34) > 0 {
		result["tx_prepayment_34"] = txPrepayments34
	}
	if len(txPrepayments84) > 0 {
		result["tx_prepayment_84"] = txPrepayments84
	}
	if len(txDocumentDiscounts35) > 0 {
		result["tx_document_discount_35"] = txDocumentDiscounts35
	}
	if len(txDocumentDiscounts37) > 0 {
		result["tx_document_discount_37"] = txDocumentDiscounts37
	}
	if len(txDocumentDiscounts85) > 0 {
		result["tx_document_discount_85"] = txDocumentDiscounts85
	}
	if len(txDocumentDiscounts87) > 0 {
		result["tx_document_discount_87"] = txDocumentDiscounts87
	}
	if len(txDocumentRoundings38) > 0 {
		result["tx_document_rounding_38"] = txDocumentRoundings38
	}
	if len(txNonFiscalPayments36) > 0 {
		result["tx_non_fiscal_payment_36"] = txNonFiscalPayments36
	}
	if len(txNonFiscalPayments86) > 0 {
		result["tx_non_fiscal_payment_86"] = txNonFiscalPayments86
	}
	if len(txFiscalPayments40) > 0 {
		result["tx_fiscal_payment_40"] = txFiscalPayments40
	}
	if len(txFiscalPayments43) > 0 {
		result["tx_fiscal_payment_43"] = txFiscalPayments43
	}
	if len(txDocumentOpens42) > 0 {
		result["tx_document_open_42"] = txDocumentOpens42
	}
	if len(txDocumentCloseKKT45) > 0 {
		result["tx_document_close_kkt_45"] = txDocumentCloseKKT45
	}
	if len(txDocumentCloseGp49) > 0 {
		result["tx_document_close_gp_49"] = txDocumentCloseGp49
	}
	if len(txDocumentCloses55) > 0 {
		result["tx_document_close_55"] = txDocumentCloses55
	}
	if len(txDocumentCancels56) > 0 {
		result["tx_document_cancel_56"] = txDocumentCancels56
	}
	if len(txDocumentNonFinCloses58) > 0 {
		result["tx_document_non_fin_close_58"] = txDocumentNonFinCloses58
	}
	if len(txDocumentClients65) > 0 {
		result["tx_document_clients_65"] = txDocumentClients65
	}
	if len(txDocumentEGAIS120) > 0 {
		result["tx_document_egais_120"] = txDocumentEGAIS120
	}
	if len(txVatKKT88) > 0 {
		result["tx_vat_kkt_88"] = txVatKKT88
	}
	if len(txCashIns50) > 0 {
		result["tx_cash_in_50"] = txCashIns50
	}
	if len(txCashOuts51) > 0 {
		result["tx_cash_out_51"] = txCashOuts51
	}
	if len(txCounterChanges57) > 0 {
		result["tx_counter_change_57"] = txCounterChanges57
	}
	if len(txReportZless60) > 0 {
		result["tx_report_zless_60"] = txReportZless60
	}
	if len(txReportZ63) > 0 {
		result["tx_report_z_63"] = txReportZ63
	}
	if len(txShiftOpenDocs64) > 0 {
		result["tx_shift_open_doc_64"] = txShiftOpenDocs64
	}
	if len(txShiftCloses61) > 0 {
		result["tx_shift_close_61"] = txShiftCloses61
	}
	if len(txShiftOpens62) > 0 {
		result["tx_shift_open_62"] = txShiftOpens62
	}
	if len(txMarkUnits121) > 0 {
		result["tx_mark_unit_121"] = txMarkUnits121
	}

	return result, nil
}

// parseTransactionLine parses a single transaction line
func parseTransactionLine(line string, sourceFolder string) (interface{}, error) {
	fields := strings.Split(line, ";")
	if len(fields) < 4 {
		return nil, fmt.Errorf("insufficient fields in line")
	}

	// Parse transaction type
	transactionType, err := strconv.Atoi(fields[3])
	if err != nil {
		return nil, fmt.Errorf("invalid transaction type: %s", fields[3])
	}

	// Use dispatcher to route to appropriate parser
	transactionTypeInfo, err := GetTransactionType(transactionType)
	if err != nil {
		return nil, fmt.Errorf("unhandled transaction type: %d", transactionType)
	}

	return transactionTypeInfo.Parser(fields, sourceFolder)
}

// parseBaseTransactionData parses common fields for all transaction types
func parseBaseTransactionData(fields []string, sourceFolder string) (models.BaseTransactionData, error) {
	base := models.BaseTransactionData{
		SourceFolder: sourceFolder,
	}

	// Parse required fields - need at least 4 fields for basic transaction data
	if len(fields) < 4 {
		return base, fmt.Errorf("insufficient fields for base transaction data")
	}

	// ID (field 1)
	base.ID = fields[0]

	// Date (field 2) - format: DD.MM.YYYY
	dateStr := fields[1]
	if dateStr != "" {
		date, err := time.Parse("02.01.2006", dateStr)
		if err != nil {
			// Use current date if parsing fails
			base.Date = time.Now()
		} else {
			base.Date = date
		}
	} else {
		// Use current date if field is empty
		base.Date = time.Now()
	}

	// Time (field 3) - format: HH:MM:SS
	timeStr := fields[2]
	if timeStr != "" {
		parsedTime, err := time.Parse("15:04:05", timeStr)
		if err != nil {
			// Use current time if parsing fails
			base.Time = time.Now()
		} else {
			base.Time = parsedTime
		}
	} else {
		// Use current time if field is empty
		base.Time = time.Now()
	}

	// Transaction Type (field 4)
	transactionType, err := strconv.Atoi(fields[3])
	if err != nil {
		return base, fmt.Errorf("invalid transaction type: %s", fields[3])
	}
	base.TransactionType = transactionType

	// Cash Register Code (field 5)
	if len(fields) > 4 {
		cashRegisterCode, err := strconv.ParseInt(fields[4], 10, 64)
		if err != nil {
			base.CashRegisterCode = 0
		} else {
			base.CashRegisterCode = cashRegisterCode
		}
	}

	// Document Number (field 6)
	if len(fields) > 5 {
		base.DocumentNumber = fields[5]
	}

	// Cashier Code (field 7)
	if len(fields) > 6 {
		base.CashierCode = fields[6]
	}

	// Shift Number (field 8)
	if len(fields) > 7 {
		if fields[7] != "" {
			shiftNumber, err := strconv.ParseInt(fields[7], 10, 64)
			if err != nil {
				base.ShiftNumber = 0
			} else {
				base.ShiftNumber = shiftNumber
			}
		} else {
			base.ShiftNumber = 0
		}
	}

	// Operation Type (field 9)
	if len(fields) > 8 {
		if fields[8] != "" {
			operationType, err := strconv.ParseInt(fields[8], 10, 64)
			if err != nil {
				base.OperationType = 0
			} else {
				base.OperationType = operationType
			}
		} else {
			base.OperationType = 0
		}
	}

	// Document Type Code (field 10)
	if len(fields) > 9 {
		if fields[9] != "" {
			documentTypeCode, err := strconv.ParseInt(fields[9], 10, 64)
			if err != nil {
				base.DocumentTypeCode = 0
			} else {
				base.DocumentTypeCode = documentTypeCode
			}
		} else {
			base.DocumentTypeCode = 0
		}
	}

	// Document Info (field 11)
	if len(fields) > 10 {
		base.DocumentInfo = fields[10]
	}

	// Enterprise ID (field 12)
	if len(fields) > 11 {
		if fields[11] != "" {
			enterpriseID, err := strconv.ParseInt(fields[11], 10, 64)
			if err != nil {
				base.EnterpriseID = 0
			} else {
				base.EnterpriseID = enterpriseID
			}
		} else {
			base.EnterpriseID = 0
		}
	}

	// Employee Code (field 13)
	if len(fields) > 12 {
		base.EmployeeCode = fields[12]
	}

	// Raw Data
	base.RawData = strings.Join(fields, ";")

	return base, nil
}
