//nolint:unused // Legacy mappers kept for reference while new tx parsing is in use.
package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/user/go-frontol-loader/pkg/models"
)

// parseFloatWithComma parses a float string that may use comma as decimal separator
func parseFloatWithComma(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("invalid float")
	}

	commaIndex := strings.Index(s, ",")
	if commaIndex == -1 {
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}
		return val, nil
	}
	if strings.Count(s, ",") > 1 {
		return 0, fmt.Errorf("invalid float")
	}

	afterComma := s[commaIndex+1:]
	if !strings.Contains(s, ".") && len(afterComma) == 3 {
		allDigits := true
		for i := 0; i < len(afterComma); i++ {
			if afterComma[i] < '0' || afterComma[i] > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			return 0, fmt.Errorf("invalid float")
		}
	}

	s = strings.Replace(s, ",", ".", 1)
	if strings.Count(s, ".") > 1 {
		parts := strings.SplitN(s, ".", 3)
		if len(parts) >= 2 {
			s = parts[0] + "." + parts[1]
		}
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// safeParseInt safely parses an integer string, returning 0 if parsing fails
func safeParseInt(s string) int {
	if s == "" {
		return 0
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

// safeParseInt64 safely parses an int64 string, returning 0 if parsing fails
func safeParseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// safeParseFloat safely parses a float string, returning 0.0 if parsing fails
func safeParseFloat(s string) float64 {
	if s == "" {
		return 0.0
	}
	val, err := parseFloatWithComma(s)
	if err != nil {
		return 0.0
	}
	return val
}

// parseTransactionRegistration parses transaction registration data
func parseTransactionRegistration(fields []string, sourceFolder string) (models.TransactionRegistration, error) {
	transaction := models.TransactionRegistration{
		SourceFolder: sourceFolder,
	}

	// Field 1: Transaction ID Unique
	if len(fields) > 0 {
		transaction.TransactionIDUnique = safeParseInt64(fields[0])
	}

	// Field 2: Transaction Date
	if len(fields) > 1 {
		// Parse date from DD.MM.YYYY format to YYYY-MM-DD
		if date, err := time.Parse("02.01.2006", fields[1]); err == nil {
			transaction.TransactionDate = date.Format("2006-01-02")
		} else {
			transaction.TransactionDate = fields[1] // Keep original if parsing fails
		}
	}

	// Field 3: Transaction Time
	if len(fields) > 2 {
		// Parse time from HH:MM:SS format
		if time, err := time.Parse("15:04:05", fields[2]); err == nil {
			transaction.TransactionTime = time.Format("15:04:05")
		} else {
			transaction.TransactionTime = fields[2] // Keep original if parsing fails
		}
	}

	// Field 4: Transaction Type
	if len(fields) > 3 {
		transaction.TransactionType = safeParseInt(fields[3])
	}

	// Field 5: Cash Register Code
	if len(fields) > 4 {
		transaction.CashRegisterCode = safeParseInt64(fields[4])
	}

	// Field 6: Document Number
	if len(fields) > 5 {
		transaction.DocumentNumber = safeParseInt64(fields[5])
	}

	// Field 7: Cashier Code
	if len(fields) > 6 {
		transaction.CashierCode = safeParseInt64(fields[6])
	}

	// Field 8: Item Code
	if len(fields) > 7 {
		transaction.ItemCode = fields[7]
	}

	// Field 9: Group Code
	if len(fields) > 8 {
		transaction.GroupCode = fields[8]
	}

	// Field 10: Amount Total
	if len(fields) > 9 {
		transaction.AmountTotal = safeParseFloat(fields[9])
	}

	// Field 11: Quantity
	if len(fields) > 10 {
		transaction.Quantity = safeParseFloat(fields[10])
	}

	// Field 12: Amount Cash Register
	if len(fields) > 11 {
		transaction.AmountCashRegister = safeParseFloat(fields[11])
	}

	// Field 13: Operation Type
	if len(fields) > 12 {
		transaction.OperationType = safeParseInt64(fields[12])
	}

	// Field 14: Shift Number
	if len(fields) > 13 {
		transaction.ShiftNumber = safeParseInt64(fields[13])
	}

	// Field 15: Item Price
	if len(fields) > 14 {
		transaction.ItemPrice = safeParseFloat(fields[14])
	}

	// Field 16: Item Sum
	if len(fields) > 15 {
		transaction.ItemSum = safeParseFloat(fields[15])
	}

	// Field 17: Print Group Code (Код группы печати) - according to Frontol 6 documentation
	if len(fields) > 16 {
		transaction.PrintGroupCode = safeParseInt64(fields[16])
	}

	// Field 18: Article SKU
	if len(fields) > 17 {
		transaction.ArticleSKU = fields[17]
	}

	// Field 19: Registration Barcode
	if len(fields) > 18 {
		transaction.RegistrationBarcode = safeParseFloat(fields[18])
	}

	// Field 20: Position Amount
	if len(fields) > 19 {
		transaction.PositionAmount = safeParseFloat(fields[19])
	}

	// Field 21: KKT Section
	if len(fields) > 20 {
		transaction.KKTSection = fields[20]
	}

	// Field 22: Reserved Field 22
	if len(fields) > 21 {
		transaction.ReservedField22 = safeParseInt64(fields[21])
	}

	// Field 23: Document Type Code
	if len(fields) > 22 {
		transaction.DocumentTypeCode = safeParseInt64(fields[22])
	}

	// Field 24: Comment Code
	if len(fields) > 23 {
		transaction.CommentCode = safeParseInt64(fields[23])
	}

	// Field 25: Reserved Field 25
	if len(fields) > 24 {
		transaction.ReservedField25 = fields[24]
	}

	// Field 26: Document Info
	if len(fields) > 25 {
		transaction.DocumentInfo = fields[25]
	}

	// Field 27: Enterprise ID
	if len(fields) > 26 {
		transaction.EnterpriseID = safeParseInt64(fields[26])
	}

	// Field 28: Employee Code
	if len(fields) > 27 {
		transaction.EmployeeCode = safeParseInt64(fields[27])
	}

	// Field 29: Divided Pack Qty
	if len(fields) > 28 {
		transaction.DividedPackQty = safeParseInt64(fields[28])
	}

	// Field 30: Gift Card Number
	if len(fields) > 29 {
		transaction.GiftCardNumber = fields[29]
	}

	// Field 31: Pack Quantity
	if len(fields) > 30 {
		transaction.PackQuantity = safeParseInt64(fields[30])
	}

	// Field 32: Nomenclature Type
	if len(fields) > 31 {
		transaction.NomenclatureType = safeParseInt(fields[31])
	}

	// Field 33: Marking Code
	if len(fields) > 32 {
		transaction.MarkingCode = fields[32]
	}

	// Field 34: Excise Stamp
	if len(fields) > 33 {
		transaction.ExciseStamp = fields[33]
	}

	// Field 35: Personal Mod Group
	if len(fields) > 34 {
		transaction.PersonalModGroup = fields[34]
	}

	// Field 36: Lottery Time
	if len(fields) > 35 {
		transaction.LotteryTime = fields[35]
	}

	// Field 37: Lottery ID
	if len(fields) > 36 {
		transaction.LotteryID = safeParseInt64(fields[36])
	}

	// Field 38: Reserved Field 38
	if len(fields) > 37 {
		transaction.ReservedField38 = safeParseInt64(fields[37])
	}

	// Field 39: ALC Code
	if len(fields) > 38 {
		transaction.ALCCode = fields[38]
	}

	// Field 40: Reserved Field 40
	if len(fields) > 39 {
		transaction.ReservedField40 = safeParseFloat(fields[39])
	}

	// Field 41: Prescription Data 1
	if len(fields) > 40 {
		transaction.PrescriptionData1 = fields[40]
	}

	// Field 42: Prescription Data 2
	if len(fields) > 41 {
		transaction.PrescriptionData2 = fields[41]
	}

	// Field 43: Coupons Per Item
	if len(fields) > 42 {
		transaction.CouponsPerItem = fields[42]
	}

	// Raw data for debugging
	transaction.RawData = strings.Join(fields, ";")

	return transaction, nil
}

// parseTransactionRegistrationOld parses transaction registration data (old version with detailed error handling)
func parseTransactionRegistrationOld(fields []string) (models.TransactionRegistration, error) {
	transaction := models.TransactionRegistration{}

	// Parse specific fields for transaction registration
	if len(fields) > 13 {
		// Item Code (field 14)
		transaction.ItemCode = fields[13]
	}

	if len(fields) > 14 {
		// Group Code (field 15)
		transaction.GroupCode = fields[14]
	}

	if len(fields) > 15 {
		// Amount Total (field 16)
		if fields[15] != "" {
			amountTotal, err := parseFloatWithComma(fields[15])
			if err != nil {
				return transaction, fmt.Errorf("invalid amount total: %s", fields[15])
			}
			transaction.AmountTotal = amountTotal
		}
	}

	if len(fields) > 16 {
		// Quantity (field 17)
		if fields[16] != "" {
			quantity, err := parseFloatWithComma(fields[16])
			if err != nil {
				return transaction, fmt.Errorf("invalid quantity: %s", fields[16])
			}
			transaction.Quantity = quantity
		}
	}

	if len(fields) > 17 {
		// Amount Cash Register (field 18)
		if fields[17] != "" {
			amountCashRegister, err := parseFloatWithComma(fields[17])
			if err != nil {
				return transaction, fmt.Errorf("invalid amount cash register: %s", fields[17])
			}
			transaction.AmountCashRegister = amountCashRegister
		}
	}

	if len(fields) > 18 {
		// Item Price (field 19)
		if fields[18] != "" {
			itemPrice, err := parseFloatWithComma(fields[18])
			if err != nil {
				return transaction, fmt.Errorf("invalid item price: %s", fields[18])
			}
			transaction.ItemPrice = itemPrice
		}
	}

	if len(fields) > 19 {
		// Item Sum (field 20)
		if fields[19] != "" {
			itemSum, err := parseFloatWithComma(fields[19])
			if err != nil {
				return transaction, fmt.Errorf("invalid item sum: %s", fields[19])
			}
			transaction.ItemSum = itemSum
		}
	}

	if len(fields) > 20 {
		// Item Line Number (field 21)
		if fields[20] != "" {
			itemLineNumber, err := parseFloatWithComma(fields[20])
			if err != nil {
				return transaction, fmt.Errorf("invalid item line number: %s", fields[20])
			}
			transaction.ItemLineNumber = int(itemLineNumber)
		}
	}

	// Print Group Code (field 22) - removed from new structure
	// if len(fields) > 21 { ... }

	if len(fields) > 22 {
		// Article SKU (field 23)
		transaction.ArticleSKU = fields[22]
	}

	if len(fields) > 23 {
		// Registration Barcode (field 24)
		if fields[23] != "" {
			registrationBarcode, err := parseFloatWithComma(fields[23])
			if err != nil {
				return transaction, fmt.Errorf("invalid registration barcode: %s", fields[23])
			}
			transaction.RegistrationBarcode = registrationBarcode
		}
	}

	if len(fields) > 24 {
		// Position Amount (field 25)
		if fields[24] != "" {
			positionAmount, err := parseFloatWithComma(fields[24])
			if err != nil {
				return transaction, fmt.Errorf("invalid position amount: %s", fields[24])
			}
			transaction.PositionAmount = positionAmount
		}
	}

	if len(fields) > 25 {
		// KKT Section (field 26)
		transaction.KKTSection = fields[25]
	}

	if len(fields) > 26 {
		// Reserved Field 22 (field 27)
		if fields[26] != "" {
			reservedField22, err := strconv.ParseInt(fields[26], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid reserved field 22: %s", fields[26])
			}
			transaction.ReservedField22 = reservedField22
		}
	}

	if len(fields) > 27 {
		// Comment Code (field 28)
		if fields[27] != "" {
			commentCode, err := strconv.ParseInt(fields[27], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid comment code: %s", fields[27])
			}
			transaction.CommentCode = commentCode
		}
	}

	if len(fields) > 28 {
		// Reserved Field 25 (field 29)
		transaction.ReservedField25 = fields[28]
	}

	if len(fields) > 29 {
		// Divided Pack Qty (field 30)
		if fields[29] != "" {
			dividedPackQty, err := strconv.ParseInt(fields[29], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid divided pack qty: %s", fields[29])
			}
			transaction.DividedPackQty = dividedPackQty
		}
	}

	if len(fields) > 30 {
		// Gift Card Number (field 31)
		transaction.GiftCardNumber = fields[30]
	}

	if len(fields) > 31 {
		// Pack Quantity (field 32)
		if fields[31] != "" {
			packQuantity, err := strconv.ParseInt(fields[31], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid pack quantity: %s", fields[31])
			}
			transaction.PackQuantity = packQuantity
		}
	}

	if len(fields) > 32 {
		// Nomenclature Type (field 33)
		if fields[32] != "" {
			nomenclatureType, err := strconv.Atoi(fields[32])
			if err != nil {
				// Skip invalid nomenclature type, use default value 0
				transaction.NomenclatureType = 0
			} else {
				transaction.NomenclatureType = nomenclatureType
			}
		}
	}

	if len(fields) > 33 {
		// Marking Code (field 34)
		transaction.MarkingCode = fields[33]
	}

	if len(fields) > 34 {
		// Excise Stamp (field 35)
		transaction.ExciseStamp = fields[34]
	}

	if len(fields) > 35 {
		// Personal Mod Group (field 36)
		transaction.PersonalModGroup = fields[35]
	}

	if len(fields) > 36 {
		// Lottery Time (field 37)
		transaction.LotteryTime = fields[36]
	}

	if len(fields) > 37 {
		// Lottery ID (field 38)
		if fields[37] != "" {
			lotteryID, err := strconv.ParseInt(fields[37], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid lottery ID: %s", fields[37])
			}
			transaction.LotteryID = lotteryID
		}
	}

	if len(fields) > 38 {
		// Reserved Field 38 (field 39)
		if fields[38] != "" {
			reservedField38, err := strconv.ParseInt(fields[38], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid reserved field 38: %s", fields[38])
			}
			transaction.ReservedField38 = reservedField38
		}
	}

	if len(fields) > 39 {
		// ALC Code (field 40)
		transaction.ALCCode = fields[39]
	}

	if len(fields) > 40 {
		// Reserved Field 40 (field 41)
		transaction.ReservedField40 = safeParseFloat(fields[40])
	}

	if len(fields) > 41 {
		// Prescription Data 1 (field 42)
		transaction.PrescriptionData1 = fields[41]
	}

	if len(fields) > 42 {
		// Prescription Data 2 (field 43)
		transaction.PrescriptionData2 = fields[42]
	}

	if len(fields) > 43 {
		// Coupons Per Item (field 44)
		transaction.CouponsPerItem = fields[43]
	}

	return transaction, nil
}

// parseSpecialPrice parses special price data
// parseSpecialPrice parses special price transaction data
// Тип транзакции: 3 (Установка спеццены/цены из прайс-листа)
// Документация: frontol_6_integration.md, стр. 934-983
func parseSpecialPrice(fields []string, sourceFolder string) (models.SpecialPrice, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.SpecialPrice{}, err
	}

	price := models.SpecialPrice{
		BaseTransactionData: base,
	}

	// Поле 8: Код прайс-листа - fields[7]
	if len(fields) > 7 {
		price.PriceListCode = fields[7]
	}

	// Поле 9: пустое - fields[8] (не парсится)

	// Поле 10: Тип цены (0=спеццена, 1=из прайс-листа) - fields[9]
	if len(fields) > 9 {
		if fields[9] != "" {
			priceType, err := strconv.Atoi(fields[9])
			if err != nil {
				return price, fmt.Errorf("invalid price type: %s", fields[9])
			}
			price.PriceType = priceType
		}
	}

	// Поле 11: Спеццена или цена из прайс-листа - fields[10]
	if len(fields) > 10 {
		if fields[10] != "" {
			specialPrice, err := parseFloatWithComma(fields[10])
			if err != nil {
				return price, fmt.Errorf("invalid special price: %s", fields[10])
			}
			price.SpecialPrice = specialPrice
		}
	}

	// Поле 12: Цена из карточки товара - fields[11]
	if len(fields) > 11 {
		if fields[11] != "" {
			productCardPrice, err := parseFloatWithComma(fields[11])
			if err != nil {
				return price, fmt.Errorf("invalid product card price: %s", fields[11])
			}
			price.ProductCardPrice = productCardPrice
		}
	}

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				price.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				price.ShiftNumber = shiftNumber
			}
		}
	}

	// Поле 15: Код акции - fields[14]
	if len(fields) > 14 {
		if fields[14] != "" {
			promotionCode, err := strconv.ParseInt(fields[14], 10, 64)
			if err != nil {
				return price, fmt.Errorf("invalid promotion code: %s", fields[14])
			}
			price.PromotionCode = promotionCode
		}
	}

	// Поле 16: Код мероприятия - fields[15]
	if len(fields) > 15 {
		if fields[15] != "" {
			eventCode, err := strconv.ParseInt(fields[15], 10, 64)
			if err != nil {
				return price, fmt.Errorf("invalid event code: %s", fields[15])
			}
			price.EventCode = eventCode
		}
	}

	// Поле 17: Код группы печати - fields[16]
	if len(fields) > 16 {
		if fields[16] != "" {
			printGroupCode, err := strconv.ParseInt(fields[16], 10, 64)
			if err != nil {
				return price, fmt.Errorf("invalid print group code: %s", fields[16])
			}
			price.PrintGroupCode = printGroupCode
		}
	}

	return price, nil
}

