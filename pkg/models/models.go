package models

import (
	"time"
)

// Config represents application configuration
type Config struct {
	// Database settings
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// FTP settings
	FTPHost        string
	FTPPort        int
	FTPUser        string
	FTPPassword    string
	FTPRequestDir  string
	FTPResponseDir string
	FTPPoolSize    int // Number of FTP connections in pool (default: 5)
	KassaStructure map[string][]string

	// Application settings
	LocalDir         string
	BatchSize        int
	MaxRetries       int
	RetryDelay       time.Duration
	WaitDelayMinutes time.Duration
	WorkerPoolSize   int // Number of concurrent file processing workers (default: 10)
	LogLevel         string
	LogBackend       string // slog or zerolog

	// Webhook server settings
	ServerPort            int
	WebhookReportURL      string
	WebhookTimeoutMinutes int           // Timeout for sending webhook report (0 = no timeout, send only on completion)
	WebhookBearerToken    string        // Bearer token for webhook authorization
	ShutdownTimeout       time.Duration // Graceful shutdown timeout (default: 30 seconds)

	// RabbitMQ settings
	RabbitMQURL             string
	RabbitMQHost            string
	RabbitMQPort            int
	RabbitMQVHost           string
	RabbitMQUser            string
	RabbitMQPassword        string
	RabbitMQPrefetch        int
	QueueRetryMax           int
	QueueRetryBackoffs      []time.Duration // Parsed from CSV milliseconds
	QueueDLQRequeueInterval time.Duration
	QueueDeclareOnPublish   bool
	QueueProvider           string // rabbitmq | memory (for gradual rollout)
	RabbitMQManagementURL   string
}

// KassaFolder represents a kassa folder structure
type KassaFolder struct {
	KassaCode    string
	FolderName   string
	RequestPath  string
	ResponsePath string
}

// BaseTransactionData represents common fields for all transaction types
type BaseTransactionData struct {
	ID               string    `json:"id"`
	SourceFolder     string    `json:"source_folder"`
	Date             time.Time `json:"date"`
	Time             time.Time `json:"time"`
	TransactionType  int       `json:"transaction_type"`
	CashRegisterCode int64     `json:"cash_register_code"`
	DocumentNumber   string    `json:"document_number"`
	CashierCode      string    `json:"cashier_code"`
	ShiftNumber      int64     `json:"shift_number"`
	OperationType    int64     `json:"operation_type"`
	PrintGroupCode   int64     `json:"print_group_code"`
	DocumentTypeCode int64     `json:"document_type_code"`
	DocumentInfo     string    `json:"document_info"`
	EnterpriseID     int64     `json:"enterprise_id"`
	EmployeeCode     string    `json:"employee_code"`
	RawData          string    `json:"raw_data"`
}

