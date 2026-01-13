package models

import "time"

type TxItemRegistration1_11 struct {
	TransactionIDUnique        int64     `json:"transaction_id_unique"`
	SourceFolder               string    `json:"source_folder"`
	TransactionDate            time.Time `json:"transaction_date"`
	TransactionTime            time.Time `json:"transaction_time"`
	TransactionType            int64     `json:"transaction_type"`
	CashRegisterCode           int64     `json:"cash_register_code"`
	DocumentNumber             int64     `json:"document_number"`
	CashierCode                int64     `json:"cashier_code"`
	ItemIdentifier             string    `json:"item_identifier"`
	DimensionValueCodes        string    `json:"dimension_value_codes"`
	PriceWithoutDiscounts      float64   `json:"price_without_discounts"`
	Quantity                   float64   `json:"quantity"`
	PositionAmountWithRounding float64   `json:"position_amount_with_rounding"`
	OperationType              int64     `json:"operation_type"`
	ShiftNumber                int64     `json:"shift_number"`
	FinalPrice                 float64   `json:"final_price"`
	PositionTotalAmount        float64   `json:"position_total_amount"`
	PrintGroupCode             int64     `json:"print_group_code"`
	ArticleSKU                 string    `json:"article_sku"`
	RegistrationBarcode        string    `json:"registration_barcode"`
	PositionAmountBase         float64   `json:"position_amount_base"`
	KKTSection                 int64     `json:"kkt_section"`
	Reserved22                 int64     `json:"reserved_22"`
	DocumentTypeCode           int64     `json:"document_type_code"`
	CommentCode                int64     `json:"comment_code"`
	Reserved25                 int64     `json:"reserved_25"`
	DocumentInfo               string    `json:"document_info"`
	EnterpriseID               int64     `json:"enterprise_id"`
	EmployeeCode               int64     `json:"employee_code"`
	SplitPackQuantity          int64     `json:"split_pack_quantity"`
	GiftCardExternalNumber     string    `json:"gift_card_external_number"`
	PackQuantity               int64     `json:"pack_quantity"`
	ItemTypeCode               int64     `json:"item_type_code"`
	MarkingCode                string    `json:"marking_code"`
	ExciseMarks                string    `json:"excise_marks"`
	PersonalModifierGroupCode  string    `json:"personal_modifier_group_code"`
	StolotoRegistrationTime    time.Time `json:"stoloto_registration_time"`
	StolotoTicketID            int64     `json:"stoloto_ticket_id"`
	Reserved38                 int64     `json:"reserved_38"`
	ALCCode                    string    `json:"alc_code"`
	Reserved40                 float64   `json:"reserved_40"`
	PrescriptionData1          string    `json:"prescription_data_1"`
	PrescriptionData2          string    `json:"prescription_data_2"`
	PositionCoupons            string    `json:"position_coupons"`
	Reserved44                 time.Time `json:"reserved_44"`
}

type TxItemTax4_14 struct {
	TransactionIDUnique          int64     `json:"transaction_id_unique"`
	SourceFolder                 string    `json:"source_folder"`
	TransactionDate              time.Time `json:"transaction_date"`
	TransactionTime              time.Time `json:"transaction_time"`
	TransactionType              int64     `json:"transaction_type"`
	CashRegisterCode             int64     `json:"cash_register_code"`
	DocumentNumber               int64     `json:"document_number"`
	CashierCode                  int64     `json:"cashier_code"`
	Reserved8                    string    `json:"reserved_8"`
	DimensionValueCodes          string    `json:"dimension_value_codes"`
	TaxGroupCode                 int64     `json:"tax_group_code"`
	TaxRateCode                  int64     `json:"tax_rate_code"`
	TaxAmountBase                float64   `json:"tax_amount_base"`
	OperationType                int64     `json:"operation_type"`
	ShiftNumber                  int64     `json:"shift_number"`
	Reserved15                   float64   `json:"reserved_15"`
	TotalAmountBaseWithDiscounts float64   `json:"total_amount_base_with_discounts"`
	PrintGroupCode               int64     `json:"print_group_code"`
	Reserved18                   string    `json:"reserved_18"`
	Reserved19                   int64     `json:"reserved_19"`
	AmountBaseWithoutDiscounts   float64   `json:"amount_base_without_discounts"`
	Reserved21                   int64     `json:"reserved_21"`
	Reserved22                   int64     `json:"reserved_22"`
	DocumentTypeCode             int64     `json:"document_type_code"`
	Reserved24                   int64     `json:"reserved_24"`
	Reserved25                   int64     `json:"reserved_25"`
	DocumentInfo                 string    `json:"document_info"`
	EnterpriseID                 int64     `json:"enterprise_id"`
	Reserved28                   int64     `json:"reserved_28"`
	Reserved29                   int64     `json:"reserved_29"`
	Reserved30                   string    `json:"reserved_30"`
	Reserved31                   int64     `json:"reserved_31"`
	Reserved32                   int64     `json:"reserved_32"`
	Reserved33                   string    `json:"reserved_33"`
	Reserved34                   string    `json:"reserved_34"`
	Reserved35                   string    `json:"reserved_35"`
	Reserved36                   time.Time `json:"reserved_36"`
}