// parseBonusTransaction parses bonus transaction data
// Типы транзакций: 9 (Начисление бонуса), 10 (Возврат бонуса)
// Документация: frontol_6_integration.md, стр. 985-1037
func parseBonusTransaction(fields []string, sourceFolder string) (models.BonusTransaction, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.BonusTransaction{}, err
	}

	transaction := models.BonusTransaction{
		BaseTransactionData: base,
	}

	// Поле 10: Тип бонуса (0=внутренний, 1=внешний) - fields[9]
	if len(fields) > 9 {
		if fields[9] != "" {
			bonusType, err := strconv.Atoi(fields[9])
			if err != nil {
				return transaction, fmt.Errorf("invalid bonus type: %s", fields[9])
			}
			transaction.BonusType = bonusType
		}
	}

	// Поле 12: Начисленная сумма бонуса - fields[11]
	if len(fields) > 11 {
		if fields[11] != "" {
			accruedAmount, err := parseFloatWithComma(fields[11])
			if err != nil {
				return transaction, fmt.Errorf("invalid accrued bonus amount: %s", fields[11])
			}
			transaction.AccruedBonusAmount = accruedAmount
			transaction.BonusAmount = accruedAmount // Для обратной совместимости
		}
	}

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	// parseBaseTransactionData парсит operation_type из fields[8], но для бонусов
	// поле 8 пустое, а operation_type находится в поле 13
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				transaction.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	// parseBaseTransactionData парсит shift_number из fields[7], но для бонусов
	// поле 7 пустое, а shift_number находится в поле 14
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				transaction.ShiftNumber = shiftNumber
			}
		}
	}

	// Поле 15: Код акции - fields[14]
	if len(fields) > 14 {
		if fields[14] != "" {
			promotionCode, err := strconv.ParseInt(fields[14], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid promotion code: %s", fields[14])
			}
			transaction.PromotionCode = promotionCode
		}
	}

	// Поле 16: Код мероприятия - fields[15]
	if len(fields) > 15 {
		if fields[15] != "" {
			eventCode, err := strconv.ParseInt(fields[15], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid event code: %s", fields[15])
			}
			transaction.EventCode = eventCode
		}
	}

	// Поле 17: Код группы печати - fields[16]
	if len(fields) > 16 {
		if fields[16] != "" {
			printGroupCode, err := strconv.ParseInt(fields[16], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid print group code: %s", fields[16])
			}
			transaction.PrintGroupCode = printGroupCode
		}
	}

	// Поле 29: Номер протокола ПС (для внешних бонусов) - fields[28]
	if len(fields) > 28 {
		if fields[28] != "" {
			psProtocolNumber, err := strconv.ParseInt(fields[28], 10, 64)
			if err == nil {
				transaction.PSProtocolNumber = psProtocolNumber
			}
		}
	}

	return transaction, nil
}