// TransactionRegistration represents main transaction data (transaction_registrations table)
// Fields are ordered according to their position in the file (1, 2, 3, 4...)
type TransactionRegistration struct {
	// Fields 1-7: Basic transaction info
	TransactionIDUnique int64  `json:"transaction_id_unique"` // Field 1
	SourceFolder        string `json:"source_folder"`         // Source folder
	TransactionDate     string `json:"transaction_date"`      // Field 2
	TransactionTime     string `json:"transaction_time"`      // Field 3
	TransactionType     int    `json:"transaction_type"`      // Field 4
	CashRegisterCode    int64  `json:"cash_register_code"`    // Field 5
	DocumentNumber      int64  `json:"document_number"`       // Field 6
	CashierCode         int64  `json:"cashier_code"`          // Field 7

	// Fields 8-14: Item and operation details
	ItemCode           string  `json:"item_code"`            // Field 8
	GroupCode          string  `json:"group_code"`           // Field 9
	AmountTotal        float64 `json:"amount_total"`         // Field 10
	Quantity           float64 `json:"quantity"`             // Field 11
	AmountCashRegister float64 `json:"amount_cash_register"` // Field 12
	OperationType      int64   `json:"operation_type"`       // Field 13
	ShiftNumber        int64   `json:"shift_number"`         // Field 14

	// Fields 15-23: Item details and pricing
	ItemPrice           float64 `json:"item_price"`           // Field 15
	ItemSum             float64 `json:"item_sum"`             // Field 16
	PrintGroupCode      int64   `json:"print_group_code"`     // Field 17: Код группы печати (according to Frontol 6 docs)
	ItemLineNumber      int     `json:"item_line_number"`     // Field 17: Legacy field (kept for backward compatibility)
	ArticleSKU          string  `json:"article_sku"`          // Field 18
	RegistrationBarcode float64 `json:"registration_barcode"` // Field 19
	PositionAmount      float64 `json:"position_amount"`      // Field 20
	KKTSection          string  `json:"kkt_section"`          // Field 21
	ReservedField22     int64   `json:"reserved_field22"`     // Field 22
	DocumentTypeCode    int64   `json:"document_type_code"`   // Field 23

	// Fields 24-31: Document and enterprise info
	CommentCode     int64  `json:"comment_code"`     // Field 24
	ReservedField25 string `json:"reserved_field25"` // Field 25
	DocumentInfo    string `json:"document_info"`    // Field 26
	EnterpriseID    int64  `json:"enterprise_id"`    // Field 27
	EmployeeCode    int64  `json:"employee_code"`    // Field 28
	DividedPackQty  int64  `json:"divided_pack_qty"` // Field 29
	GiftCardNumber  string `json:"gift_card_number"` // Field 30
	PackQuantity    int64  `json:"pack_quantity"`    // Field 31

	// Fields 32-43: Nomenclature and marking
	NomenclatureType  int     `json:"nomenclature_type"`  // Field 32
	MarkingCode       string  `json:"marking_code"`       // Field 33
	ExciseStamp       string  `json:"excise_stamp"`       // Field 34
	PersonalModGroup  string  `json:"personal_mod_group"` // Field 35
	LotteryTime       string  `json:"lottery_time"`       // Field 36
	LotteryID         int64   `json:"lottery_id"`         // Field 37
	ReservedField38   int64   `json:"reserved_field38"`   // Field 38
	ALCCode           string  `json:"alc_code"`           // Field 39
	ReservedField40   float64 `json:"reserved_field40"`   // Field 40
	PrescriptionData1 string  `json:"prescription_data1"` // Field 41
	PrescriptionData2 string  `json:"prescription_data2"` // Field 42
	CouponsPerItem    string  `json:"coupons_per_item"`   // Field 43

	// For debugging and audit
	RawData string `json:"raw_data"`
}

// SpecialPrice represents special price data (special_prices table)
type SpecialPrice struct {
	BaseTransactionData
	PriceListCode    string  `json:"price_list_code"`
	GroupCode        string  `json:"group_code"`
	PriceType        int     `json:"price_type"`
	SpecialPrice     float64 `json:"special_price"`
	ProductCardPrice float64 `json:"product_card_price"`
	PromotionCode    int64   `json:"promotion_code"`
	EventCode        int64   `json:"event_code"`
	PrintGroupCode   int64   `json:"print_group_code"`
}

// BonusTransaction represents bonus transaction data (bonus_transactions table)
type BonusTransaction struct {
	BaseTransactionData
	BonusAmount        float64 `json:"bonus_amount"`
	AccruedBonusAmount float64 `json:"accrued_bonus_amount"` // Поле 12: Начисленная сумма бонуса
	BonusType          int     `json:"bonus_type"`           // Поле 10: Тип бонуса (0=внутренний, 1=внешний)
	CardNumber         string  `json:"card_number"`
	PromotionCode      int64   `json:"promotion_code"`     // Поле 15: Код акции
	EventCode          int64   `json:"event_code"`         // Поле 16: Код мероприятия
	PrintGroupCode     int64   `json:"print_group_code"`   // Поле 17: Код группы печати
	PSProtocolNumber   int64   `json:"ps_protocol_number"` // Поле 29: Номер протокола ПС
}

// DiscountTransaction represents discount transaction data (discount_transactions table)
type DiscountTransaction struct {
	BaseTransactionData
	DiscountAmount  float64 `json:"discount_amount"`
	DiscountType    int     `json:"discount_type"`
	DiscountValue   float64 `json:"discount_value"`   // Поле 11: Значение скидки (процент для типа 17, сумма для типа 15)
	DiscountPercent float64 `json:"discount_percent"` // Для типа 17 берется из discount_value
	PromotionCode   int64   `json:"promotion_code"`
	EventCode       int64   `json:"event_code"`
	PrintGroupCode  int64   `json:"print_group_code"`
}

