# Unified transaction tables specification

Назначение: объединить DDL-спецификацию и пометку полей выгрузки Frontol 6 в одном файле.
Этот документ применим к миграциям и используется как единая точка истины по структуре таблиц транзакций.

Источники:
- `docs/DDL_SPEC.md`
- `docs/TRANSACTION_TABLES_SPEC.md`

---

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


---

# Документация по таблицам транзакций (поля по документации)

Назначение: дать проверяемое соответствие полей из выгрузки Frontol 6 колонкам БД.  
Для каждого типа транзакции приведен отдельный "виртуальный" набор колонок.  
Если в реальной БД типы объединены, в документации они разделены, чтобы проверять корректность заполнения.

Правила:
- Все таблицы имеют `transaction_date`, `transaction_time`, `source_folder` для сортировки по дате и месту загрузки.
- Типы полей соответствуют документации (`Целое` → INT/BIGINT, `Дробное` → NUMERIC, `Дата` → DATE, `Время` → TIME, `Дата и время` → TIMESTAMP, `Строка` → TEXT).
- Номер телефона/карты всегда TEXT.
- Если поле не описано в документации, оно именуется `reserved_<N>`.

---

## 1/11 Регистрация товара (по свободной цене / из справочника)

Таблица документации: `tx_item_registration_1_11`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 1/11 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | item_identifier | TEXT | Идентификатор товара |
| 9 | dimension_value_codes | TEXT | Коды значений разрезов |
| 10 | price_without_discounts | NUMERIC | Цена без скидок |
| 11 | quantity | NUMERIC | Количество товара |
| 12 | position_amount_with_rounding | NUMERIC | Сумма позиции + сумма округления |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | final_price | NUMERIC | Итоговая цена со скидками |
| 16 | position_total_amount | NUMERIC | Итоговая сумма позиции (см. примечания) |
| 17 | print_group_code | INT | Код группы печати |
| 18 | article_sku | TEXT | Артикул товара |
| 19 | registration_barcode | TEXT | Штрихкод регистрации |
| 20 | position_amount_base | NUMERIC | Цена/спеццена × количество |
| 21 | kkt_section | INT | Секция ККМ |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | comment_code | INT | Код комментария |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | employee_code | INT | Код сотрудника |
| 29 | split_pack_quantity | INT | Количество позиции для деленной упаковки |
| 30 | gift_card_external_number | TEXT | Номер активированной/деактивированной подарочной карты |
| 31 | pack_quantity | INT | Количество в упаковке |
| 32 | item_type_code | INT | Тип номенклатуры |
| 33 | marking_code | TEXT | SGTIN / УИН / КиЗ и т.д. |
| 34 | excise_marks | TEXT | Акцизные марки |
| 35 | personal_modifier_group_code | TEXT | Код группы персональных модификаторов |
| 36 | stoloto_registration_time | DATE | Время регистрации лотерейного билета |
| 37 | stoloto_ticket_id | INT | Идентификатор лотерейного билета |
| 38 | reserved_38 | INT | – |
| 39 | alc_code | TEXT | AlcCode |
| 40 | reserved_40 | NUMERIC | – |
| 41 | prescription_data_1 | TEXT | Данные рецепта |
| 42 | prescription_data_2 | TEXT | Данные рецепта |
| 43 | position_coupons | TEXT | Купоны на позицию |
| 44 | reserved_44 | DATE | – |

Примечания: см. раздел "Регистрация товара" + "Особенности полей" в `docs/frontol_6_integration.md`.

Особенности (по PDF):
- поле №8: код или артикул в зависимости от настройки идентификатора товара; если строковый идентификатор — поле №18 пустое.
- поле №19: штрихкод регистрации (зависит от сценария запроса ШК/коэффициента).
- поле №10: при спеццене/прайс-листе выгружается эта цена (может отличаться от транзакции 3, поле №11).
- поля №29 и №31 заполняются только для транзакций 11/12/16.
- поле №30 (для 1/11): номер внешней подарочной карты.
- поле №33 для маркированных товаров может содержать SGTIN/УИН/КиЗ; символ `;` заменяется на `¤`.

---

## 2/12 Сторно товара (по свободной цене / из справочника)

Таблица документации: `tx_item_storno_2_12`  
Колонки: как в `tx_item_registration_1_11`.

Различия по смыслу:
- поля №10, №12, №15, №16, №20 передаются со знаками, обратными регистрации.
- применяются правила знаков из раздела "Общие особенности транзакций".

---

## 4/14 Налог на товар (по свободной цене / из справочника)