// parseDiscountTransaction parses discount transaction data
// По документации Frontol 6 для скидок на позицию (типы 15, 17):
// Поле 8 = discount_info (fields[7])
// Поле 9 = пустое (fields[8])
// Поле 10 = discount_type (fields[9])
// Поле 11 = discount_value (fields[10]) - значение скидки (процент для типа 17, сумма для типа 15)
// Поле 12 = discount_amount (fields[11]) - сумма скидки в базовой валюте
// Поле 13 = operation_type (fields[12]) - уже парсится в base
// Поле 14 = shift_number (fields[13]) - уже парсится в base
// Поле 15 = promotion_code (fields[14])
// Поле 16 = event_code (fields[15])
// Поле 17 = print_group_code (fields[16])
func parseDiscountTransaction(fields []string, sourceFolder string) (models.DiscountTransaction, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.DiscountTransaction{}, err
	}

	transaction := models.DiscountTransaction{
		BaseTransactionData: base,
	}

	// Поле 10: Тип скидки (fields[9])
	if len(fields) > 9 {
		transaction.DiscountType = safeParseInt(fields[9])
	}

	// Поле 11: Значение скидки (fields[10])
	// Для типа 17 (скидка %) это процент скидки
	// Для типа 15 (скидка суммой) это сумма скидки
	if len(fields) > 10 {
		transaction.DiscountValue = safeParseFloat(fields[10])
		// Для типа 17 (скидка %) discount_value = процент скидки
		if base.TransactionType == 17 {
			transaction.DiscountPercent = safeParseFloat(fields[10])
		}
	}

	// Поле 12: Сумма скидки в базовой валюте (fields[11])
	if len(fields) > 11 {
		transaction.DiscountAmount = safeParseFloat(fields[11])
	}

	// Поле 15: Код акции (fields[14])
	if len(fields) > 14 {
		transaction.PromotionCode = safeParseInt64(fields[14])
	}

	// Поле 16: Код мероприятия (fields[15])
	if len(fields) > 15 {
		transaction.EventCode = safeParseInt64(fields[15])
	}

	// Поле 17: Код группы печати (fields[16])
	if len(fields) > 16 {
		transaction.PrintGroupCode = safeParseInt64(fields[16])
	}

	return transaction, nil
}