// BillRegistration represents bill registration data (bill_registrations table)
type BillRegistration struct {
	BaseTransactionData
	BillCode         string  `json:"bill_code"`         // Поле №8: Код купюры
	GroupCode        string  `json:"group_code"`        // Поле №9: Код группы
	BillDenomination float64 `json:"bill_denomination"` // Поле №10: Достоинство купюры
	BillQuantity     float64 `json:"bill_quantity"`     // Поле №11: Количество купюр
	BillAmount       float64 `json:"bill_amount"`       // Поле №12: Сумма купюр
	BillNumber       string  `json:"bill_number"`       // Поле №15: Номер купюры
	BillTotalAmount  float64 `json:"bill_total_amount"` // Поле №16: Общая сумма купюр
	BillType         int64   `json:"bill_type"`         // Поле №17: Тип купюры
	PrintGroupCode   int64   `json:"print_group_code"`  // Поле №18: Код группы печати
	CustomerCode     string  `json:"customer_code"`     // Поле №19: Код клиента
	ReservedField20  string  `json:"reserved_field20"`  // Поле №20: Зарезервировано
	ReservedField21  int64   `json:"reserved_field21"`  // Поле №21: Зарезервировано
	ReservedField22  int64   `json:"reserved_field22"`  // Поле №22: Зарезервировано
	ReservedField24  int64   `json:"reserved_field24"`  // Поле №24: Зарезервировано
	ReservedField25  int64   `json:"reserved_field25"`  // Поле №25: Зарезервировано
}

// EmployeeEdit represents employee edit data (employee_edits table)
type EmployeeEdit struct {
	BaseTransactionData
	EmployeeCode string `json:"employee_code"` // Поле 8: Код сотрудника
	// Поля 9-17 пустые по документации для типа 25
	// EmployeeName, EmployeePosition, EmployeeDepartment, EditType, PrintGroupCode
	// не используются для типа 25 согласно документации
}

// EmployeeAccounting represents employee accounting data (employee_accounting table)
type EmployeeAccounting struct {
	BaseTransactionData
	EmployeeCode   string `json:"employee_code"`    // Поле 8: Код сотрудника
	PrintGroupCode int64  `json:"print_group_code"` // Поле 17: Код группы печати документа
	// Поля 9-16 пустые по документации для типов 26, 29
	// AccountingType, AccountingAmount, AccountingDate не описаны в документации
}

// VatKKTTransaction represents VAT KKT transaction data (vat_kkt_transactions table)
type VatKKTTransaction struct {
	BaseTransactionData
	Vat0Amount       float64 `json:"vat_0_amount"`
	Vat10Amount      float64 `json:"vat_10_amount"`
	Vat20Amount      float64 `json:"vat_20_amount"`
	AmountWithoutVat float64 `json:"amount_without_vat"`
	Vat10_110Amount  float64 `json:"vat_10_110_amount"`
	Vat20_120Amount  float64 `json:"vat_20_120_amount"`
	ReservedFields   string  `json:"reserved_fields"`
}

// AdditionalTransaction represents additional transaction data (additional_transactions table)
type AdditionalTransaction struct {
	BaseTransactionData
	AdditionalType   int     `json:"additional_type"`
	AdditionalAmount float64 `json:"additional_amount"`
	AdditionalInfo   string  `json:"additional_info"`
	PrintGroupCode   int64   `json:"print_group_code"`
}

// AstuExchangeTransaction represents ASTU exchange transaction data (astu_exchange_transactions table)
type AstuExchangeTransaction struct {
	BaseTransactionData
	ExchangeType   int     `json:"exchange_type"`
	ExchangeAmount float64 `json:"exchange_amount"`
	ExchangeRate   float64 `json:"exchange_rate"`
	PrintGroupCode int64   `json:"print_group_code"`
}