Таблица документации: `tx_item_tax_4_14`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 4/14 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | reserved_8 | TEXT | – |
| 9 | dimension_value_codes | TEXT | Коды значений разрезов |
| 10 | tax_group_code | INT | Код налоговой группы |
| 11 | tax_rate_code | INT | Код налоговой ставки |
| 12 | tax_amount_base | NUMERIC | Сумма налога в базовой валюте |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | reserved_15 | NUMERIC | – |
| 16 | total_amount_base_with_discounts | NUMERIC | Итоговая сумма в базовой валюте со скидками |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | amount_base_without_discounts | NUMERIC | Сумма в базовой валюте без скидок |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

---

## 6/16 Регистрация товара в ККТ (по свободной цене / из справочника)

Таблица документации: `tx_item_kkt_6_16`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 6/16 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | item_identifier | TEXT | Идентификатор товара |
| 9 | dimension_value_codes | TEXT | Коды значений разрезов |
| 10 | reserved_10 | NUMERIC | – |
| 11 | quantity_kkt | NUMERIC | Количество, регистрируемое в ККТ |
| 12 | reserved_12 | NUMERIC | – |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | final_price_kkt_currency | NUMERIC | Итоговая цена со скидками в валюте ККТ |
| 16 | position_total_amount_kkt_currency | NUMERIC | Итоговая сумма позиции в валюте ККТ |
| 17 | print_group_code | INT | Код группы печати |
| 18 | article_sku | TEXT | Артикул товара |
| 19 | registration_barcode | TEXT | Штрихкод регистрации |
| 20 | reserved_20 | NUMERIC | – |
| 21 | kkt_section | INT | Секция ККМ |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | comment_code | INT | Код комментария |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | employee_code | INT | Код сотрудника |
| 29 | split_pack_quantity | INT | Количество позиции для деленной упаковки |
| 30 | gift_card_external_number | TEXT | Номер активированной/деактивированной подарочной карты |
| 31 | pack_quantity | INT | Количество в упаковке |
| 32 | item_type_code | INT | Тип номенклатуры |
| 33 | marking_code | TEXT | SGTIN / УИН / КиЗ и т.д. |
| 34 | excise_marks | TEXT | Акцизные марки |
| 35 | personal_modifier_group_code | TEXT | Код группы персональных модификаторов |
| 36 | stoloto_registration_time | DATE | Время регистрации лотерейного билета |
| 37 | stoloto_ticket_id | INT | Идентификатор лотерейного билета |
| 38 | reserved_38 | INT | – |
| 39 | alc_code | TEXT | AlcCode |
| 40 | reserved_40 | NUMERIC | – |
| 41 | prescription_data_1 | TEXT | Данные рецепта |
| 42 | prescription_data_2 | TEXT | Данные рецепта |
| 43 | position_coupons | TEXT | Купоны на позицию |
| 44 | reserved_44 | DATE | – |

---

## 35 Скидка суммой на документ

Таблица документации: `tx_document_discount_35`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 35 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | discount_info | TEXT | Информация по скидке |
| 9 | reserved_9 | TEXT | – |
| 10 | discount_type | NUMERIC | Тип скидки |
| 11 | discount_value | NUMERIC | Значение скидки |
| 12 | discount_amount_base | NUMERIC | Сумма скидки в базовой валюте |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | promotion_code | INT | Код акции |
| 16 | event_code | INT | Код мероприятия |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

Особенности (по PDF):
- поле №8: дисконтная карта/классификатор, если сработало условие, иначе пусто.
- поле №10: тип скидки (0,1,2,3,6,10,11).
- поля №11 и №12: передаются со знаком (отрицательные скидки/возвраты).
- для 37/87 при рецептурных товарах в поле №8 передаются реквизиты рецепта, а в №15/16 — акция/мероприятие субсидии.
- перевод строки в поле №8 заменяется на служебный символ `#166`.

## 37 Скидка % на документ

Таблица документации: `tx_document_discount_37`  
Колонки: как в `tx_document_discount_35`.

## 85 Скидка суммой на документ, распределенная по позициям

Таблица документации: `tx_document_discount_85`  
Колонки: как в `tx_document_discount_35` (сумма в поле №12 — распределение по позициям).

## 87 Скидка % на документ, распределенная по позициям

Таблица документации: `tx_document_discount_87`  
Колонки: как в `tx_document_discount_35` (сумма в поле №12 — распределение по позициям).

## 38 Округление чека к расчету

Таблица документации: `tx_document_rounding_38`  
Колонки: как в `tx_document_discount_35`, но поле №12 = сумма округления, а поля №8, №10, №11, №15, №16 могут быть пустыми.

---

## 36 Нефискальная оплата