// parseBillRegistration parses bill registration data
func parseBillRegistration(fields []string, sourceFolder string) (models.BillRegistration, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.BillRegistration{}, err
	}

	bill := models.BillRegistration{
		BaseTransactionData: base,
	}

	// Parse specific fields for bill registration according to database schema
	// Field 8: Bill Code
	if len(fields) > 7 {
		bill.BillCode = fields[7]
	}

	// Field 9: Group Code
	if len(fields) > 8 {
		bill.GroupCode = fields[8]
	}

	// Field 10: Bill Denomination
	if len(fields) > 9 {
		if fields[9] != "" {
			billDenomination, err := parseFloatWithComma(fields[9])
			if err != nil {
				return bill, fmt.Errorf("invalid bill denomination: %s", fields[9])
			}
			bill.BillDenomination = billDenomination
		}
	}

	// Field 11: Bill Quantity
	if len(fields) > 10 {
		if fields[10] != "" {
			billQuantity, err := parseFloatWithComma(fields[10])
			if err != nil {
				return bill, fmt.Errorf("invalid bill quantity: %s", fields[10])
			}
			bill.BillQuantity = billQuantity
		}
	}

	// Field 12: Bill Amount
	if len(fields) > 11 {
		if fields[11] != "" {
			billAmount, err := parseFloatWithComma(fields[11])
			if err != nil {
				return bill, fmt.Errorf("invalid bill amount: %s", fields[11])
			}
			bill.BillAmount = billAmount
		}
	}

	// Field 15: Bill Number
	if len(fields) > 14 {
		bill.BillNumber = fields[14]
	}

	// Field 16: Bill Total Amount
	if len(fields) > 15 {
		if fields[15] != "" {
			billTotalAmount, err := parseFloatWithComma(fields[15])
			if err != nil {
				return bill, fmt.Errorf("invalid bill total amount: %s", fields[15])
			}
			bill.BillTotalAmount = billTotalAmount
		}
	}

	// Field 17: Bill Type
	if len(fields) > 16 {
		if fields[16] != "" {
			billType, err := strconv.ParseInt(fields[16], 10, 64)
			if err != nil {
				return bill, fmt.Errorf("invalid bill type: %s", fields[16])
			}
			bill.BillType = billType
		}
	}

	// Field 18: Print Group Code
	if len(fields) > 17 {
		if fields[17] != "" {
			printGroupCode, err := strconv.ParseInt(fields[17], 10, 64)
			if err != nil {
				return bill, fmt.Errorf("invalid print group code: %s", fields[17])
			}
			bill.PrintGroupCode = printGroupCode
		}
	}

	// Field 19: Customer Code
	if len(fields) > 18 {
		bill.CustomerCode = fields[18]
	}

	// Field 20: Reserved Field 20
	if len(fields) > 19 {
		bill.ReservedField20 = fields[19]
	}

	// Field 21: Reserved Field 21
	if len(fields) > 20 {
		if fields[20] != "" {
			reservedField21, err := strconv.ParseInt(fields[20], 10, 64)
			if err != nil {
				return bill, fmt.Errorf("invalid reserved field 21: %s", fields[20])
			}
			bill.ReservedField21 = reservedField21
		}
	}

	// Field 22: Reserved Field 22
	if len(fields) > 21 {
		if fields[21] != "" {
			reservedField22, err := strconv.ParseInt(fields[21], 10, 64)
			if err != nil {
				return bill, fmt.Errorf("invalid reserved field 22: %s", fields[21])
			}
			bill.ReservedField22 = reservedField22
		}
	}

	// Field 24: Reserved Field 24
	if len(fields) > 23 {
		if fields[23] != "" {
			reservedField24, err := strconv.ParseInt(fields[23], 10, 64)
			if err != nil {
				return bill, fmt.Errorf("invalid reserved field 24: %s", fields[23])
			}
			bill.ReservedField24 = reservedField24
		}
	}

	// Field 25: Reserved Field 25
	if len(fields) > 24 {
		if fields[24] != "" {
			reservedField25, err := strconv.ParseInt(fields[24], 10, 64)
			if err != nil {
				return bill, fmt.Errorf("invalid reserved field 25: %s", fields[24])
			}
			bill.ReservedField25 = reservedField25
		}
	}

	// Raw data for debugging
	bill.RawData = strings.Join(fields, ";")

	return bill, nil
}

// parseEmployeeEdit parses employee edit data
// Тип транзакции: 25 (Регистрация сотрудников в документе редактирования списка сотрудников)
// Документация: frontol_6_integration.md, стр. 1141-1181
func parseEmployeeEdit(fields []string, sourceFolder string) (models.EmployeeEdit, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.EmployeeEdit{}, err
	}

	edit := models.EmployeeEdit{
		BaseTransactionData: base,
	}

	// Поле 8: Код сотрудника - fields[7]
	if len(fields) > 7 {
		edit.EmployeeCode = fields[7]
	}

	// Поля 9-12: пустые - не парсятся

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				edit.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				edit.ShiftNumber = shiftNumber
			}
		}
	}

	// Поля 15-17: пустые - не парсятся

	return edit, nil
}

// parseEmployeeAccounting parses employee accounting data
// parseEmployeeAccounting parses employee accounting data
// Типы транзакций: 26 (Учет сотрудников по документу), 29 (Учет сотрудников по позиции)
// Документация: frontol_6_integration.md, стр. 1184-1224
func parseEmployeeAccounting(fields []string, sourceFolder string) (models.EmployeeAccounting, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.EmployeeAccounting{}, err
	}

	accounting := models.EmployeeAccounting{
		BaseTransactionData: base,
	}

	// Поле 8: Код сотрудника - fields[7]
	if len(fields) > 7 {
		accounting.EmployeeCode = fields[7]
	}

	// Поля 9-12: пустые - не парсятся

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				accounting.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				accounting.ShiftNumber = shiftNumber
			}
		}
	}

	// Поля 15-16: пустые - не парсятся

	// Поле 17: Код группы печати документа - fields[16]
	if len(fields) > 16 {
		if fields[16] != "" {
			printGroupCode, err := strconv.ParseInt(fields[16], 10, 64)
			if err == nil {
				accounting.PrintGroupCode = printGroupCode
			}
		}
	}

	return accounting, nil
}

