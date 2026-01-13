# DDL спецификация (на основе TRANSACTION_TABLES_SPEC.md)

Назначение: зафиксировать целевую структуру таблиц транзакций с типами полей.  
Базируется на `docs/TRANSACTION_TABLES_SPEC.md`.

## Общие правила

- Во всех таблицах транзакций присутствуют:
  - `transaction_id_unique` BIGINT
  - `source_folder` TEXT
  - `transaction_date` DATE
  - `transaction_time` TIME
- PK: `(transaction_id_unique, source_folder)`
- Индексы:
  - `(<table>_date_idx)` на `transaction_date`
  - `(<table>_source_idx)` на `source_folder`
- Типы полей:
  - `Целое` → BIGINT (если не указанно иное)
  - `Дробное` → NUMERIC(18,6)
  - `Дата` → DATE
  - `Время` → TIME
  - `Дата и время` → DATE (ослаблено по данным)
  - `Строка` → TEXT

---

## Таблицы транзакций

### tx_item_registration_1_11
```sql
CREATE TABLE tx_item_registration_1_11 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  item_identifier TEXT,
  dimension_value_codes TEXT,
  price_without_discounts NUMERIC(18,6),
  quantity NUMERIC(18,6),
  position_amount_with_rounding NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  final_price NUMERIC(18,6),
  position_total_amount NUMERIC(18,6),
  print_group_code BIGINT,
  article_sku TEXT,
  registration_barcode TEXT,
  position_amount_base NUMERIC(18,6),
  kkt_section BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  comment_code BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  employee_code BIGINT,
  split_pack_quantity BIGINT,
  gift_card_external_number TEXT,
  pack_quantity BIGINT,
  item_type_code BIGINT,
  marking_code TEXT,
  excise_marks TEXT,
  personal_modifier_group_code TEXT,
  stoloto_registration_time DATE,
  stoloto_ticket_id BIGINT,
  reserved_38 BIGINT,
  alc_code TEXT,
  reserved_40 NUMERIC(18,6),
  prescription_data_1 TEXT,
  prescription_data_2 TEXT,
  position_coupons TEXT,
  reserved_44 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_item_storno_2_12
Колонки как в `tx_item_registration_1_11`.

### tx_item_tax_4_14
```sql
CREATE TABLE tx_item_tax_4_14 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  reserved_8 TEXT,
  dimension_value_codes TEXT,
  tax_group_code BIGINT,
  tax_rate_code BIGINT,
  tax_amount_base NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  reserved_15 NUMERIC(18,6),
  total_amount_base_with_discounts NUMERIC(18,6),
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  amount_base_without_discounts NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_item_kkt_6_16
```sql
CREATE TABLE tx_item_kkt_6_16 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  item_identifier TEXT,
  dimension_value_codes TEXT,
  reserved_10 NUMERIC(18,6),
  quantity_kkt NUMERIC(18,6),
  reserved_12 NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  final_price_kkt_currency NUMERIC(18,6),
  position_total_amount_kkt_currency NUMERIC(18,6),
  print_group_code BIGINT,
  article_sku TEXT,
  registration_barcode TEXT,
  reserved_20 NUMERIC(18,6),
  kkt_section BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  comment_code BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  employee_code BIGINT,
  split_pack_quantity BIGINT,
  gift_card_external_number TEXT,
  pack_quantity BIGINT,
  item_type_code BIGINT,
  marking_code TEXT,
  excise_marks TEXT,
  personal_modifier_group_code TEXT,
  stoloto_registration_time DATE,
  stoloto_ticket_id BIGINT,
  reserved_38 BIGINT,
  alc_code TEXT,
  reserved_40 NUMERIC(18,6),
  prescription_data_1 TEXT,
  prescription_data_2 TEXT,
  position_coupons TEXT,
  reserved_44 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_document_discount_35
```sql
CREATE TABLE tx_document_discount_35 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  discount_info TEXT,
  reserved_9 TEXT,
  discount_type NUMERIC(18,6),
  discount_value NUMERIC(18,6),
  discount_amount_base NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  promotion_code BIGINT,
  event_code BIGINT,
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_document_discount_37
Колонки как в `tx_document_discount_35`.