type TxItemKKT6_16 struct {
	TransactionIDUnique            int64     `json:"transaction_id_unique"`
	SourceFolder                   string    `json:"source_folder"`
	TransactionDate                time.Time `json:"transaction_date"`
	TransactionTime                time.Time `json:"transaction_time"`
	TransactionType                int64     `json:"transaction_type"`
	CashRegisterCode               int64     `json:"cash_register_code"`
	DocumentNumber                 int64     `json:"document_number"`
	CashierCode                    int64     `json:"cashier_code"`
	ItemIdentifier                 string    `json:"item_identifier"`
	DimensionValueCodes            string    `json:"dimension_value_codes"`
	Reserved10                     float64   `json:"reserved_10"`
	QuantityKKT                    float64   `json:"quantity_kkt"`
	Reserved12                     float64   `json:"reserved_12"`
	OperationType                  int64     `json:"operation_type"`
	ShiftNumber                    int64     `json:"shift_number"`
	FinalPriceKKTCurrency          float64   `json:"final_price_kkt_currency"`
	PositionTotalAmountKKTCurrency float64   `json:"position_total_amount_kkt_currency"`
	PrintGroupCode                 int64     `json:"print_group_code"`
	ArticleSKU                     string    `json:"article_sku"`
	RegistrationBarcode            string    `json:"registration_barcode"`
	Reserved20                     float64   `json:"reserved_20"`
	KKTSection                     int64     `json:"kkt_section"`
	Reserved22                     int64     `json:"reserved_22"`
	DocumentTypeCode               int64     `json:"document_type_code"`
	CommentCode                    int64     `json:"comment_code"`
	Reserved25                     int64     `json:"reserved_25"`
	DocumentInfo                   string    `json:"document_info"`
	EnterpriseID                   int64     `json:"enterprise_id"`
	EmployeeCode                   int64     `json:"employee_code"`
	SplitPackQuantity              int64     `json:"split_pack_quantity"`
	GiftCardExternalNumber         string    `json:"gift_card_external_number"`
	PackQuantity                   int64     `json:"pack_quantity"`
	ItemTypeCode                   int64     `json:"item_type_code"`
	MarkingCode                    string    `json:"marking_code"`
	ExciseMarks                    string    `json:"excise_marks"`
	PersonalModifierGroupCode      string    `json:"personal_modifier_group_code"`
	StolotoRegistrationTime        time.Time `json:"stoloto_registration_time"`
	StolotoTicketID                int64     `json:"stoloto_ticket_id"`
	Reserved38                     int64     `json:"reserved_38"`
	ALCCode                        string    `json:"alc_code"`
	Reserved40                     float64   `json:"reserved_40"`
	PrescriptionData1              string    `json:"prescription_data_1"`
	PrescriptionData2              string    `json:"prescription_data_2"`
	PositionCoupons                string    `json:"position_coupons"`
	Reserved44                     time.Time `json:"reserved_44"`
}