// parseVatKKTTransaction parses VAT KKT transaction data
func parseVatKKTTransaction(fields []string, sourceFolder string) (models.VatKKTTransaction, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.VatKKTTransaction{}, err
	}

	transaction := models.VatKKTTransaction{
		BaseTransactionData: base,
	}

	// Parse specific fields for VAT KKT transaction according to database schema
	// Field 8: VAT 0% amount
	if len(fields) > 7 {
		if fields[7] != "" {
			vat0Amount, err := parseFloatWithComma(fields[7])
			if err != nil {
				return transaction, fmt.Errorf("invalid VAT 0%% amount: %s", fields[7])
			}
			transaction.Vat0Amount = vat0Amount
		}
	}

	// Field 9: VAT 10% amount
	if len(fields) > 8 {
		if fields[8] != "" {
			vat10Amount, err := parseFloatWithComma(fields[8])
			if err != nil {
				return transaction, fmt.Errorf("invalid VAT 10%% amount: %s", fields[8])
			}
			transaction.Vat10Amount = vat10Amount
		}
	}

	// Field 10: VAT 20% amount
	if len(fields) > 9 {
		if fields[9] != "" {
			vat20Amount, err := parseFloatWithComma(fields[9])
			if err != nil {
				return transaction, fmt.Errorf("invalid VAT 20%% amount: %s", fields[9])
			}
			transaction.Vat20Amount = vat20Amount
		}
	}

	// Field 11: Amount without VAT
	if len(fields) > 10 {
		if fields[10] != "" {
			amountWithoutVat, err := parseFloatWithComma(fields[10])
			if err != nil {
				return transaction, fmt.Errorf("invalid amount without VAT: %s", fields[10])
			}
			transaction.AmountWithoutVat = amountWithoutVat
		}
	}

	// Field 12: VAT 10/110 amount
	if len(fields) > 11 {
		if fields[11] != "" {
			vat10_110Amount, err := parseFloatWithComma(fields[11])
			if err != nil {
				return transaction, fmt.Errorf("invalid VAT 10/110 amount: %s", fields[11])
			}
			transaction.Vat10_110Amount = vat10_110Amount
		}
	}

	// Field 13: VAT 20/120 amount
	if len(fields) > 12 {
		if fields[12] != "" {
			vat20_120Amount, err := parseFloatWithComma(fields[12])
			if err != nil {
				return transaction, fmt.Errorf("invalid VAT 20/120 amount: %s", fields[12])
			}
			transaction.Vat20_120Amount = vat20_120Amount
		}
	}

	return transaction, nil
}

// parseAdditionalTransaction parses additional transaction data
func parseAdditionalTransaction(fields []string, sourceFolder string) (models.AdditionalTransaction, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.AdditionalTransaction{}, err
	}

	transaction := models.AdditionalTransaction{
		BaseTransactionData: base,
	}

	// Parse specific fields for additional transaction
	if len(fields) > 13 {
		if fields[13] != "" {
			additionalType, err := strconv.Atoi(fields[13])
			if err != nil {
				return transaction, fmt.Errorf("invalid additional type: %s", fields[13])
			}
			transaction.AdditionalType = additionalType
		}
	}

	if len(fields) > 14 {
		if fields[14] != "" {
			additionalAmount, err := parseFloatWithComma(fields[14])
			if err != nil {
				return transaction, fmt.Errorf("invalid additional amount: %s", fields[14])
			}
			transaction.AdditionalAmount = additionalAmount
		}
	}

	if len(fields) > 15 {
		transaction.AdditionalInfo = fields[15]
	}

	if len(fields) > 16 {
		if fields[16] != "" {
			printGroupCode, err := strconv.ParseInt(fields[16], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid print group code: %s", fields[16])
			}
			transaction.PrintGroupCode = printGroupCode
		}
	}

	return transaction, nil
}

// parseAstuExchangeTransaction parses ASTU exchange transaction data
func parseAstuExchangeTransaction(fields []string, sourceFolder string) (models.AstuExchangeTransaction, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.AstuExchangeTransaction{}, err
	}

	transaction := models.AstuExchangeTransaction{
		BaseTransactionData: base,
	}

	// Parse specific fields for ASTU exchange transaction
	if len(fields) > 13 {
		if fields[13] != "" {
			exchangeType, err := strconv.Atoi(fields[13])
			if err != nil {
				return transaction, fmt.Errorf("invalid exchange type: %s", fields[13])
			}
			transaction.ExchangeType = exchangeType
		}
	}

	if len(fields) > 14 {
		if fields[14] != "" {
			exchangeAmount, err := parseFloatWithComma(fields[14])
			if err != nil {
				return transaction, fmt.Errorf("invalid exchange amount: %s", fields[14])
			}
			transaction.ExchangeAmount = exchangeAmount
		}
	}

	if len(fields) > 15 {
		if fields[15] != "" {
			exchangeRate, err := parseFloatWithComma(fields[15])
			if err != nil {
				return transaction, fmt.Errorf("invalid exchange rate: %s", fields[15])
			}
			transaction.ExchangeRate = exchangeRate
		}
	}

	if len(fields) > 16 {
		if fields[16] != "" {
			printGroupCode, err := strconv.ParseInt(fields[16], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid print group code: %s", fields[16])
			}
			transaction.PrintGroupCode = printGroupCode
		}
	}

	return transaction, nil
}

// parseCounterChangeTransaction parses counter change transaction data
// Тип транзакции: 57 (Изменение счетчика)
// Документация: frontol_6_integration.md, стр. 619-666
func parseCounterChangeTransaction(fields []string, sourceFolder string) (models.CounterChangeTransaction, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.CounterChangeTransaction{}, err
	}

	transaction := models.CounterChangeTransaction{
		BaseTransactionData: base,
	}

	// Поле 8: Номер карты/Код клиента - fields[7]
	if len(fields) > 7 {
		transaction.CardNumberOrClientCode = fields[7]
	}

	// Поле 9: Код вида карты - fields[8]
	if len(fields) > 8 {
		transaction.CardTypeCode = fields[8]
	}

	// Поле 10: Привязка (1=глобальный, 2=клиент, 3=карта, 4=подарочная карта) - fields[9]
	if len(fields) > 9 {
		if fields[9] != "" {
			bindingType, err := parseFloatWithComma(fields[9])
			if err != nil {
				return transaction, fmt.Errorf("invalid binding type: %s", fields[9])
			}
			transaction.BindingType = bindingType
		}
	}

	// Поле 11: Значение после изменений - fields[10]
	if len(fields) > 10 {
		if fields[10] != "" {
			valueAfterChanges, err := parseFloatWithComma(fields[10])
			if err != nil {
				return transaction, fmt.Errorf("invalid value after changes: %s", fields[10])
			}
			transaction.ValueAfterChanges = valueAfterChanges
		}
	}

	// Поле 12: Сумма изменения счетчика - fields[11]
	if len(fields) > 11 {
		if fields[11] != "" {
			changeAmount, err := parseFloatWithComma(fields[11])
			if err != nil {
				return transaction, fmt.Errorf("invalid change amount: %s", fields[11])
			}
			transaction.ChangeAmount = changeAmount
		}
	}

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				transaction.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				transaction.ShiftNumber = shiftNumber
			}
		}
	}

	// Поле 15: Код акции - fields[14]
	if len(fields) > 14 {
		if fields[14] != "" {
			promotionCode, err := strconv.ParseInt(fields[14], 10, 64)
			if err == nil {
				transaction.PromotionCode = promotionCode
			}
		}
	}

	// Поле 16: Код мероприятия - fields[15]
	if len(fields) > 15 {
		if fields[15] != "" {
			eventCode, err := strconv.ParseInt(fields[15], 10, 64)
			if err == nil {
				transaction.EventCode = eventCode
			}
		}
	}

	// Поле 17: пустое - fields[16] (не парсится)

	// Поле 21: Код вида счетчика - fields[20]
	if len(fields) > 20 {
		if fields[20] != "" {
			counterTypeCode, err := strconv.ParseInt(fields[20], 10, 64)
			if err == nil {
				transaction.CounterTypeCode = counterTypeCode
			}
		}
	}

	// Поле 22: Код счетчика - fields[21]
	if len(fields) > 21 {
		if fields[21] != "" {
			counterCode, err := strconv.ParseInt(fields[21], 10, 64)
			if err == nil {
				transaction.CounterCode = counterCode
			}
		}
	}

	// Поле 30: Дата начала действия движения счетчика - fields[29]
	if len(fields) > 29 {
		transaction.MovementStartDate = fields[29]
	}

	// Поле 33: Дата начала действия карты - fields[32]
	if len(fields) > 32 {
		transaction.CardStartDate = fields[32]
	}

	// Поле 34: Дата окончания действия карты - fields[33]
	if len(fields) > 33 {
		transaction.CardEndDate = fields[33]
	}

	// Поле 35: Дата окончания действия движения счетчика - fields[34]
	if len(fields) > 34 {
		transaction.MovementEndDate = fields[34]
	}

	return transaction, nil
}