Таблица документации: `tx_non_fiscal_payment_36`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 36 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | gift_card_number | TEXT | Номер подарочной карты |
| 9 | payment_type_code | TEXT | Код вида оплаты |
| 10 | payment_type_operation | NUMERIC | Операция вида оплаты |
| 11 | reserved_11 | NUMERIC | – |
| 12 | payment_amount | NUMERIC | Сумма оплаты |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | promotion_code | INT | Код акции |
| 16 | event_code | INT | Код мероприятия |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | counter_type_code | INT | Код вида счетчика |
| 22 | counter_code | INT | Код счетчика |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

Особенности (по PDF):
- поля №15 и №16 заполняются только при оплате подарочными картами.
- поле №10: код операции вида оплаты (6 — внутренняя подарочная карта, 8 — внешняя).

## 86 Нефискальная оплата, распределенная по позициям

Таблица документации: `tx_non_fiscal_payment_86`  
Колонки: как в `tx_non_fiscal_payment_36`, но поле №12 = распределение суммы по позициям, поле №17 = код группы печати позиции.

---

## 40 Фискальная оплата

Таблица документации: `tx_fiscal_payment_40`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 40 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | card_number | TEXT | Номер карты |
| 9 | payment_type_code | TEXT | Код вида оплаты |
| 10 | payment_type_operation | NUMERIC | Операция вида оплаты |
| 11 | customer_amount_payment_currency | NUMERIC | Сумма клиента в валюте оплаты |
| 12 | customer_amount_base_currency | NUMERIC | Сумма клиента в базовой валюте |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | promotion_code | INT | Код акции |
| 16 | event_code | INT | Код мероприятия |
| 17 | current_print_group_code | INT | Код текущей группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | currency_code | INT | Код валюты |
| 20 | cash_out_amount | NUMERIC | Сумма выдачи наличных |
| 21 | counter_type_code | INT | Код вида счетчика |
| 22 | counter_code | INT | Код счетчика |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | ps_protocol_number | INT | Номер протокола ПС |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

Особенности (по PDF):
- поле №8: ссылочный номер карты (RS.Loyalty.АСТОР) или номер подарочной карты.
- поля №15 и №16 заполняются только при внутренней предоплате или подарочных картах.
- поле №10: код операции (0 наличные, 1 карта/QR, 3 предоплата, 6 внутр. подарочная, 7 пользовательская, 8 внешняя подарочная, 9 карта с выдачей наличных).
- сдача в полях №11/12 передается отрицательными значениями.
- поле №20: сумма выдачи наличных (0 при отказе).
- поле №17 в режиме "один чек на несколько ГП" не использовать как источник распределения.
- поле №29 заполняется только при оплате через Frontol Driver Unit.

## 43 Фискальная оплата, распределенная по группам печати

Таблица документации: `tx_fiscal_payment_43`  
Колонки: как в `tx_fiscal_payment_40`, но поле №11/12 = распределение по ГП, поле №17 = код группы печати распределения.

---

## 42 Открытие документа

Таблица документации: `tx_document_open_42`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 42 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | customer_card_numbers | TEXT | Номера карт клиента через `|` |
| 9 | dimension_value_codes | TEXT | Коды значений разрезов |
| 10 | reserved_10 | NUMERIC | – |
| 11 | reserved_11 | NUMERIC | – |
| 12 | reserved_12 | NUMERIC | – |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | customer_code | NUMERIC | Код клиента |
| 16 | reserved_16 | NUMERIC | – |
| 17 | document_print_group_code | INT | Код группы печати документа |
| 18 | reserved_18 | TEXT | – |
| 19 | order_id | TEXT | Идентификатор заказа |
| 20 | document_amount_without_discounts | NUMERIC | Сумма без скидок |
| 21 | visitor_count | INT | Количество посетителей |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | comment_code | INT | Код комментария |
| 25 | base_document_number | INT | Номер документа основания |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | employee_code | INT | Код сотрудника |
| 29 | employee_edit_document_number | INT | Номер документа редактирования сотрудников |
| 30 | department_code | TEXT | Код подразделения |
| 31 | hall_code | INT | Код зала |
| 32 | service_point_code | INT | Код точки обслуживания |
| 33 | reservation_id | TEXT | Идентификатор откладывания/резервирования |
| 34 | user_variables | TEXT | Пользовательские переменные |
| 35 | external_comment | TEXT | Внешний комментарий документа |
| 36 | revaluation_datetime | DATE | Дата и время переоценки |
| 37 | contractor_code | INT | Код контрагента |
| 38 | subdivision_id | TEXT | Идентификатор подразделения |
| 39 | reserved_39 | TEXT | – |
| 43 | document_coupons | TEXT | Купоны на документ |
| 44 | reserved_44 | DATE | – |