type TxDocumentDiscount35 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	DiscountInfo        string    `json:"discount_info"`
	Reserved9           string    `json:"reserved_9"`
	DiscountType        float64   `json:"discount_type"`
	DiscountValue       float64   `json:"discount_value"`
	DiscountAmountBase  float64   `json:"discount_amount_base"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	PromotionCode       int64     `json:"promotion_code"`
	EventCode           int64     `json:"event_code"`
	PrintGroupCode      int64     `json:"print_group_code"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	Reserved21          int64     `json:"reserved_21"`
	Reserved22          int64     `json:"reserved_22"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	Reserved29          int64     `json:"reserved_29"`
	Reserved30          string    `json:"reserved_30"`
	Reserved31          int64     `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	Reserved33          string    `json:"reserved_33"`
	Reserved34          string    `json:"reserved_34"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxNonFiscalPayment36 struct {
	TransactionIDUnique  int64     `json:"transaction_id_unique"`
	SourceFolder         string    `json:"source_folder"`
	TransactionDate      time.Time `json:"transaction_date"`
	TransactionTime      time.Time `json:"transaction_time"`
	TransactionType      int64     `json:"transaction_type"`
	CashRegisterCode     int64     `json:"cash_register_code"`
	DocumentNumber       int64     `json:"document_number"`
	CashierCode          int64     `json:"cashier_code"`
	GiftCardNumber       string    `json:"gift_card_number"`
	PaymentTypeCode      string    `json:"payment_type_code"`
	PaymentTypeOperation float64   `json:"payment_type_operation"`
	Reserved11           float64   `json:"reserved_11"`
	PaymentAmount        float64   `json:"payment_amount"`
	OperationType        int64     `json:"operation_type"`
	ShiftNumber          int64     `json:"shift_number"`
	PromotionCode        int64     `json:"promotion_code"`
	EventCode            int64     `json:"event_code"`
	PrintGroupCode       int64     `json:"print_group_code"`
	Reserved18           string    `json:"reserved_18"`
	Reserved19           int64     `json:"reserved_19"`
	Reserved20           float64   `json:"reserved_20"`
	CounterTypeCode      int64     `json:"counter_type_code"`
	CounterCode          int64     `json:"counter_code"`
	DocumentTypeCode     int64     `json:"document_type_code"`
	Reserved24           int64     `json:"reserved_24"`
	Reserved25           int64     `json:"reserved_25"`
	DocumentInfo         string    `json:"document_info"`
	EnterpriseID         int64     `json:"enterprise_id"`
	Reserved28           int64     `json:"reserved_28"`
	Reserved29           int64     `json:"reserved_29"`
	Reserved30           string    `json:"reserved_30"`
	Reserved31           int64     `json:"reserved_31"`
	Reserved32           int64     `json:"reserved_32"`
	Reserved33           string    `json:"reserved_33"`
	Reserved34           string    `json:"reserved_34"`
	Reserved35           string    `json:"reserved_35"`
	Reserved36           time.Time `json:"reserved_36"`
}

type TxFiscalPayment40 struct {
	TransactionIDUnique           int64     `json:"transaction_id_unique"`
	SourceFolder                  string    `json:"source_folder"`
	TransactionDate               time.Time `json:"transaction_date"`
	TransactionTime               time.Time `json:"transaction_time"`
	TransactionType               int64     `json:"transaction_type"`
	CashRegisterCode              int64     `json:"cash_register_code"`
	DocumentNumber                int64     `json:"document_number"`
	CashierCode                   int64     `json:"cashier_code"`
	CardNumber                    string    `json:"card_number"`
	PaymentTypeCode               string    `json:"payment_type_code"`
	PaymentTypeOperation          float64   `json:"payment_type_operation"`
	CustomerAmountPaymentCurrency float64   `json:"customer_amount_payment_currency"`
	CustomerAmountBaseCurrency    float64   `json:"customer_amount_base_currency"`
	OperationType                 int64     `json:"operation_type"`
	ShiftNumber                   int64     `json:"shift_number"`
	PromotionCode                 int64     `json:"promotion_code"`
	EventCode                     int64     `json:"event_code"`
	CurrentPrintGroupCode         int64     `json:"current_print_group_code"`
	Reserved18                    string    `json:"reserved_18"`
	CurrencyCode                  int64     `json:"currency_code"`
	CashOutAmount                 float64   `json:"cash_out_amount"`
	CounterTypeCode               int64     `json:"counter_type_code"`
	CounterCode                   int64     `json:"counter_code"`
	DocumentTypeCode              int64     `json:"document_type_code"`
	Reserved24                    int64     `json:"reserved_24"`
	Reserved25                    int64     `json:"reserved_25"`
	DocumentInfo                  string    `json:"document_info"`
	EnterpriseID                  int64     `json:"enterprise_id"`
	Reserved28                    int64     `json:"reserved_28"`
	PSProtocolNumber              int64     `json:"ps_protocol_number"`
	Reserved30                    string    `json:"reserved_30"`
	Reserved31                    int64     `json:"reserved_31"`
	Reserved32                    int64     `json:"reserved_32"`
	Reserved33                    string    `json:"reserved_33"`
	Reserved34                    string    `json:"reserved_34"`
	Reserved35                    string    `json:"reserved_35"`
	Reserved36                    time.Time `json:"reserved_36"`
}

type TxDocumentOpen42 struct {
	TransactionIDUnique            int64     `json:"transaction_id_unique"`
	SourceFolder                   string    `json:"source_folder"`
	TransactionDate                time.Time `json:"transaction_date"`
	TransactionTime                time.Time `json:"transaction_time"`
	TransactionType                int64     `json:"transaction_type"`
	CashRegisterCode               int64     `json:"cash_register_code"`
	DocumentNumber                 int64     `json:"document_number"`
	CashierCode                    int64     `json:"cashier_code"`
	CustomerCardNumbers            string    `json:"customer_card_numbers"`
	DimensionValueCodes            string    `json:"dimension_value_codes"`
	Reserved10                     float64   `json:"reserved_10"`
	Reserved11                     float64   `json:"reserved_11"`
	Reserved12                     float64   `json:"reserved_12"`
	OperationType                  int64     `json:"operation_type"`
	ShiftNumber                    int64     `json:"shift_number"`
	CustomerCode                   float64   `json:"customer_code"`
	Reserved16                     float64   `json:"reserved_16"`
	DocumentPrintGroupCode         int64     `json:"document_print_group_code"`
	Reserved18                     string    `json:"reserved_18"`
	OrderID                        string    `json:"order_id"`
	DocumentAmountWithoutDiscounts float64   `json:"document_amount_without_discounts"`
	VisitorCount                   int64     `json:"visitor_count"`
	Reserved22                     int64     `json:"reserved_22"`
	DocumentTypeCode               int64     `json:"document_type_code"`
	CommentCode                    int64     `json:"comment_code"`
	BaseDocumentNumber             int64     `json:"base_document_number"`
	DocumentInfo                   string    `json:"document_info"`
	EnterpriseID                   int64     `json:"enterprise_id"`
	EmployeeCode                   int64     `json:"employee_code"`
	EmployeeEditDocumentNumber     int64     `json:"employee_edit_document_number"`
	DepartmentCode                 string    `json:"department_code"`
	HallCode                       int64     `json:"hall_code"`
	ServicePointCode               int64     `json:"service_point_code"`
	ReservationID                  string    `json:"reservation_id"`
	UserVariables                  string    `json:"user_variables"`
	ExternalComment                string    `json:"external_comment"`
	RevaluationDatetime            time.Time `json:"revaluation_datetime"`
	ContractorCode                 int64     `json:"contractor_code"`
	SubdivisionID                  string    `json:"subdivision_id"`
	Reserved39                     string    `json:"reserved_39"`
	DocumentCoupons                string    `json:"document_coupons"`
	Reserved44                     time.Time `json:"reserved_44"`
}

type TxVATKKT88 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	Reserved8           string    `json:"reserved_8"`
	Reserved9           string    `json:"reserved_9"`
	Reserved10          float64   `json:"reserved_10"`
	Reserved11          float64   `json:"reserved_11"`
	Reserved12          float64   `json:"reserved_12"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	Reserved15          int64     `json:"reserved_15"`
	Reserved16          int64     `json:"reserved_16"`
	PrintGroupCode      int64     `json:"print_group_code"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	Reserved21          int64     `json:"reserved_21"`
	Reserved22          int64     `json:"reserved_22"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	VAT0Amount          float64   `json:"vat_0_amount"`
	VAT10Amount         float64   `json:"vat_10_amount"`
	VAT20Amount         float64   `json:"vat_20_amount"`
	NoVATAmount         float64   `json:"no_vat_amount"`
	VAT10_110Amount     float64   `json:"vat_10_110_amount"`
	VAT20_120Amount     float64   `json:"vat_20_120_amount"`
	Reserved34          string    `json:"reserved_34"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
	Reserved37          int64     `json:"reserved_37"`
	Reserved38          string    `json:"reserved_38"`
	Reserved39          string    `json:"reserved_39"`
	Reserved43          string    `json:"reserved_43"`
}

type TxCashIn50 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	Reserved8           string    `json:"reserved_8"`
	Reserved9           string    `json:"reserved_9"`
	Reserved10          float64   `json:"reserved_10"`
	Reserved11          float64   `json:"reserved_11"`
	AmountBase          float64   `json:"amount_base"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	Reserved15          float64   `json:"reserved_15"`
	Reserved16          float64   `json:"reserved_16"`
	PrintGroupCode      int64     `json:"print_group_code"`
	Reserved18          string    `json:"reserved_18"`
	OrderID             int64     `json:"order_id"`
	Reserved20          float64   `json:"reserved_20"`
	Reserved21          int64     `json:"reserved_21"`
	Reserved22          int64     `json:"reserved_22"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	Reserved29          int64     `json:"reserved_29"`
	Reserved30          string    `json:"reserved_30"`
	Reserved31          int64     `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	Reserved33          string    `json:"reserved_33"`
	Reserved34          string    `json:"reserved_34"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxCounterChange57 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	CardOrClientCode    string    `json:"card_or_client_code"`
	CardTypeCode        string    `json:"card_type_code"`
	BindingType         float64   `json:"binding_type"`
	ValueAfterChanges   float64   `json:"value_after_changes"`
	ChangeAmount        float64   `json:"change_amount"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	PromotionCode       int64     `json:"promotion_code"`
	EventCode           int64     `json:"event_code"`
	Reserved17          int64     `json:"reserved_17"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	CounterTypeCode     int64     `json:"counter_type_code"`
	CounterCode         int64     `json:"counter_code"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	Reserved29          int64     `json:"reserved_29"`
	CounterValidFrom    string    `json:"counter_valid_from"`
	Reserved31          int64     `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	CardValidFrom       string    `json:"card_valid_from"`
	CardValidTo         string    `json:"card_valid_to"`
	CounterValidTo      string    `json:"counter_valid_to"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxReportZless60 struct {
	TransactionIDUnique           int64     `json:"transaction_id_unique"`
	SourceFolder                  string    `json:"source_folder"`
	TransactionDate               time.Time `json:"transaction_date"`
	TransactionTime               time.Time `json:"transaction_time"`
	TransactionType               int64     `json:"transaction_type"`
	CashRegisterCode              int64     `json:"cash_register_code"`
	DocumentNumber                int64     `json:"document_number"`
	CashierCode                   int64     `json:"cashier_code"`
	Reserved8                     string    `json:"reserved_8"`
	Reserved9                     string    `json:"reserved_9"`
	ShiftRevenue                  float64   `json:"shift_revenue"`
	CashInDrawer                  float64   `json:"cash_in_drawer"`
	ShiftIncomeTotal              float64   `json:"shift_income_total"`
	Reserved13                    int64     `json:"reserved_13"`
	ShiftNumber                   int64     `json:"shift_number"`
	Reserved15                    float64   `json:"reserved_15"`
	Reserved16                    float64   `json:"reserved_16"`
	PrintGroupCode                int64     `json:"print_group_code"`
	Reserved18                    string    `json:"reserved_18"`
	Reserved19                    int64     `json:"reserved_19"`
	Reserved20                    float64   `json:"reserved_20"`
	Reserved21                    int64     `json:"reserved_21"`
	Reserved22                    int64     `json:"reserved_22"`
	Reserved23                    int64     `json:"reserved_23"`
	Reserved24                    int64     `json:"reserved_24"`
	Reserved25                    string    `json:"reserved_25"`
	CashDocumentNumber            string    `json:"cash_document_number"`
	EnterpriseID                  int64     `json:"enterprise_id"`
	Reserved28                    int64     `json:"reserved_28"`
	Reserved29                    int64     `json:"reserved_29"`
	Reserved30                    string    `json:"reserved_30"`
	Reserved31                    int64     `json:"reserved_31"`
	Reserved32                    int64     `json:"reserved_32"`
	Reserved33                    string    `json:"reserved_33"`
	UnreportedDocsCount           string    `json:"unreported_docs_count"`
	ExchangeErrorCodes            string    `json:"exchange_error_codes"`
	EarliestUnreportedDocDatetime time.Time `json:"earliest_unreported_doc_datetime"`
	Reserved44                    time.Time `json:"reserved_44"`
}