// parseKKTShiftReport parses KKT shift report data
func parseKKTShiftReport(fields []string, sourceFolder string) (models.KKTShiftReport, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.KKTShiftReport{}, err
	}

	report := models.KKTShiftReport{
		BaseTransactionData: base,
	}

	// Parse specific fields for KKT shift report
	if len(fields) > 13 {
		if fields[13] != "" {
			reportType, err := strconv.Atoi(fields[13])
			if err != nil {
				return report, fmt.Errorf("invalid report type: %s", fields[13])
			}
			report.ReportType = reportType
		}
	}

	if len(fields) > 14 {
		report.ReportData = fields[14]
	}

	if len(fields) > 15 {
		if fields[15] != "" {
			reportAmount, err := parseFloatWithComma(fields[15])
			if err != nil {
				return report, fmt.Errorf("invalid report amount: %s", fields[15])
			}
			report.ReportAmount = reportAmount
		}
	}

	if len(fields) > 16 {
		if fields[16] != "" {
			printGroupCode, err := strconv.ParseInt(fields[16], 10, 64)
			if err != nil {
				return report, fmt.Errorf("invalid print group code: %s", fields[16])
			}
			report.PrintGroupCode = printGroupCode
		}
	}

	return report, nil
}

// parseFrontolMarkUnitTransaction parses Frontol mark unit transaction data
func parseFrontolMarkUnitTransaction(fields []string, sourceFolder string) (models.FrontolMarkUnitTransaction, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.FrontolMarkUnitTransaction{}, err
	}

	transaction := models.FrontolMarkUnitTransaction{
		BaseTransactionData: base,
	}

	// Parse specific fields for Frontol mark unit transaction
	if len(fields) > 13 {
		if fields[13] != "" {
			markUnitType, err := strconv.Atoi(fields[13])
			if err != nil {
				return transaction, fmt.Errorf("invalid mark unit type: %s", fields[13])
			}
			transaction.MarkUnitType = markUnitType
		}
	}

	if len(fields) > 14 {
		transaction.MarkUnitCode = fields[14]
	}

	if len(fields) > 15 {
		transaction.MarkUnitData = fields[15]
	}

	if len(fields) > 16 {
		if fields[16] != "" {
			printGroupCode, err := strconv.ParseInt(fields[16], 10, 64)
			if err != nil {
				return transaction, fmt.Errorf("invalid print group code: %s", fields[16])
			}
			transaction.PrintGroupCode = printGroupCode
		}
	}

	return transaction, nil
}

// parseBonusPayment parses bonus payment data
// Типы транзакций: 32 (Оплата бонусом), 33 (Возврат оплаты бонусом), 82 (Распределение оплаты), 83 (Распределение возврата)
// Документация: frontol_6_integration.md, стр. 1332-1381
func parseBonusPayment(fields []string, sourceFolder string) (models.BonusPayment, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.BonusPayment{}, err
	}

	payment := models.BonusPayment{
		BaseTransactionData: base,
	}

	// Поле 8: Номер бонусной карты - fields[7]
	if len(fields) > 7 {
		payment.BonusCardNumber = fields[7]
		payment.CardNumber = fields[7] // Для обратной совместимости
	}

	// Поле 9: пустое - fields[8] (не парсится)

	// Поле 10: Тип оплаты бонусом (0=внутренний, 1=внешний) - fields[9]
	if len(fields) > 9 {
		if fields[9] != "" {
			paymentType, err := strconv.Atoi(fields[9])
			if err != nil {
				return payment, fmt.Errorf("invalid payment type: %s", fields[9])
			}
			payment.PaymentType = paymentType
		}
	}

	// Поле 11: Величина изменения счетчика - fields[10]
	if len(fields) > 10 {
		if fields[10] != "" {
			counterChangeValue, err := parseFloatWithComma(fields[10])
			if err != nil {
				return payment, fmt.Errorf("invalid counter change value: %s", fields[10])
			}
			payment.CounterChangeValue = counterChangeValue
		}
	}

	// Поле 12: Сумма оплаты - fields[11]
	if len(fields) > 11 {
		if fields[11] != "" {
			paymentAmount, err := parseFloatWithComma(fields[11])
			if err != nil {
				return payment, fmt.Errorf("invalid payment amount: %s", fields[11])
			}
			payment.PaymentAmount = paymentAmount
		}
	}

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				payment.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				payment.ShiftNumber = shiftNumber
			}
		}
	}

	// Поле 15: Код акции - fields[14]
	if len(fields) > 14 {
		if fields[14] != "" {
			promotionCode, err := strconv.ParseInt(fields[14], 10, 64)
			if err != nil {
				return payment, fmt.Errorf("invalid promotion code: %s", fields[14])
			}
			payment.PromotionCode = promotionCode
		}
	}

	// Поле 16: Код мероприятия - fields[15]
	if len(fields) > 15 {
		if fields[15] != "" {
			eventCode, err := strconv.ParseInt(fields[15], 10, 64)
			if err != nil {
				return payment, fmt.Errorf("invalid event code: %s", fields[15])
			}
			payment.EventCode = eventCode
		}
	}

	// Поле 17: Код группы печати - fields[16]
	if len(fields) > 16 {
		if fields[16] != "" {
			printGroupCode, err := strconv.ParseInt(fields[16], 10, 64)
			if err != nil {
				return payment, fmt.Errorf("invalid print group code: %s", fields[16])
			}
			payment.PrintGroupCode = printGroupCode
		}
	}

	// Поле 29: Номер протокола ПС (для внешних бонусов) - fields[28]
	if len(fields) > 28 {
		if fields[28] != "" {
			psProtocolNumber, err := strconv.ParseInt(fields[28], 10, 64)
			if err == nil {
				payment.PSProtocolNumber = psProtocolNumber
			}
		}
	}

	return payment, nil
}

// parseCardStatusChange parses card status change transaction (type 27)
func parseCardStatusChange(fields []string, sourceFolder string) (models.CardStatusChange, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.CardStatusChange{}, err
	}

	cardStatusChange := models.CardStatusChange{
		BaseTransactionData: base,
	}

	// Parse specific fields for card status change
	if len(fields) > 7 {
		cardStatusChange.CardNumber = fields[7]
	}
	if len(fields) > 8 {
		cardStatusChange.CardTypeCode = fields[8]
	}
	if len(fields) > 9 {
		cardStatusChange.CardType = safeParseFloat(fields[9])
	}
	if len(fields) > 14 {
		cardStatusChange.CampaignCode = safeParseFloat(fields[14])
	}
	if len(fields) > 15 {
		cardStatusChange.EventCode = safeParseFloat(fields[15])
	}
	if len(fields) > 30 {
		cardStatusChange.OldStatus = safeParseInt(fields[30])
	}
	if len(fields) > 31 {
		cardStatusChange.NewStatus = safeParseInt(fields[31])
	}
	if len(fields) > 32 {
		cardStatusChange.NewStartDate = fields[32]
	}
	if len(fields) > 33 {
		cardStatusChange.NewEndDate = fields[33]
	}

	return cardStatusChange, nil
}