Различия по типам в группе "Открытие/закрытие документа":
- 45: поле №12 = итоговая сумма в валюте ККТ; поле №21 = рег. номер ККТ; №24 = фискальный номер документа; №25 = фискальный признак; №26 = кассовый номер чека/документа/смены; №29 = номер смены ФН; №35 = коды ошибок обмена; №36 = дата/время закрытия; №19 = серийный номер ККМ.
- 49: поле №9 = 3; поле №12 = итоговая сумма по ГП; поле №17 = код ГП; поле №34 = количество неотправленных документов; поле №44 = дата/время расчета.
- 55: поле №12 = итоговая сумма документа в базовой валюте; поле №18 = начисленная сумма бонуса ПС.
- 56: поле №12 = итоговая сумма документа в базовой валюте.
- 58: поле №11 = количество товара; поле №22 = тип коррекции; поле №36 = дата документа основания коррекции.
- 65: поле №11 = количество товара; поле №30 = данные клиента; поле №34 = значения полей карточки клиента.
- 120: поле №17 = код указанной ГП; поле №9 = коды значений разрезов.
- 45: поле №34 = URL-адрес страницы просмотра чека в сети Интернет.

Особенности (по PDF):
- транзакция №65 выгружается только при наличии вида документа с операцией "Служебная операция".
- 45 выгружается только при закрытии документа на ККМ (не на принтере чеков/документов).
- для 45 при ДККТ 10 поле №26 может быть `0/0/0`.
- поле №17: для 45 — код ГП ККМ, для 49 — код ГП закрытия, для 120 — код указанной ГП.
- поле №29: для 42/55/56/58 — номер документа редактирования сотрудников (если есть).
- поле №33: внешний идентификатор для 42/55.
- поле №43: купоны на документ для 42/55/56.
- для 45 в поле №35 коды ошибок обмена (формат зависит от онлайн/не онлайн ККМ).
- для 42/55/56/58 поле №8 может содержать номера карт с префиксами `@`, `#`, `$` (авторизация по телефону).
- для 42 поле №36 заполняется только для документов переоценки; для 55 — только для коррекции; для 45 — для всех закрытых в ККМ.
- для 45 поля №2/3 содержат дату/время операции закрытия (а не открытия документа).
- поле №34: при включенной настройке выгружаются пользовательские переменные в формате "Имя = значение" (переводы строки заменяются на `|`).
- для 56/58 в поле №11 при документе редактирования сотрудников передается количество сотрудников.
- для 42/55/56 поле №25 содержит номер документа продажи-основания (для предоплат/кредита).
- для рецептурных товаров в 42/55/56 поле №8 содержит префикс вида карт + реквизиты рецепта через `$`.
- для постановки/снятия кега (42/49/55) поля №11/12/16/20 несут объемные значения (см. PDF).
- при режиме "один чек на несколько групп печати" итоги по ГП брать из транзакции 49 (или рассчитывать по 1/11, 2/12, 4/14).

## 55 Закрытие документа

Таблица документации: `tx_document_close_55`  
Колонки: как в `tx_document_open_42`, но поле №12 = итоговая сумма документа в базовой валюте, поле №18 = начисленная сумма бонуса ПС.

Смысловые поля (вместо `reserved_*`):
- поле №18 = bonus_amount_base

## 56 Отмена документа

Таблица документации: `tx_document_cancel_56`  
Колонки: как в `tx_document_open_42`, поле №12 = итоговая сумма документа в базовой валюте.

## 58 Нефинансовое закрытие возврата

Таблица документации: `tx_document_non_fin_close_58`  
Колонки: как в `tx_document_open_42`, поле №11 = количество товара, поле №22 = тип коррекции.

Смысловые поля (вместо `reserved_*`):
- поле №22 = correction_type
- поле №36 = base_document_date

## 65 Информация о клиентах

Таблица документации: `tx_document_clients_65`  
Колонки: как в `tx_document_open_42`, поле №11 = количество товара, поле №30 = данные клиента, поле №34 = значения полей карточки клиента.

Смысловые поля (вместо `reserved_*`):
- поле №30 = client_data
- поле №34 = client_card_fields
## 45 Закрытие документа в ККМ

Таблица документации: `tx_document_close_kkt_45`  
Колонки: как в `tx_document_open_42`, поле №12 = итог. сумма в валюте ККТ, поле №21 = рег. номер ККТ, поле №24 = фискальный номер, поле №25 = фискальный признак, поле №26 = кассовый номер, поле №29 = номер смены ФН, поле №35 = коды ошибок обмена, поле №36 = дата/время закрытия.

Смысловые поля (вместо `reserved_*`):
- поле №19 = kkm_serial_number
- поле №21 = kkt_registration_number
- поле №22 = fiscal_storage_factory_number
- поле №24 = fiscal_document_number
- поле №25 = fiscal_document_sign
- поле №29 = fn_shift_number
- поле №34 = receipt_view_url
- поле №35 = exchange_error_codes
- поле №36 = close_date