type TxMarkUnit121 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	Reserved8           string    `json:"reserved_8"`
	Reserved9           string    `json:"reserved_9"`
	Reserved10          float64   `json:"reserved_10"`
	Reserved11          float64   `json:"reserved_11"`
	Reserved12          float64   `json:"reserved_12"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	Reserved15          int64     `json:"reserved_15"`
	Reserved16          int64     `json:"reserved_16"`
	PrintGroupCode      int64     `json:"print_group_code"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          string    `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	Reserved21          int64     `json:"reserved_21"`
	Reserved22          int64     `json:"reserved_22"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	Reserved26          string    `json:"reserved_26"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	Reserved29          int64     `json:"reserved_29"`
	Reserved30          string    `json:"reserved_30"`
	Reserved31          string    `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	Reserved33          string    `json:"reserved_33"`
	Reserved34          string    `json:"reserved_34"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
	Reserved39          string    `json:"reserved_39"`
	Reserved40          float64   `json:"reserved_40"`
	Reserved41          string    `json:"reserved_41"`
	Reserved42          string    `json:"reserved_42"`
	Reserved43          int64     `json:"reserved_43"`
}

type TxSpecialPrice3 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	PriceListCode       string    `json:"price_list_code"`
	Reserved9           string    `json:"reserved_9"`
	PriceType           float64   `json:"price_type"`
	SpecialPrice        float64   `json:"special_price"`
	ProductCardPrice    float64   `json:"product_card_price"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	PromotionCode       int64     `json:"promotion_code"`
	EventCode           int64     `json:"event_code"`
	PrintGroupCode      int64     `json:"print_group_code"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	Reserved21          int64     `json:"reserved_21"`
	Reserved22          int64     `json:"reserved_22"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	Reserved29          int64     `json:"reserved_29"`
	Reserved30          string    `json:"reserved_30"`
	Reserved31          int64     `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	Reserved33          string    `json:"reserved_33"`
	Reserved34          string    `json:"reserved_34"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxBonusAccrual9 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	Reserved8           string    `json:"reserved_8"`
	Reserved9           string    `json:"reserved_9"`
	BonusType           float64   `json:"bonus_type"`
	Reserved11          float64   `json:"reserved_11"`
	BonusAmount         float64   `json:"bonus_amount"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	PromotionCode       int64     `json:"promotion_code"`
	EventCode           int64     `json:"event_code"`
	PrintGroupCode      int64     `json:"print_group_code"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	CounterTypeCode     int64     `json:"counter_type_code"`
	CounterCode         int64     `json:"counter_code"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	PSProtocolNumber    int64     `json:"ps_protocol_number"`
	Reserved30          string    `json:"reserved_30"`
	Reserved31          int64     `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	ActivationDate      string    `json:"activation_date"`
	ExpirationDate      string    `json:"expiration_date"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxPositionDiscount15 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	DiscountInfo        string    `json:"discount_info"`
	Reserved9           string    `json:"reserved_9"`
	DiscountType        float64   `json:"discount_type"`
	DiscountValue       float64   `json:"discount_value"`
	DiscountAmountBase  float64   `json:"discount_amount_base"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	PromotionCode       int64     `json:"promotion_code"`
	EventCode           int64     `json:"event_code"`
	PrintGroupCode      int64     `json:"print_group_code"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	Reserved21          int64     `json:"reserved_21"`
	Reserved22          int64     `json:"reserved_22"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	Reserved29          int64     `json:"reserved_29"`
	Reserved30          string    `json:"reserved_30"`
	Reserved31          int64     `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	Reserved33          string    `json:"reserved_33"`
	Reserved34          string    `json:"reserved_34"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxBillRegistration21_23 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	BillCode            string    `json:"bill_code"`
	Reserved9           string    `json:"reserved_9"`
	BillDenomination    float64   `json:"bill_denomination"`
	BillQuantity        float64   `json:"bill_quantity"`
	BillAmountBase      float64   `json:"bill_amount_base"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	Reserved15          float64   `json:"reserved_15"`
	Reserved16          float64   `json:"reserved_16"`
	Reserved17          int64     `json:"reserved_17"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	Reserved21          int64     `json:"reserved_21"`
	Reserved22          int64     `json:"reserved_22"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	Reserved29          int64     `json:"reserved_29"`
	Reserved30          string    `json:"reserved_30"`
	Reserved31          int64     `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	Reserved33          string    `json:"reserved_33"`
	Reserved34          string    `json:"reserved_34"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxEmployeeRegistration25 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	EmployeeCode        string    `json:"employee_code"`
	Reserved9           string    `json:"reserved_9"`
	Reserved10          float64   `json:"reserved_10"`
	Reserved11          float64   `json:"reserved_11"`
	Reserved12          float64   `json:"reserved_12"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	Reserved15          float64   `json:"reserved_15"`
	Reserved16          float64   `json:"reserved_16"`
	Reserved17          int64     `json:"reserved_17"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	Reserved21          int64     `json:"reserved_21"`
	Reserved22          int64     `json:"reserved_22"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	Reserved29          int64     `json:"reserved_29"`
	Reserved30          string    `json:"reserved_30"`
	Reserved31          int64     `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	Reserved33          string    `json:"reserved_33"`
	Reserved34          string    `json:"reserved_34"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxEmployeeAccountingDoc26 struct {
	TransactionIDUnique    int64     `json:"transaction_id_unique"`
	SourceFolder           string    `json:"source_folder"`
	TransactionDate        time.Time `json:"transaction_date"`
	TransactionTime        time.Time `json:"transaction_time"`
	TransactionType        int64     `json:"transaction_type"`
	CashRegisterCode       int64     `json:"cash_register_code"`
	DocumentNumber         int64     `json:"document_number"`
	CashierCode            int64     `json:"cashier_code"`
	EmployeeCode           string    `json:"employee_code"`
	Reserved9              string    `json:"reserved_9"`
	Reserved10             float64   `json:"reserved_10"`
	Reserved11             float64   `json:"reserved_11"`
	Reserved12             float64   `json:"reserved_12"`
	OperationType          int64     `json:"operation_type"`
	ShiftNumber            int64     `json:"shift_number"`
	Reserved15             float64   `json:"reserved_15"`
	Reserved16             float64   `json:"reserved_16"`
	DocumentPrintGroupCode int64     `json:"document_print_group_code"`
	Reserved18             string    `json:"reserved_18"`
	Reserved19             int64     `json:"reserved_19"`
	Reserved20             float64   `json:"reserved_20"`
	Reserved21             int64     `json:"reserved_21"`
	Reserved22             int64     `json:"reserved_22"`
	DocumentTypeCode       int64     `json:"document_type_code"`
	Reserved24             int64     `json:"reserved_24"`
	Reserved25             int64     `json:"reserved_25"`
	DocumentInfo           string    `json:"document_info"`
	EnterpriseID           int64     `json:"enterprise_id"`
	Reserved28             int64     `json:"reserved_28"`
	Reserved29             int64     `json:"reserved_29"`
	Reserved30             string    `json:"reserved_30"`
	Reserved31             int64     `json:"reserved_31"`
	Reserved32             int64     `json:"reserved_32"`
	Reserved33             string    `json:"reserved_33"`
	Reserved34             string    `json:"reserved_34"`
	Reserved35             string    `json:"reserved_35"`
	Reserved36             time.Time `json:"reserved_36"`
}