// parseModifierTransaction parses modifier transaction (types 30, 31)
func parseModifierTransaction(fields []string, sourceFolder string) (models.ModifierTransaction, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.ModifierTransaction{}, err
	}

	modifierTransaction := models.ModifierTransaction{
		BaseTransactionData: base,
	}

	// Parse specific fields for modifier transaction
	if len(fields) > 7 {
		modifierTransaction.ItemID = fields[7]
	}
	if len(fields) > 10 {
		modifierTransaction.Quantity = safeParseFloat(fields[10])
	}
	if len(fields) > 16 {
		modifierTransaction.DocumentPrintGroupCode = safeParseInt(fields[16])
	}
	if len(fields) > 34 {
		modifierTransaction.ModifierCode = fields[34]
	}

	return modifierTransaction, nil
}

// parsePrepaymentTransaction parses prepayment transaction (types 34, 84)
// parsePrepaymentTransaction parses prepayment transaction data
// Типы транзакций: 34 (Предоплата), 84 (Распределение предоплаты)
// Документация: frontol_6_integration.md, стр. 1383-1404
func parsePrepaymentTransaction(fields []string, sourceFolder string) (models.PrepaymentTransaction, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.PrepaymentTransaction{}, err
	}

	prepaymentTransaction := models.PrepaymentTransaction{
		BaseTransactionData: base,
	}

	// Поле 10: Тип предоплаты (0=внутренний документ) - fields[9]
	if len(fields) > 9 {
		prepaymentTransaction.PrepaymentType = safeParseFloat(fields[9])
	}

	// Поле 12: Сумма оплаты предоплатой - fields[11]
	if len(fields) > 11 {
		prepaymentTransaction.Amount = safeParseFloat(fields[11])
	}

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				prepaymentTransaction.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				prepaymentTransaction.ShiftNumber = shiftNumber
			}
		}
	}

	// Поле 17: Код группы печати - fields[16]
	if len(fields) > 16 {
		if fields[16] != "" {
			printGroupCode, err := strconv.ParseInt(fields[16], 10, 64)
			if err == nil {
				prepaymentTransaction.PrintGroupCode = printGroupCode
			}
		}
	}

	return prepaymentTransaction, nil
}

// parseDocumentDiscount parses document discount transaction (types 35, 37, 38, 85, 87)
// Типы транзакций: 35 (Скидка суммой), 37 (Скидка %), 38 (Округление), 85 (Распределенная суммой), 87 (Распределенная %)
// Документация: frontol_6_integration.md, стр. 235-295
func parseDocumentDiscount(fields []string, sourceFolder string) (models.DocumentDiscount, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.DocumentDiscount{}, err
	}

	documentDiscount := models.DocumentDiscount{
		BaseTransactionData: base,
	}

	// Поле 8: Информация по скидке (дисконтная карта) - fields[7]
	if len(fields) > 7 {
		documentDiscount.DiscountInfo = fields[7]
	}

	// Поле 9: пустое - fields[8] (не парсится)

	// Поле 10: Тип скидки (0-11) - fields[9]
	if len(fields) > 9 {
		documentDiscount.DiscountType = safeParseFloat(fields[9])
	}

	// Поле 11: Значение скидки - fields[10]
	if len(fields) > 10 {
		documentDiscount.DiscountValue = safeParseFloat(fields[10])
	}

	// Поле 12: Сумма скидки в базовой валюте - fields[11]
	if len(fields) > 11 {
		documentDiscount.DiscountAmount = safeParseFloat(fields[11])
	}

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				documentDiscount.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				documentDiscount.ShiftNumber = shiftNumber
			}
		}
	}

	// Поле 15: Код акции - fields[14]
	if len(fields) > 14 {
		documentDiscount.CampaignCode = safeParseInt(fields[14])
	}

	// Поле 16: Код мероприятия - fields[15]
	if len(fields) > 15 {
		documentDiscount.EventCode = safeParseInt(fields[15])
	}

	// Поле 17: Код группы печати - fields[16]
	if len(fields) > 16 {
		if fields[16] != "" {
			printGroupCode, err := strconv.ParseInt(fields[16], 10, 64)
			if err == nil {
				documentDiscount.PrintGroupCode = printGroupCode
			}
		}
	}

	return documentDiscount, nil
}

// parseNonFiscalPayment parses non-fiscal payment transaction (types 36, 86)
// Типы транзакций: 36 (Нефискальная оплата), 86 (Распределение нефискальной оплаты)
// Документация: frontol_6_integration.md, стр. 297-346
func parseNonFiscalPayment(fields []string, sourceFolder string) (models.NonFiscalPayment, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.NonFiscalPayment{}, err
	}

	nonFiscalPayment := models.NonFiscalPayment{
		BaseTransactionData: base,
	}

	// Поле 8: Номер подарочной карты - fields[7]
	if len(fields) > 7 {
		nonFiscalPayment.GiftCardNumber = fields[7]
	}

	// Поле 9: Код вида оплаты - fields[8]
	if len(fields) > 8 {
		nonFiscalPayment.PaymentTypeCode = fields[8]
	}

	// Поле 10: Операция вида оплаты (6=внутренняя, 8=внешняя) - fields[9]
	if len(fields) > 9 {
		nonFiscalPayment.PaymentTypeOperation = safeParseFloat(fields[9])
	}

	// Поле 12: Сумма оплаты - fields[11]
	if len(fields) > 11 {
		nonFiscalPayment.Amount = safeParseFloat(fields[11])
	}

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	// parseBaseTransactionData парсит operation_type из fields[8] (это PaymentTypeCode),
	// но для нефискальной оплаты operation_type находится в поле 13
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				nonFiscalPayment.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	// parseBaseTransactionData парсит shift_number из fields[7] (это GiftCardNumber),
	// но для нефискальной оплаты shift_number находится в поле 14
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				nonFiscalPayment.ShiftNumber = shiftNumber
			}
		}
	}

	// Поле 15: Код акции (для подарочных карт) - fields[14]
	if len(fields) > 14 {
		nonFiscalPayment.CampaignCode = safeParseInt(fields[14])
	}

	// Поле 16: Код мероприятия (для подарочных карт) - fields[15]
	if len(fields) > 15 {
		nonFiscalPayment.EventCode = safeParseInt(fields[15])
	}

	// Поле 17: Код группы печати - fields[16]
	if len(fields) > 16 {
		nonFiscalPayment.PositionPrintGroupCode = safeParseInt(fields[16])
	}

	// Поле 21: Код вида счетчика - fields[20]
	if len(fields) > 20 {
		nonFiscalPayment.CounterTypeCode = safeParseInt(fields[20])
	}

	// Поле 22: Код счетчика - fields[21]
	if len(fields) > 21 {
		nonFiscalPayment.CounterCode = safeParseInt(fields[21])
	}

	return nonFiscalPayment, nil
}