## 49 Закрытие документа по ГП

Таблица документации: `tx_document_close_gp_49`  
Колонки: как в `tx_document_open_42`, поле №9 = 3, поле №12 = сумма по ГП, поле №17 = код ГП, поле №34 = количество неотправленных документов, поле №44 = дата/время расчета.

Смысловые поля (вместо `reserved_*`):
- поле №34 = unreported_docs_count
- поле №44 = settlement_date

## 120 Отправка в ЕГАИС товаров с указанной ГП

Таблица документации: `tx_document_egais_120`  
Колонки: как в `tx_document_open_42`, поле №9 = коды значений разрезов, поле №17 = код ГП.

---

## 88 НДС по чеку из ККТ

Таблица документации: `tx_vat_kkt_88`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 88 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | reserved_8 | TEXT | – |
| 9 | reserved_9 | TEXT | – |
| 10 | reserved_10 | NUMERIC | – |
| 11 | reserved_11 | NUMERIC | – |
| 12 | reserved_12 | NUMERIC | – |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | reserved_15 | INT | – |
| 16 | reserved_16 | INT | – |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | vat_0_amount | NUMERIC | Сумма НДС 0% |
| 29 | vat_10_amount | NUMERIC | Сумма НДС 10% |
| 30 | vat_20_amount | NUMERIC | Сумма НДС 20% |
| 31 | no_vat_amount | NUMERIC | Сумма без НДС |
| 32 | vat_10_110_amount | NUMERIC | Сумма НДС 10/110 |
| 33 | vat_20_120_amount | NUMERIC | Сумма НДС 20/120 |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |
| 37 | reserved_37 | INT | – |
| 38 | reserved_38 | TEXT | – |
| 39 | reserved_39 | TEXT | – |
| 43 | reserved_43 | TEXT | – |

Особенности (по PDF):
- значение `-1` в полях №28–33 означает, что детализация НДС не поддерживается для данной ККМ.

---

## 50 Внесение

Таблица документации: `tx_cash_in_50`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 50 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | reserved_8 | TEXT | – |
| 9 | reserved_9 | TEXT | – |
| 10 | reserved_10 | NUMERIC | – |
| 11 | reserved_11 | NUMERIC | – |
| 12 | amount_base | NUMERIC | Сумма в базовой валюте |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | reserved_15 | NUMERIC | – |
| 16 | reserved_16 | NUMERIC | – |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | order_id | INT | Идентификатор заказа |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

## 51 Выплата

Таблица документации: `tx_cash_out_51`  
Колонки: как в `tx_cash_in_50`.

Особенности (по PDF):
- в режиме "Один чек на несколько групп печати" поле №17 содержит текущую ГП (для совместимости) и не должно использоваться для аналитики по позициям.

---

## 57 Изменение счетчика

Таблица документации: `tx_counter_change_57`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 57 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | card_or_client_code | TEXT | Номер карты или код клиента |
| 9 | card_type_code | TEXT | Код вида карты |
| 10 | binding_type | NUMERIC | Привязка (1..4) |
| 11 | value_after_changes | NUMERIC | Значение после изменений |
| 12 | change_amount | NUMERIC | Сумма изменения |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | promotion_code | INT | Код акции |
| 16 | event_code | INT | Код мероприятия |
| 17 | reserved_17 | INT | – |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | counter_type_code | INT | Код вида счетчика |
| 22 | counter_code | INT | Код счетчика |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | counter_valid_from | TEXT | Дата начала действия движения |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | card_valid_from | TEXT | Дата начала действия карты |
| 34 | card_valid_to | TEXT | Дата окончания действия карты |
| 35 | counter_valid_to | TEXT | Дата окончания действия движения |
| 36 | reserved_36 | DATE | – |

Особенности (по PDF):
- при активации подарочной карты после транзакции №27 в поле №11 выгружается баланс подарочной карты, а в поле №12 — 0.

---

## 60 Отчет без гашения

Таблица документации: `tx_report_zless_60`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 60 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | reserved_8 | TEXT | – |
| 9 | reserved_9 | TEXT | – |
| 10 | shift_revenue | NUMERIC | Выручка за смену |
| 11 | cash_in_drawer | NUMERIC | Наличность в кассе |
| 12 | shift_income_total | NUMERIC | Сменный итог приходов |
| 13 | reserved_13 | INT | – |
| 14 | shift_number | INT | Номер смены |
| 15 | reserved_15 | NUMERIC | – |
| 16 | reserved_16 | NUMERIC | – |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | reserved_23 | INT | – |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | TEXT | – |
| 26 | cash_document_number | TEXT | Кассовый номер чека/документа/смены |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | unreported_docs_count | TEXT | Кол-во неотправленных документов |
| 35 | exchange_error_codes | TEXT | Коды ошибок обмена |
| 36 | earliest_unreported_doc_datetime | DATE | Дата/время раннего неотправленного |
| 44 | reserved_44 | DATE | – |