### tx_document_discount_85
Колонки как в `tx_document_discount_35`.

### tx_document_discount_87
Колонки как в `tx_document_discount_35`.

### tx_document_rounding_38
Колонки как в `tx_document_discount_35`.

### tx_non_fiscal_payment_36
```sql
CREATE TABLE tx_non_fiscal_payment_36 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  gift_card_number TEXT,
  payment_type_code TEXT,
  payment_type_operation NUMERIC(18,6),
  reserved_11 NUMERIC(18,6),
  payment_amount NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  promotion_code BIGINT,
  event_code BIGINT,
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  counter_type_code BIGINT,
  counter_code BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_non_fiscal_payment_86
Колонки как в `tx_non_fiscal_payment_36`.

### tx_fiscal_payment_40
```sql
CREATE TABLE tx_fiscal_payment_40 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  card_number TEXT,
  payment_type_code TEXT,
  payment_type_operation NUMERIC(18,6),
  customer_amount_payment_currency NUMERIC(18,6),
  customer_amount_base_currency NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  promotion_code BIGINT,
  event_code BIGINT,
  current_print_group_code BIGINT,
  reserved_18 TEXT,
  currency_code BIGINT,
  cash_out_amount NUMERIC(18,6),
  counter_type_code BIGINT,
  counter_code BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  ps_protocol_number BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_fiscal_payment_43
Колонки как в `tx_fiscal_payment_40`.

### tx_document_open_42
```sql
CREATE TABLE tx_document_open_42 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  customer_card_numbers TEXT,
  dimension_value_codes TEXT,
  reserved_10 NUMERIC(18,6),
  reserved_11 NUMERIC(18,6),
  reserved_12 NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  customer_code NUMERIC(18,6),
  reserved_16 NUMERIC(18,6),
  document_print_group_code BIGINT,
  reserved_18 TEXT,
  order_id TEXT,
  document_amount_without_discounts NUMERIC(18,6),
  visitor_count BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  comment_code BIGINT,
  base_document_number BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  employee_code BIGINT,
  employee_edit_document_number BIGINT,
  department_code TEXT,
  hall_code BIGINT,
  service_point_code BIGINT,
  reservation_id TEXT,
  user_variables TEXT,
  external_comment TEXT,
  revaluation_datetime DATE,
  contractor_code BIGINT,
  subdivision_id TEXT,
  reserved_39 TEXT,
  document_coupons TEXT,
  reserved_44 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_document_close_55
Колонки как в `tx_document_open_42`.

### tx_document_cancel_56
Колонки как в `tx_document_open_42`.

### tx_document_non_fin_close_58
Колонки как в `tx_document_open_42`.

### tx_document_clients_65
Колонки как в `tx_document_open_42`.

### tx_document_close_kkt_45
Колонки как в `tx_document_open_42`.

### tx_document_close_gp_49
Колонки как в `tx_document_open_42`.

### tx_document_egais_120
Колонки как в `tx_document_open_42`.

### tx_vat_kkt_88
```sql
CREATE TABLE tx_vat_kkt_88 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  reserved_8 TEXT,
  reserved_9 TEXT,
  reserved_10 NUMERIC(18,6),
  reserved_11 NUMERIC(18,6),
  reserved_12 NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  reserved_15 BIGINT,
  reserved_16 BIGINT,
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  vat_0_amount NUMERIC(18,6),
  vat_10_amount NUMERIC(18,6),
  vat_20_amount NUMERIC(18,6),
  no_vat_amount NUMERIC(18,6),
  vat_10_110_amount NUMERIC(18,6),
  vat_20_120_amount NUMERIC(18,6),
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  reserved_37 BIGINT,
  reserved_38 TEXT,
  reserved_39 TEXT,
  reserved_43 TEXT,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_cash_in_50