// CounterChangeTransaction represents counter change transaction data (counter_change_transactions table)
type CounterChangeTransaction struct {
	BaseTransactionData
	CardNumberOrClientCode string  `json:"card_number_client_code"` // Поле 8: Номер карты/Код клиента
	CardTypeCode           string  `json:"card_type_code"`          // Поле 9: Код вида карты
	BindingType            float64 `json:"binding_type"`            // Поле 10: Привязка (1-4)
	ValueAfterChanges      float64 `json:"value_after_changes"`     // Поле 11: Значение после изменений
	ChangeAmount           float64 `json:"change_amount"`           // Поле 12: Сумма изменения счетчика
	PromotionCode          int64   `json:"promotion_code"`          // Поле 15: Код акции
	EventCode              int64   `json:"event_code"`              // Поле 16: Код мероприятия
	CounterTypeCode        int64   `json:"counter_type_code"`       // Поле 21: Код вида счетчика
	CounterCode            int64   `json:"counter_code"`            // Поле 22: Код счетчика
	MovementStartDate      string  `json:"movement_start_date"`     // Поле 30: Дата начала действия движения
	CardStartDate          string  `json:"card_start_date"`         // Поле 33: Дата начала действия карты
	CardEndDate            string  `json:"card_end_date"`           // Поле 34: Дата окончания действия карты
	MovementEndDate        string  `json:"movement_end_date"`       // Поле 35: Дата окончания действия движения
	// Legacy поля для обратной совместимости
	CounterType   int   `json:"counter_type"`
	CounterValue  int64 `json:"counter_value"`
	CounterChange int64 `json:"counter_change"`
}

// KKTShiftReport represents KKT shift report data (kkt_shift_reports table)
type KKTShiftReport struct {
	BaseTransactionData
	ReportType     int     `json:"report_type"`
	ReportData     string  `json:"report_data"`
	ReportAmount   float64 `json:"report_amount"`
	PrintGroupCode int64   `json:"print_group_code"`
}

// FrontolMarkUnitTransaction represents Frontol mark unit transaction data (frontol_mark_unit_transactions table)
type FrontolMarkUnitTransaction struct {
	BaseTransactionData
	MarkUnitType   int    `json:"mark_unit_type"`
	MarkUnitCode   string `json:"mark_unit_code"`
	MarkUnitData   string `json:"mark_unit_data"`
	PrintGroupCode int64  `json:"print_group_code"`
}

// BonusPayment represents bonus payment data (bonus_payments table)
type BonusPayment struct {
	BaseTransactionData
	BonusCardNumber    string  `json:"bonus_card_number"`    // Поле 8: Номер бонусной карты
	PaymentType        int     `json:"payment_type"`         // Поле 10: Тип оплаты бонусом (0=внутренний, 1=внешний)
	CounterChangeValue float64 `json:"counter_change_value"` // Поле 11: Величина изменения счетчика
	PaymentAmount      float64 `json:"payment_amount"`       // Поле 12: Сумма оплаты
	CardNumber         string  `json:"card_number"`          // Legacy поле (для обратной совместимости)
	PromotionCode      int64   `json:"promotion_code"`       // Поле 15: Код акции
	EventCode          int64   `json:"event_code"`           // Поле 16: Код мероприятия
	PrintGroupCode     int64   `json:"print_group_code"`     // Поле 17: Код группы печати
	PSProtocolNumber   int64   `json:"ps_protocol_number"`   // Поле 29: Номер протокола ПС
}

// FileHeader represents the first 3 lines of a Frontol file
type FileHeader struct {
	Processed bool   `json:"processed"`
	DBID      string `json:"db_id"`
	ReportNum string `json:"report_num"`
}

// ProcessingStats represents processing statistics
type ProcessingStats struct {
	StartTime          time.Time
	FilesProcessed     int
	FilesSkipped       int
	TransactionsLoaded int
	Errors             int
}

// AvgTimePerFile calculates average processing time per file
func (ps *ProcessingStats) AvgTimePerFile() time.Duration {
	if ps.FilesProcessed == 0 {
		return 0
	}
	return time.Since(ps.StartTime) / time.Duration(ps.FilesProcessed)
}

// CardStatusChange represents card status change transaction (type 27)
type CardStatusChange struct {
	BaseTransactionData
	CardNumber   string
	CardTypeCode string
	CardType     float64
	CampaignCode float64
	EventCode    float64
	OldStatus    int
	NewStatus    int
	NewStartDate string
	NewEndDate   string
}

// ModifierTransaction represents modifier transaction (types 30, 31)
type ModifierTransaction struct {
	BaseTransactionData
	ItemID                 string
	Quantity               float64
	DocumentPrintGroupCode int
	ModifierCode           string
}

// PrepaymentTransaction represents prepayment transaction (types 34, 84)
type PrepaymentTransaction struct {
	BaseTransactionData
	PrepaymentType float64 `json:"prepayment_type"`  // Поле 10: Тип предоплаты (0=внутренний документ)
	Amount         float64 `json:"amount"`           // Поле 12: Сумма оплаты предоплатой
	PrintGroupCode int64   `json:"print_group_code"` // Поле 17: Код группы печати
}