Различия по типам в группе "Отчеты":
- 60: поля №10, №11, №12 заполнены (выручка, наличность, сменный итог приходов).
- 63: поля №10, №12 содержат аппаратные данные (Z-отчет).
- 64: поля №10, №12 содержат аппаратные данные (документ открытия смены).
- 61: поле №10 = выручка за смену, поле №12 = сменный итог продаж, поле №21 = рег. номер ККТ, №22 = номер ФН, №24 = номер фискального документа, №25 = фискальный признак, №29 = номер смены ФН, №44 = дата/время расчета.
- 62: большинство полей пустые, важны №14 (номер смены) и общие поля документа.

Особенности (по PDF):
- для 61/62 при закрытии/открытии смены через приложение администратора поле №17 выгружается как 0.
- для 60/63/64 при ДККТ 10 поле №26 может быть `0/0/0`.
- поле №35 для 60/63/64 содержит коды ошибок обмена в формате `x, y, z` (онлайн ККМ) или `<код> – <текст>` (не онлайн).
- для 61 поля №10/12 — программные данные (выручка/сумма продаж).
- для 60/63/64 поля №10/12 — аппаратные данные (фискальные документы ККТ).

## 63 Отчет с гашением

Таблица документации: `tx_report_z_63`  
Колонки: как в `tx_report_zless_60`.

## 64 Документ открытия смены

Таблица документации: `tx_shift_open_doc_64`  
Колонки: как в `tx_report_zless_60`.

## 61 Закрытие смены

Таблица документации: `tx_shift_close_61`  
Колонки: как в `tx_report_zless_60`, поле №21 = рег. номер ККТ, №22 = номер ФН, №24 = номер фискального документа, №25 = фискальный признак, №29 = номер смены ФН, №44 = дата/время расчета.

Смысловые поля (вместо `reserved_*`):
- поле №21 = kkt_registration_number
- поле №22 = fiscal_storage_factory_number
- поле №24 = fiscal_document_number
- поле №25 = fiscal_document_sign
- поле №29 = fn_shift_number
- поле №44 = settlement_date

## 62 Открытие смены

Таблица документации: `tx_shift_open_62`  
Колонки: как в `tx_report_zless_60` (большинство полей пустые).

---

## 121 Frontol Mark Unit

Таблица документации: `tx_mark_unit_121`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 121 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | reserved_8 | TEXT | – |
| 9 | reserved_9 | TEXT | – |
| 10 | reserved_10 | NUMERIC | – |
| 11 | reserved_11 | NUMERIC | – |
| 12 | reserved_12 | NUMERIC | – |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | reserved_15 | INT | – |
| 16 | reserved_16 | INT | – |
| 17 | print_group_code | INT | Код группы печати чека |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | TEXT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | reserved_26 | TEXT | – |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | TEXT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |
| 39 | reserved_39 | TEXT | – |
| 40 | reserved_40 | NUMERIC | – |
| 41 | reserved_41 | TEXT | – |
| 42 | reserved_42 | TEXT | – |
| 43 | reserved_43 | INT | – |

Особенности (по PDF):
- поле №17 всегда выгружается как 0.

---

## 3 Установка спеццены / цены из прайс-листа

Таблица документации: `tx_special_price_3`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 3 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | price_list_code | TEXT | Код прайс-листа |
| 9 | reserved_9 | TEXT | – |
| 10 | price_type | NUMERIC | Тип цены |
| 11 | special_price | NUMERIC | Спеццена/цена из прайс-листа |
| 12 | product_card_price | NUMERIC | Цена из карточки товара |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | promotion_code | INT | Код акции |
| 16 | event_code | INT | Код мероприятия |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

---

## 9 Начисление бонуса

Таблица документации: `tx_bonus_accrual_9`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 9 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | reserved_8 | TEXT | – |
| 9 | reserved_9 | TEXT | – |
| 10 | bonus_type | NUMERIC | Тип бонуса |
| 11 | reserved_11 | NUMERIC | – |
| 12 | bonus_amount | NUMERIC | Начисленная сумма |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | promotion_code | INT | Код акции |
| 16 | event_code | INT | Код мероприятия |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | counter_type_code | INT | Код вида счетчика |
| 22 | counter_code | INT | Код счетчика |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | ps_protocol_number | INT | Номер протокола ПС |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | activation_date | TEXT | Дата активации начисления |
| 34 | expiration_date | TEXT | Дата сгорания начисления |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