type TxCardStatusChange27 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	CardNumber          string    `json:"card_number"`
	CardTypeCode        string    `json:"card_type_code"`
	CardType            float64   `json:"card_type"`
	Reserved11          float64   `json:"reserved_11"`
	Reserved12          float64   `json:"reserved_12"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	PromotionCode       float64   `json:"promotion_code"`
	EventCode           float64   `json:"event_code"`
	Reserved17          int64     `json:"reserved_17"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	Reserved21          int64     `json:"reserved_21"`
	Reserved22          int64     `json:"reserved_22"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	Reserved29          int64     `json:"reserved_29"`
	Reserved30          string    `json:"reserved_30"`
	OldCardStatus       int64     `json:"old_card_status"`
	NewCardStatus       int64     `json:"new_card_status"`
	NewValidFrom        string    `json:"new_valid_from"`
	NewValidTo          string    `json:"new_valid_to"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxModifierRegistration30 struct {
	TransactionIDUnique    int64     `json:"transaction_id_unique"`
	SourceFolder           string    `json:"source_folder"`
	TransactionDate        time.Time `json:"transaction_date"`
	TransactionTime        time.Time `json:"transaction_time"`
	TransactionType        int64     `json:"transaction_type"`
	CashRegisterCode       int64     `json:"cash_register_code"`
	DocumentNumber         int64     `json:"document_number"`
	CashierCode            int64     `json:"cashier_code"`
	ItemIdentifier         string    `json:"item_identifier"`
	Reserved9              string    `json:"reserved_9"`
	Reserved10             float64   `json:"reserved_10"`
	ItemQuantity           float64   `json:"item_quantity"`
	Reserved12             float64   `json:"reserved_12"`
	OperationType          int64     `json:"operation_type"`
	ShiftNumber            int64     `json:"shift_number"`
	Reserved15             float64   `json:"reserved_15"`
	Reserved16             float64   `json:"reserved_16"`
	DocumentPrintGroupCode int64     `json:"document_print_group_code"`
	Reserved18             string    `json:"reserved_18"`
	Reserved19             int64     `json:"reserved_19"`
	Reserved20             float64   `json:"reserved_20"`
	Reserved21             int64     `json:"reserved_21"`
	Reserved22             int64     `json:"reserved_22"`
	DocumentTypeCode       int64     `json:"document_type_code"`
	Reserved24             int64     `json:"reserved_24"`
	Reserved25             int64     `json:"reserved_25"`
	DocumentInfo           string    `json:"document_info"`
	EnterpriseID           int64     `json:"enterprise_id"`
	Reserved28             int64     `json:"reserved_28"`
	Reserved29             int64     `json:"reserved_29"`
	Reserved30             string    `json:"reserved_30"`
	Reserved31             int64     `json:"reserved_31"`
	Reserved32             int64     `json:"reserved_32"`
	Reserved33             string    `json:"reserved_33"`
	Reserved34             string    `json:"reserved_34"`
	ModifierCode           string    `json:"modifier_code"`
	Reserved36             time.Time `json:"reserved_36"`
}