```sql
CREATE TABLE tx_cash_in_50 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  reserved_8 TEXT,
  reserved_9 TEXT,
  reserved_10 NUMERIC(18,6),
  reserved_11 NUMERIC(18,6),
  amount_base NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  reserved_15 NUMERIC(18,6),
  reserved_16 NUMERIC(18,6),
  print_group_code BIGINT,
  reserved_18 TEXT,
  order_id BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_cash_out_51
Колонки как в `tx_cash_in_50`.

### tx_counter_change_57
```sql
CREATE TABLE tx_counter_change_57 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  card_or_client_code TEXT,
  card_type_code TEXT,
  binding_type NUMERIC(18,6),
  value_after_changes NUMERIC(18,6),
  change_amount NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  promotion_code BIGINT,
  event_code BIGINT,
  reserved_17 BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  counter_type_code BIGINT,
  counter_code BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  counter_valid_from TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  card_valid_from TEXT,
  card_valid_to TEXT,
  counter_valid_to TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_report_zless_60
```sql
CREATE TABLE tx_report_zless_60 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  reserved_8 TEXT,
  reserved_9 TEXT,
  shift_revenue NUMERIC(18,6),
  cash_in_drawer NUMERIC(18,6),
  shift_income_total NUMERIC(18,6),
  reserved_13 BIGINT,
  shift_number BIGINT,
  reserved_15 NUMERIC(18,6),
  reserved_16 NUMERIC(18,6),
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  reserved_23 BIGINT,
  reserved_24 BIGINT,
  reserved_25 TEXT,
  cash_document_number TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  unreported_docs_count TEXT,
  exchange_error_codes TEXT,
  earliest_unreported_doc_datetime DATE,
  reserved_44 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_report_z_63
Колонки как в `tx_report_zless_60`.

### tx_shift_open_doc_64
Колонки как в `tx_report_zless_60`.

### tx_shift_close_61
Колонки как в `tx_report_zless_60`.

### tx_shift_open_62
Колонки как в `tx_report_zless_60`.

### tx_mark_unit_121
```sql
CREATE TABLE tx_mark_unit_121 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  reserved_8 TEXT,
  reserved_9 TEXT,
  reserved_10 NUMERIC(18,6),
  reserved_11 NUMERIC(18,6),
  reserved_12 NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  reserved_15 BIGINT,
  reserved_16 BIGINT,
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 TEXT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  reserved_26 TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 TEXT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  reserved_39 TEXT,
  reserved_40 NUMERIC(18,6),
  reserved_41 TEXT,
  reserved_42 TEXT,
  reserved_43 BIGINT,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_special_price_3
```sql
CREATE TABLE tx_special_price_3 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  price_list_code TEXT,
  reserved_9 TEXT,
  price_type NUMERIC(18,6),
  special_price NUMERIC(18,6),
  product_card_price NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  promotion_code BIGINT,
  event_code BIGINT,
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_bonus_accrual_9
```sql
CREATE TABLE tx_bonus_accrual_9 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  reserved_8 TEXT,
  reserved_9 TEXT,
  bonus_type NUMERIC(18,6),
  reserved_11 NUMERIC(18,6),
  bonus_amount NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  promotion_code BIGINT,
  event_code BIGINT,
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  counter_type_code BIGINT,
  counter_code BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  ps_protocol_number BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  activation_date TEXT,
  expiration_date TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_bonus_refund_10
Колонки как в `tx_bonus_accrual_9`.

### tx_position_discount_15
```sql
CREATE TABLE tx_position_discount_15 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  discount_info TEXT,
  reserved_9 TEXT,
  discount_type NUMERIC(18,6),
  discount_value NUMERIC(18,6),
  discount_amount_base NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  promotion_code BIGINT,
  event_code BIGINT,
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_position_discount_17
Колонки как в `tx_position_discount_15`.

### tx_bill_registration_21_23
```sql
CREATE TABLE tx_bill_registration_21_23 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  bill_code TEXT,
  reserved_9 TEXT,
  bill_denomination NUMERIC(18,6),
  bill_quantity NUMERIC(18,6),
  bill_amount_base NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  reserved_15 NUMERIC(18,6),
  reserved_16 NUMERIC(18,6),
  reserved_17 BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_bill_storno_22_24
Колонки как в `tx_bill_registration_21_23`.