Особенности (по PDF):
- поле №10: тип бонуса (0 внутренний, 1 внешний).
- поле №20 (для 10): сгоревшая сумма бонуса, не подлежащая возврату.
- поле №33: дата начала действия начисления (если не сразу).
- поле №34: дата сгорания начисления (если не бессрочно).

## 10 Возврат бонуса

Таблица документации: `tx_bonus_refund_10`  
Колонки: как в `tx_bonus_accrual_9`, но поле №12 = возвращенная сумма, поле №20 = сгоревшая сумма.

---

## 15 Скидка суммой на позицию

Таблица документации: `tx_position_discount_15`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 15 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | discount_info | TEXT | Информация по скидке |
| 9 | reserved_9 | TEXT | – |
| 10 | discount_type | NUMERIC | Тип скидки |
| 11 | discount_value | NUMERIC | Значение скидки |
| 12 | discount_amount_base | NUMERIC | Сумма скидки в базовой валюте |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | promotion_code | INT | Код акции |
| 16 | event_code | INT | Код мероприятия |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

Особенности (по PDF):
- поле №8: дисконтная карта/классификатор, если сработало условие.
- поле №10: тип скидки (0,1,2,3,6,10,11).
- поля №11 и №12: знак зависит от начисления/возврата скидки.

## 17 Скидка % на позицию

Таблица документации: `tx_position_discount_17`  
Колонки: как в `tx_position_discount_15`.

---

## 21/23 Регистрация купюр

Таблица документации: `tx_bill_registration_21_23`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 21/23 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | bill_code | TEXT | Код купюры |
| 9 | reserved_9 | TEXT | – |
| 10 | bill_denomination | NUMERIC | Достоинство купюры |
| 11 | bill_quantity | NUMERIC | Количество купюр |
| 12 | bill_amount_base | NUMERIC | Сумма купюр |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | reserved_15 | NUMERIC | – |
| 16 | reserved_16 | NUMERIC | – |
| 17 | reserved_17 | INT | – |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

## 22/24 Сторно купюр

Таблица документации: `tx_bill_storno_22_24`  
Колонки: как в `tx_bill_registration_21_23`.

---

## 25 Регистрация сотрудников

Таблица документации: `tx_employee_registration_25`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 25 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | employee_code | TEXT | Код сотрудника |
| 9 | reserved_9 | TEXT | – |
| 10 | reserved_10 | NUMERIC | – |
| 11 | reserved_11 | NUMERIC | – |
| 12 | reserved_12 | NUMERIC | – |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | reserved_15 | NUMERIC | – |
| 16 | reserved_16 | NUMERIC | – |
| 17 | reserved_17 | INT | – |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

---

## 26 Учет сотрудников по документу

Таблица документации: `tx_employee_accounting_doc_26`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 26 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | employee_code | TEXT | Код сотрудника |
| 9 | reserved_9 | TEXT | – |
| 10 | reserved_10 | NUMERIC | – |
| 11 | reserved_11 | NUMERIC | – |
| 12 | reserved_12 | NUMERIC | – |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | reserved_15 | NUMERIC | – |
| 16 | reserved_16 | NUMERIC | – |
| 17 | document_print_group_code | INT | Код группы печати документа |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

## 29 Учет сотрудников по позиции

Таблица документации: `tx_employee_accounting_pos_29`  
Колонки: как в `tx_employee_accounting_doc_26`.

---

## 27 Изменение статуса карты

Таблица документации: `tx_card_status_change_27`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 27 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | card_number | TEXT | Номер карты |
| 9 | card_type_code | TEXT | Код вида карты |
| 10 | card_type | NUMERIC | Тип карты |
| 11 | reserved_11 | NUMERIC | – |
| 12 | reserved_12 | NUMERIC | – |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | promotion_code | NUMERIC | Код акции |
| 16 | event_code | NUMERIC | Код мероприятия |
| 17 | reserved_17 | INT | – |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | old_card_status | INT | Старый статус |
| 32 | new_card_status | INT | Новый статус |
| 33 | new_valid_from | TEXT | Новая дата начала действия |
| 34 | new_valid_to | TEXT | Новая дата окончания действия |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

Особенности (по PDF):
- поле №10: тип карты (0 дисконтная, 1 подарочная).
- поля №31/32: статусы (0 неактивна, 1 активна, 2 погашена).
- поля №15/16 заполняются для подарочных карт (акция/мероприятие).

---

## 30 Регистрация модификаторов