type TxBonusPayment32 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	BonusCardNumber     string    `json:"bonus_card_number"`
	Reserved9           string    `json:"reserved_9"`
	BonusPaymentType    float64   `json:"bonus_payment_type"`
	CounterChangeAmount float64   `json:"counter_change_amount"`
	PaymentAmount       float64   `json:"payment_amount"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	PromotionCode       int64     `json:"promotion_code"`
	EventCode           int64     `json:"event_code"`
	PrintGroupCode      int64     `json:"print_group_code"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	CounterTypeCode     int64     `json:"counter_type_code"`
	CounterCode         int64     `json:"counter_code"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	PSProtocolNumber    int64     `json:"ps_protocol_number"`
	Reserved30          string    `json:"reserved_30"`
	Reserved31          int64     `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	Reserved33          string    `json:"reserved_33"`
	Reserved34          string    `json:"reserved_34"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxPrepayment34 struct {
	TransactionIDUnique int64     `json:"transaction_id_unique"`
	SourceFolder        string    `json:"source_folder"`
	TransactionDate     time.Time `json:"transaction_date"`
	TransactionTime     time.Time `json:"transaction_time"`
	TransactionType     int64     `json:"transaction_type"`
	CashRegisterCode    int64     `json:"cash_register_code"`
	DocumentNumber      int64     `json:"document_number"`
	CashierCode         int64     `json:"cashier_code"`
	Reserved8           string    `json:"reserved_8"`
	Reserved9           string    `json:"reserved_9"`
	PrepaymentType      float64   `json:"prepayment_type"`
	Reserved11          float64   `json:"reserved_11"`
	PrepaymentAmount    float64   `json:"prepayment_amount"`
	OperationType       int64     `json:"operation_type"`
	ShiftNumber         int64     `json:"shift_number"`
	Reserved15          float64   `json:"reserved_15"`
	Reserved16          float64   `json:"reserved_16"`
	PrintGroupCode      int64     `json:"print_group_code"`
	Reserved18          string    `json:"reserved_18"`
	Reserved19          int64     `json:"reserved_19"`
	Reserved20          float64   `json:"reserved_20"`
	Reserved21          int64     `json:"reserved_21"`
	Reserved22          int64     `json:"reserved_22"`
	DocumentTypeCode    int64     `json:"document_type_code"`
	Reserved24          int64     `json:"reserved_24"`
	Reserved25          int64     `json:"reserved_25"`
	DocumentInfo        string    `json:"document_info"`
	EnterpriseID        int64     `json:"enterprise_id"`
	Reserved28          int64     `json:"reserved_28"`
	Reserved29          int64     `json:"reserved_29"`
	Reserved30          string    `json:"reserved_30"`
	Reserved31          int64     `json:"reserved_31"`
	Reserved32          int64     `json:"reserved_32"`
	Reserved33          string    `json:"reserved_33"`
	Reserved34          string    `json:"reserved_34"`
	Reserved35          string    `json:"reserved_35"`
	Reserved36          time.Time `json:"reserved_36"`
}

type TxItemStorno2_12 = TxItemRegistration1_11
type TxDocumentDiscount37 = TxDocumentDiscount35
type TxDocumentDiscount85 = TxDocumentDiscount35
type TxDocumentDiscount87 = TxDocumentDiscount35
type TxDocumentRounding38 = TxDocumentDiscount35
type TxNonFiscalPayment86 = TxNonFiscalPayment36
type TxFiscalPayment43 = TxFiscalPayment40
type TxDocumentClose55 = TxDocumentOpen42
type TxDocumentCancel56 = TxDocumentOpen42
type TxDocumentNonFinClose58 = TxDocumentOpen42
type TxDocumentClients65 = TxDocumentOpen42
type TxDocumentCloseKKT45 = TxDocumentOpen42
type TxDocumentCloseGp49 = TxDocumentOpen42
type TxDocumentEGAIS120 = TxDocumentOpen42
type TxCashOut51 = TxCashIn50
type TxReportZ63 = TxReportZless60
type TxShiftOpenDoc64 = TxReportZless60
type TxShiftClose61 = TxReportZless60
type TxShiftOpen62 = TxReportZless60
type TxBonusRefund10 = TxBonusAccrual9
type TxPositionDiscount17 = TxPositionDiscount15
type TxBillStorno22_24 = TxBillRegistration21_23
type TxEmployeeAccountingPos29 = TxEmployeeAccountingDoc26
type TxModifierStorno31 = TxModifierRegistration30
type TxBonusPayment82 = TxBonusPayment32
type TxBonusPayment33 = TxBonusPayment32
type TxBonusPayment83 = TxBonusPayment32
type TxPrepayment84 = TxPrepayment34