// DocumentDiscount represents document discount transaction (types 35, 37, 38, 85, 87)
type DocumentDiscount struct {
	BaseTransactionData
	DiscountInfo   string  `json:"discount_info"`    // Поле 8: Информация по скидке
	DiscountType   float64 `json:"discount_type"`    // Поле 10: Тип скидки
	DiscountValue  float64 `json:"discount_value"`   // Поле 11: Значение скидки
	DiscountAmount float64 `json:"discount_amount"`  // Поле 12: Сумма скидки
	CampaignCode   int     `json:"campaign_code"`    // Поле 15: Код акции
	EventCode      int     `json:"event_code"`       // Поле 16: Код мероприятия
	PrintGroupCode int64   `json:"print_group_code"` // Поле 17: Код группы печати
}

// NonFiscalPayment represents non-fiscal payment transaction (types 36, 86)
type NonFiscalPayment struct {
	BaseTransactionData
	GiftCardNumber         string
	PaymentTypeCode        string
	PaymentTypeOperation   float64
	Amount                 float64
	CampaignCode           int
	EventCode              int
	PositionPrintGroupCode int
	CounterTypeCode        int
	CounterCode            int
}

// FiscalPayment represents fiscal payment transaction (types 40, 43)
type FiscalPayment struct {
	BaseTransactionData
	CardNumber                      string  `json:"card_number"`                         // Поле 8: Номер карты
	PaymentTypeCode                 string  `json:"payment_type_code"`                   // Поле 9: Код вида оплаты
	PaymentTypeOperation            float64 `json:"payment_type_operation"`              // Поле 10: Операция вида оплаты
	CustomerAmountInPaymentCurrency float64 `json:"customer_amount_in_payment_currency"` // Поле 11: Сумма клиента в валюте оплаты
	CustomerAmountInBaseCurrency    float64 `json:"customer_amount_in_base_currency"`    // Поле 12: Сумма клиента в базовой валюте
	CurrentPrintGroupCode           int64   `json:"current_print_group_code"`            // Поле 17: Код текущей группы печати
	CurrencyCode                    int64   `json:"currency_code"`                       // Поле 19: Код валюты
	CashOutAmount                   float64 `json:"cash_out_amount"`                     // Поле 20: Сумма выдачи наличных
	CounterTypeCode                 int64   `json:"counter_type_code"`                   // Поле 21: Код вида счетчика
	CounterCode                     int64   `json:"counter_code"`                        // Поле 22: Код счетчика
	PSProtocolNumber                int64   `json:"ps_protocol_number"`                  // Поле 29: Номер протокола ПС
	PromotionCode                   int64   `json:"promotion_code"`                      // Поле 15: Код акции (для предоплаты/подарочных карт)
	EventCode                       int64   `json:"event_code"`                          // Поле 16: Код мероприятия (для предоплаты/подарочных карт)
}

// DocumentOperation represents document operation transaction (types 42, 45, 49, 55, 56, 58, 65, 120)
type DocumentOperation struct {
	BaseTransactionData
	CustomerCardNumbers            string
	DimensionValueCodes            string
	ReservedField10                float64 // Field 10: – (empty, for compliance with Frontol 6)
	Quantity                       float64
	TotalAmount                    float64
	CustomerCode                   float64
	ReservedField16                float64 // Field 16: – (empty, for compliance with Frontol 6)
	DocumentPrintGroupCode         int
	BonusAmount                    string
	OrderID                        string
	DocumentAmountWithoutDiscounts float64
	VisitorCount                   int
	CorrectionType                 int
	KKTRegistrationNumber          int
	DocumentTypeCode               int
	CommentCode                    int
	BaseDocumentNumber             int
	EmployeeCode                   int
	EmployeeListEditDocumentNumber int
	DepartmentCode                 string
	HallCode                       int
	ServicePointCode               int
	ReservationID                  string
	UserVariableValue              string
	ExternalComment                string
	RevaluationDateTime            string
	ContractorCode                 int
	DepartmentID                   string
	ReservedField39                string // Field 39: – (empty, for compliance with Frontol 6)
	CouponsOnDocument              string
	CalculationDateTime            string
}