### tx_employee_registration_25
```sql
CREATE TABLE tx_employee_registration_25 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  employee_code TEXT,
  reserved_9 TEXT,
  reserved_10 NUMERIC(18,6),
  reserved_11 NUMERIC(18,6),
  reserved_12 NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  reserved_15 NUMERIC(18,6),
  reserved_16 NUMERIC(18,6),
  reserved_17 BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_employee_accounting_doc_26
```sql
CREATE TABLE tx_employee_accounting_doc_26 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  employee_code TEXT,
  reserved_9 TEXT,
  reserved_10 NUMERIC(18,6),
  reserved_11 NUMERIC(18,6),
  reserved_12 NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  reserved_15 NUMERIC(18,6),
  reserved_16 NUMERIC(18,6),
  document_print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_employee_accounting_pos_29
Колонки как в `tx_employee_accounting_doc_26`.

### tx_card_status_change_27
```sql
CREATE TABLE tx_card_status_change_27 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  card_number TEXT,
  card_type_code TEXT,
  card_type NUMERIC(18,6),
  reserved_11 NUMERIC(18,6),
  reserved_12 NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  promotion_code NUMERIC(18,6),
  event_code NUMERIC(18,6),
  reserved_17 BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  old_card_status BIGINT,
  new_card_status BIGINT,
  new_valid_from TEXT,
  new_valid_to TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_modifier_registration_30
```sql
CREATE TABLE tx_modifier_registration_30 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  item_identifier TEXT,
  reserved_9 TEXT,
  reserved_10 NUMERIC(18,6),
  item_quantity NUMERIC(18,6),
  reserved_12 NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  reserved_15 NUMERIC(18,6),
  reserved_16 NUMERIC(18,6),
  document_print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  modifier_code TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_modifier_storno_31
Колонки как в `tx_modifier_registration_30`.

### tx_bonus_payment_32
```sql
CREATE TABLE tx_bonus_payment_32 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  bonus_card_number TEXT,
  reserved_9 TEXT,
  bonus_payment_type NUMERIC(18,6),
  counter_change_amount NUMERIC(18,6),
  payment_amount NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  promotion_code BIGINT,
  event_code BIGINT,
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  counter_type_code BIGINT,
  counter_code BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  ps_protocol_number BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_bonus_payment_82
Колонки как в `tx_bonus_payment_32`.

### tx_bonus_payment_33
Колонки как в `tx_bonus_payment_32`.

### tx_bonus_payment_83
Колонки как в `tx_bonus_payment_32`.

### tx_prepayment_34
```sql
CREATE TABLE tx_prepayment_34 (
  transaction_id_unique BIGINT NOT NULL,
  source_folder TEXT NOT NULL,
  transaction_date DATE,
  transaction_time TIME,
  transaction_type BIGINT,
  cash_register_code BIGINT,
  document_number BIGINT,
  cashier_code BIGINT,
  reserved_8 TEXT,
  reserved_9 TEXT,
  prepayment_type NUMERIC(18,6),
  reserved_11 NUMERIC(18,6),
  prepayment_amount NUMERIC(18,6),
  operation_type BIGINT,
  shift_number BIGINT,
  reserved_15 NUMERIC(18,6),
  reserved_16 NUMERIC(18,6),
  print_group_code BIGINT,
  reserved_18 TEXT,
  reserved_19 BIGINT,
  reserved_20 NUMERIC(18,6),
  reserved_21 BIGINT,
  reserved_22 BIGINT,
  document_type_code BIGINT,
  reserved_24 BIGINT,
  reserved_25 BIGINT,
  document_info TEXT,
  enterprise_id BIGINT,
  reserved_28 BIGINT,
  reserved_29 BIGINT,
  reserved_30 TEXT,
  reserved_31 BIGINT,
  reserved_32 BIGINT,
  reserved_33 TEXT,
  reserved_34 TEXT,
  reserved_35 TEXT,
  reserved_36 DATE,
  PRIMARY KEY (transaction_id_unique, source_folder)
);
```

### tx_prepayment_84
Колонки как в `tx_prepayment_34`.