// parseFiscalPayment parses fiscal payment transaction (types 40, 43)
func parseFiscalPayment(fields []string, sourceFolder string) (models.FiscalPayment, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.FiscalPayment{}, err
	}

	fiscalPayment := models.FiscalPayment{
		BaseTransactionData: base,
	}

	// Parse specific fields for fiscal payment
	if len(fields) > 7 {
		fiscalPayment.CardNumber = fields[7]
	}
	if len(fields) > 8 {
		fiscalPayment.PaymentTypeCode = fields[8]
	}
	if len(fields) > 9 {
		fiscalPayment.PaymentTypeOperation = safeParseFloat(fields[9])
	}
	if len(fields) > 10 {
		fiscalPayment.CustomerAmountInPaymentCurrency = safeParseFloat(fields[10])
	}
	if len(fields) > 11 {
		fiscalPayment.CustomerAmountInBaseCurrency = safeParseFloat(fields[11])
	}

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	// В parseBaseTransactionData operation_type берется из поля 9, но для фискальных оплат
	// поле 9 - это "Код вида оплаты", а operation_type находится в поле 13
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				fiscalPayment.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	// parseBaseTransactionData парсит shift_number из fields[7] (это CardNumber),
	// но для фискальных оплат shift_number находится в поле 14
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				fiscalPayment.ShiftNumber = shiftNumber
			}
		}
	}

	// Поле 15: Код акции (для предоплаты/подарочных карт) - fields[14]
	if len(fields) > 14 {
		if fields[14] != "" {
			promotionCode, err := strconv.ParseInt(fields[14], 10, 64)
			if err == nil {
				fiscalPayment.PromotionCode = promotionCode
			}
		}
	}

	// Поле 16: Код мероприятия (для предоплаты/подарочных карт) - fields[15]
	if len(fields) > 15 {
		if fields[15] != "" {
			eventCode, err := strconv.ParseInt(fields[15], 10, 64)
			if err == nil {
				fiscalPayment.EventCode = eventCode
			}
		}
	}

	// Поле 17: Код текущей группы печати - fields[16]
	if len(fields) > 16 {
		fiscalPayment.CurrentPrintGroupCode = safeParseInt64(fields[16])
	}
	if len(fields) > 18 {
		fiscalPayment.CurrencyCode = safeParseInt64(fields[18])
	}
	if len(fields) > 19 {
		fiscalPayment.CashOutAmount = safeParseFloat(fields[19])
	}
	if len(fields) > 20 {
		fiscalPayment.CounterTypeCode = safeParseInt64(fields[20])
	}
	if len(fields) > 21 {
		fiscalPayment.CounterCode = safeParseInt64(fields[21])
	}
	if len(fields) > 28 {
		fiscalPayment.PSProtocolNumber = safeParseInt64(fields[28])
	}

	return fiscalPayment, nil
}

// parseDocumentOperation parses document operation transaction (types 42, 45, 49, 55, 56, 58, 65, 120)
func parseDocumentOperation(fields []string, sourceFolder string) (models.DocumentOperation, error) {
	base, err := parseBaseTransactionData(fields, sourceFolder)
	if err != nil {
		return models.DocumentOperation{}, err
	}

	documentOperation := models.DocumentOperation{
		BaseTransactionData: base,
	}

	// Поле 8: Номера карт клиента - fields[7]
	if len(fields) > 7 {
		documentOperation.CustomerCardNumbers = fields[7]
	}

	// Поле 9: Коды значений разрезов - fields[8]
	if len(fields) > 8 {
		documentOperation.DimensionValueCodes = fields[8]
	}

	// Поле 10: – (пустое) - fields[9]
	if len(fields) > 9 {
		documentOperation.ReservedField10 = safeParseFloat(fields[9])
	}

	// Поле 11: – (пустое для типа 55) или Количество товара - fields[10]
	if len(fields) > 10 {
		documentOperation.Quantity = safeParseFloat(fields[10])
	}

	// Поле 12: Итоговая сумма документа - fields[11]
	if len(fields) > 11 {
		documentOperation.TotalAmount = safeParseFloat(fields[11])
	}

	// Поле 13: Операция - fields[12] (переопределяем из правильного поля)
	// parseBaseTransactionData парсит operation_type из fields[8] (это DimensionValueCodes),
	// но для операций документа operation_type находится в поле 13
	if len(fields) > 12 {
		if fields[12] != "" {
			operationType, err := strconv.ParseInt(fields[12], 10, 64)
			if err == nil {
				documentOperation.OperationType = operationType
			}
		}
	}

	// Поле 14: Номер смены - fields[13] (переопределяем из правильного поля)
	// parseBaseTransactionData парсит shift_number из fields[7] (это CustomerCardNumbers),
	// но для операций документа shift_number находится в поле 14
	if len(fields) > 13 {
		if fields[13] != "" {
			shiftNumber, err := strconv.ParseInt(fields[13], 10, 64)
			if err == nil {
				documentOperation.ShiftNumber = shiftNumber
			}
		}
	}

	// Поле 15: Код клиента - fields[14]
	if len(fields) > 14 {
		documentOperation.CustomerCode = safeParseFloat(fields[14])
	}

	// Поле 16: – (пустое) - fields[15]
	if len(fields) > 15 {
		documentOperation.ReservedField16 = safeParseFloat(fields[15])
	}

	// Поле 17: Код группы печати документа - fields[16]
	if len(fields) > 16 {
		documentOperation.DocumentPrintGroupCode = safeParseInt(fields[16])
	}
	if len(fields) > 17 {
		documentOperation.BonusAmount = fields[17]
	}
	if len(fields) > 18 {
		documentOperation.OrderID = fields[18]
	}
	if len(fields) > 19 {
		documentOperation.DocumentAmountWithoutDiscounts = safeParseFloat(fields[19])
	}
	if len(fields) > 20 {
		documentOperation.VisitorCount = safeParseInt(fields[20])
	}
	if len(fields) > 21 {
		documentOperation.CorrectionType = safeParseInt(fields[21])
	}
	if len(fields) > 22 {
		documentOperation.KKTRegistrationNumber = safeParseInt(fields[22])
	}
	if len(fields) > 23 {
		documentOperation.DocumentTypeCode = safeParseInt(fields[23])
	}
	if len(fields) > 24 {
		documentOperation.CommentCode = safeParseInt(fields[24])
	}
	if len(fields) > 25 {
		documentOperation.BaseDocumentNumber = safeParseInt(fields[25])
	}
	if len(fields) > 27 {
		documentOperation.EmployeeCode = safeParseInt(fields[27])
	}
	if len(fields) > 28 {
		documentOperation.EmployeeListEditDocumentNumber = safeParseInt(fields[28])
	}
	if len(fields) > 29 {
		documentOperation.DepartmentCode = fields[29]
	}
	if len(fields) > 30 {
		documentOperation.HallCode = safeParseInt(fields[30])
	}
	if len(fields) > 31 {
		documentOperation.ServicePointCode = safeParseInt(fields[31])
	}
	if len(fields) > 32 {
		documentOperation.ReservationID = fields[32]
	}
	if len(fields) > 33 {
		documentOperation.UserVariableValue = fields[33]
	}
	if len(fields) > 34 {
		documentOperation.ExternalComment = fields[34]
	}
	if len(fields) > 35 {
		documentOperation.RevaluationDateTime = fields[35]
	}
	if len(fields) > 36 {
		documentOperation.ContractorCode = safeParseInt(fields[36])
	}
	if len(fields) > 37 {
		documentOperation.DepartmentID = fields[37]
	}
	// Field 39: – (empty, for compliance with Frontol 6)
	if len(fields) > 38 {
		documentOperation.ReservedField39 = fields[38]
	}
	if len(fields) > 42 {
		documentOperation.CouponsOnDocument = fields[42]
	}
	if len(fields) > 43 {
		documentOperation.CalculationDateTime = fields[43]
	}

	return documentOperation, nil
}