Таблица документации: `tx_modifier_registration_30`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 30 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | item_identifier | TEXT | Идентификатор товара |
| 9 | reserved_9 | TEXT | – |
| 10 | reserved_10 | NUMERIC | – |
| 11 | item_quantity | NUMERIC | Количество товара |
| 12 | reserved_12 | NUMERIC | – |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | reserved_15 | NUMERIC | – |
| 16 | reserved_16 | NUMERIC | – |
| 17 | document_print_group_code | INT | Код группы печати документа |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | modifier_code | TEXT | Код модификатора |
| 36 | reserved_36 | DATE | – |

Особенности (по PDF):
- если товар для модификатора не задан, в поле №11 выгружается количество выборов модификатора.

## 31 Сторнирование модификаторов при сторнировании товара

Таблица документации: `tx_modifier_storno_31`  
Колонки: как в `tx_modifier_registration_30`.

---

## 32 Оплата бонусом

Таблица документации: `tx_bonus_payment_32`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 32 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | bonus_card_number | TEXT | Номер бонусной карты |
| 9 | reserved_9 | TEXT | – |
| 10 | bonus_payment_type | NUMERIC | Тип оплаты бонусом |
| 11 | counter_change_amount | NUMERIC | Изменение счетчика |
| 12 | payment_amount | NUMERIC | Сумма оплаты |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | promotion_code | INT | Код акции |
| 16 | event_code | INT | Код мероприятия |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | counter_type_code | INT | Код вида счетчика |
| 22 | counter_code | INT | Код счетчика |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | ps_protocol_number | INT | Номер протокола ПС |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

Особенности (по PDF):
- поле №8 заполняется только для RS.Loyalty.АСТОР при `AstorAlwaysWriteDiscTrazForBonusCard = 1`.
- поле №10: тип оплаты/возврата (0 внутренний бонус, 1 внешний).

## 82 Распределение оплаты бонусом по позициям

Таблица документации: `tx_bonus_payment_82`  
Колонки: как в `tx_bonus_payment_32`, поле №12 = распределение по позициям.

## 33 Возврат оплаты бонусом

Таблица документации: `tx_bonus_payment_33`  
Колонки: как в `tx_bonus_payment_32`, поле №10 = тип возврата, поле №12 = сумма возврата.

## 83 Распределение возврата оплаты бонусом по позициям

Таблица документации: `tx_bonus_payment_83`  
Колонки: как в `tx_bonus_payment_32`, поле №12 = распределение суммы возврата.

---

## 34 Предоплата документом

Таблица документации: `tx_prepayment_34`

| № | Колонка | Тип | Описание |
|---|---------|-----|----------|
| 1 | transaction_id_unique | BIGINT | № транзакции |
| 2 | transaction_date | DATE | Дата транзакции |
| 3 | transaction_time | TIME | Время транзакции |
| 4 | transaction_type | INT | 34 |
| 5 | cash_register_code | INT | Код РМ |
| 6 | document_number | BIGINT | Номер документа |
| 7 | cashier_code | BIGINT | Код кассира |
| 8 | reserved_8 | TEXT | – |
| 9 | reserved_9 | TEXT | – |
| 10 | prepayment_type | NUMERIC | Тип предоплаты |
| 11 | reserved_11 | NUMERIC | – |
| 12 | prepayment_amount | NUMERIC | Сумма предоплаты |
| 13 | operation_type | INT | Операция |
| 14 | shift_number | INT | Номер смены |
| 15 | reserved_15 | NUMERIC | – |
| 16 | reserved_16 | NUMERIC | – |
| 17 | print_group_code | INT | Код группы печати |
| 18 | reserved_18 | TEXT | – |
| 19 | reserved_19 | INT | – |
| 20 | reserved_20 | NUMERIC | – |
| 21 | reserved_21 | INT | – |
| 22 | reserved_22 | INT | – |
| 23 | document_type_code | INT | Код вида документа |
| 24 | reserved_24 | INT | – |
| 25 | reserved_25 | INT | – |
| 26 | document_info | TEXT | Информация о документе |
| 27 | enterprise_id | INT | Идентификатор предприятия |
| 28 | reserved_28 | INT | – |
| 29 | reserved_29 | INT | – |
| 30 | reserved_30 | TEXT | – |
| 31 | reserved_31 | INT | – |
| 32 | reserved_32 | INT | – |
| 33 | reserved_33 | TEXT | – |
| 34 | reserved_34 | TEXT | – |
| 35 | reserved_35 | TEXT | – |
| 36 | reserved_36 | DATE | – |

Особенности (по PDF):
- функция предоплаты документом несовместима с онлайн ККТ.

## 84 Предоплата документом, распределенная по позициям

Таблица документации: `tx_prepayment_84`  
Колонки: как в `tx_prepayment_34`, поле №12 = распределение по позициям.
