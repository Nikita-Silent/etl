-- Migration: 000001_init_schema
-- Description: Create transaction tables per docs/DDL_SPEC.md

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

CREATE TABLE tx_item_storno_2_12 (
  LIKE tx_item_registration_1_11 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_document_discount_37 (
  LIKE tx_document_discount_35 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_document_discount_85 (
  LIKE tx_document_discount_35 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_document_discount_87 (
  LIKE tx_document_discount_35 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_document_rounding_38 (
  LIKE tx_document_discount_35 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_non_fiscal_payment_86 (
  LIKE tx_non_fiscal_payment_36 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_fiscal_payment_43 (
  LIKE tx_fiscal_payment_40 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_document_close_55 (
  LIKE tx_document_open_42 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_document_cancel_56 (
  LIKE tx_document_open_42 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_document_non_fin_close_58 (
  LIKE tx_document_open_42 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_document_clients_65 (
  LIKE tx_document_open_42 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_document_close_kkt_45 (
  LIKE tx_document_open_42 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_document_close_gp_49 (
  LIKE tx_document_open_42 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_document_egais_120 (
  LIKE tx_document_open_42 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_cash_out_51 (
  LIKE tx_cash_in_50 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_report_z_63 (
  LIKE tx_report_zless_60 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_shift_open_doc_64 (
  LIKE tx_report_zless_60 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_shift_close_61 (
  LIKE tx_report_zless_60 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_shift_open_62 (
  LIKE tx_report_zless_60 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_bonus_refund_10 (
  LIKE tx_bonus_accrual_9 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_position_discount_17 (
  LIKE tx_position_discount_15 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_bill_storno_22_24 (
  LIKE tx_bill_registration_21_23 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_employee_accounting_pos_29 (
  LIKE tx_employee_accounting_doc_26 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_modifier_storno_31 (
  LIKE tx_modifier_registration_30 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_bonus_payment_82 (
  LIKE tx_bonus_payment_32 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_bonus_payment_33 (
  LIKE tx_bonus_payment_32 INCLUDING CONSTRAINTS
);

CREATE TABLE tx_bonus_payment_83 (
  LIKE tx_bonus_payment_32 INCLUDING CONSTRAINTS
);

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

CREATE TABLE tx_prepayment_84 (
  LIKE tx_prepayment_34 INCLUDING CONSTRAINTS
);

CREATE INDEX tx_item_registration_1_11_date_idx ON tx_item_registration_1_11 (transaction_date);
CREATE INDEX tx_item_registration_1_11_source_idx ON tx_item_registration_1_11 (source_folder);
CREATE INDEX tx_item_storno_2_12_date_idx ON tx_item_storno_2_12 (transaction_date);
CREATE INDEX tx_item_storno_2_12_source_idx ON tx_item_storno_2_12 (source_folder);
CREATE INDEX tx_item_tax_4_14_date_idx ON tx_item_tax_4_14 (transaction_date);
CREATE INDEX tx_item_tax_4_14_source_idx ON tx_item_tax_4_14 (source_folder);
CREATE INDEX tx_item_kkt_6_16_date_idx ON tx_item_kkt_6_16 (transaction_date);
CREATE INDEX tx_item_kkt_6_16_source_idx ON tx_item_kkt_6_16 (source_folder);
CREATE INDEX tx_document_discount_35_date_idx ON tx_document_discount_35 (transaction_date);
CREATE INDEX tx_document_discount_35_source_idx ON tx_document_discount_35 (source_folder);
CREATE INDEX tx_document_discount_37_date_idx ON tx_document_discount_37 (transaction_date);
CREATE INDEX tx_document_discount_37_source_idx ON tx_document_discount_37 (source_folder);
CREATE INDEX tx_document_discount_85_date_idx ON tx_document_discount_85 (transaction_date);
CREATE INDEX tx_document_discount_85_source_idx ON tx_document_discount_85 (source_folder);
CREATE INDEX tx_document_discount_87_date_idx ON tx_document_discount_87 (transaction_date);
CREATE INDEX tx_document_discount_87_source_idx ON tx_document_discount_87 (source_folder);
CREATE INDEX tx_document_rounding_38_date_idx ON tx_document_rounding_38 (transaction_date);
CREATE INDEX tx_document_rounding_38_source_idx ON tx_document_rounding_38 (source_folder);
CREATE INDEX tx_non_fiscal_payment_36_date_idx ON tx_non_fiscal_payment_36 (transaction_date);
CREATE INDEX tx_non_fiscal_payment_36_source_idx ON tx_non_fiscal_payment_36 (source_folder);
CREATE INDEX tx_non_fiscal_payment_86_date_idx ON tx_non_fiscal_payment_86 (transaction_date);
CREATE INDEX tx_non_fiscal_payment_86_source_idx ON tx_non_fiscal_payment_86 (source_folder);
CREATE INDEX tx_fiscal_payment_40_date_idx ON tx_fiscal_payment_40 (transaction_date);
CREATE INDEX tx_fiscal_payment_40_source_idx ON tx_fiscal_payment_40 (source_folder);
CREATE INDEX tx_fiscal_payment_43_date_idx ON tx_fiscal_payment_43 (transaction_date);
CREATE INDEX tx_fiscal_payment_43_source_idx ON tx_fiscal_payment_43 (source_folder);
CREATE INDEX tx_document_open_42_date_idx ON tx_document_open_42 (transaction_date);
CREATE INDEX tx_document_open_42_source_idx ON tx_document_open_42 (source_folder);
CREATE INDEX tx_document_close_55_date_idx ON tx_document_close_55 (transaction_date);
CREATE INDEX tx_document_close_55_source_idx ON tx_document_close_55 (source_folder);
CREATE INDEX tx_document_cancel_56_date_idx ON tx_document_cancel_56 (transaction_date);
CREATE INDEX tx_document_cancel_56_source_idx ON tx_document_cancel_56 (source_folder);
CREATE INDEX tx_document_non_fin_close_58_date_idx ON tx_document_non_fin_close_58 (transaction_date);
CREATE INDEX tx_document_non_fin_close_58_source_idx ON tx_document_non_fin_close_58 (source_folder);
CREATE INDEX tx_document_clients_65_date_idx ON tx_document_clients_65 (transaction_date);
CREATE INDEX tx_document_clients_65_source_idx ON tx_document_clients_65 (source_folder);
CREATE INDEX tx_document_close_kkt_45_date_idx ON tx_document_close_kkt_45 (transaction_date);
CREATE INDEX tx_document_close_kkt_45_source_idx ON tx_document_close_kkt_45 (source_folder);
CREATE INDEX tx_document_close_gp_49_date_idx ON tx_document_close_gp_49 (transaction_date);
CREATE INDEX tx_document_close_gp_49_source_idx ON tx_document_close_gp_49 (source_folder);
CREATE INDEX tx_document_egais_120_date_idx ON tx_document_egais_120 (transaction_date);
CREATE INDEX tx_document_egais_120_source_idx ON tx_document_egais_120 (source_folder);
CREATE INDEX tx_vat_kkt_88_date_idx ON tx_vat_kkt_88 (transaction_date);
CREATE INDEX tx_vat_kkt_88_source_idx ON tx_vat_kkt_88 (source_folder);
CREATE INDEX tx_cash_in_50_date_idx ON tx_cash_in_50 (transaction_date);
CREATE INDEX tx_cash_in_50_source_idx ON tx_cash_in_50 (source_folder);
CREATE INDEX tx_cash_out_51_date_idx ON tx_cash_out_51 (transaction_date);
CREATE INDEX tx_cash_out_51_source_idx ON tx_cash_out_51 (source_folder);
CREATE INDEX tx_counter_change_57_date_idx ON tx_counter_change_57 (transaction_date);
CREATE INDEX tx_counter_change_57_source_idx ON tx_counter_change_57 (source_folder);
CREATE INDEX tx_report_zless_60_date_idx ON tx_report_zless_60 (transaction_date);
CREATE INDEX tx_report_zless_60_source_idx ON tx_report_zless_60 (source_folder);
CREATE INDEX tx_report_z_63_date_idx ON tx_report_z_63 (transaction_date);
CREATE INDEX tx_report_z_63_source_idx ON tx_report_z_63 (source_folder);
CREATE INDEX tx_shift_open_doc_64_date_idx ON tx_shift_open_doc_64 (transaction_date);
CREATE INDEX tx_shift_open_doc_64_source_idx ON tx_shift_open_doc_64 (source_folder);
CREATE INDEX tx_shift_close_61_date_idx ON tx_shift_close_61 (transaction_date);
CREATE INDEX tx_shift_close_61_source_idx ON tx_shift_close_61 (source_folder);
CREATE INDEX tx_shift_open_62_date_idx ON tx_shift_open_62 (transaction_date);
CREATE INDEX tx_shift_open_62_source_idx ON tx_shift_open_62 (source_folder);
CREATE INDEX tx_mark_unit_121_date_idx ON tx_mark_unit_121 (transaction_date);
CREATE INDEX tx_mark_unit_121_source_idx ON tx_mark_unit_121 (source_folder);
CREATE INDEX tx_special_price_3_date_idx ON tx_special_price_3 (transaction_date);
CREATE INDEX tx_special_price_3_source_idx ON tx_special_price_3 (source_folder);
CREATE INDEX tx_bonus_accrual_9_date_idx ON tx_bonus_accrual_9 (transaction_date);
CREATE INDEX tx_bonus_accrual_9_source_idx ON tx_bonus_accrual_9 (source_folder);
CREATE INDEX tx_bonus_refund_10_date_idx ON tx_bonus_refund_10 (transaction_date);
CREATE INDEX tx_bonus_refund_10_source_idx ON tx_bonus_refund_10 (source_folder);
CREATE INDEX tx_position_discount_15_date_idx ON tx_position_discount_15 (transaction_date);
CREATE INDEX tx_position_discount_15_source_idx ON tx_position_discount_15 (source_folder);
CREATE INDEX tx_position_discount_17_date_idx ON tx_position_discount_17 (transaction_date);
CREATE INDEX tx_position_discount_17_source_idx ON tx_position_discount_17 (source_folder);
CREATE INDEX tx_bill_registration_21_23_date_idx ON tx_bill_registration_21_23 (transaction_date);
CREATE INDEX tx_bill_registration_21_23_source_idx ON tx_bill_registration_21_23 (source_folder);
CREATE INDEX tx_bill_storno_22_24_date_idx ON tx_bill_storno_22_24 (transaction_date);
CREATE INDEX tx_bill_storno_22_24_source_idx ON tx_bill_storno_22_24 (source_folder);
CREATE INDEX tx_employee_registration_25_date_idx ON tx_employee_registration_25 (transaction_date);
CREATE INDEX tx_employee_registration_25_source_idx ON tx_employee_registration_25 (source_folder);
CREATE INDEX tx_employee_accounting_doc_26_date_idx ON tx_employee_accounting_doc_26 (transaction_date);
CREATE INDEX tx_employee_accounting_doc_26_source_idx ON tx_employee_accounting_doc_26 (source_folder);
CREATE INDEX tx_employee_accounting_pos_29_date_idx ON tx_employee_accounting_pos_29 (transaction_date);
CREATE INDEX tx_employee_accounting_pos_29_source_idx ON tx_employee_accounting_pos_29 (source_folder);
CREATE INDEX tx_card_status_change_27_date_idx ON tx_card_status_change_27 (transaction_date);
CREATE INDEX tx_card_status_change_27_source_idx ON tx_card_status_change_27 (source_folder);
CREATE INDEX tx_modifier_registration_30_date_idx ON tx_modifier_registration_30 (transaction_date);
CREATE INDEX tx_modifier_registration_30_source_idx ON tx_modifier_registration_30 (source_folder);
CREATE INDEX tx_modifier_storno_31_date_idx ON tx_modifier_storno_31 (transaction_date);
CREATE INDEX tx_modifier_storno_31_source_idx ON tx_modifier_storno_31 (source_folder);
CREATE INDEX tx_bonus_payment_32_date_idx ON tx_bonus_payment_32 (transaction_date);
CREATE INDEX tx_bonus_payment_32_source_idx ON tx_bonus_payment_32 (source_folder);
CREATE INDEX tx_bonus_payment_82_date_idx ON tx_bonus_payment_82 (transaction_date);
CREATE INDEX tx_bonus_payment_82_source_idx ON tx_bonus_payment_82 (source_folder);
CREATE INDEX tx_bonus_payment_33_date_idx ON tx_bonus_payment_33 (transaction_date);
CREATE INDEX tx_bonus_payment_33_source_idx ON tx_bonus_payment_33 (source_folder);
CREATE INDEX tx_bonus_payment_83_date_idx ON tx_bonus_payment_83 (transaction_date);
CREATE INDEX tx_bonus_payment_83_source_idx ON tx_bonus_payment_83 (source_folder);
CREATE INDEX tx_prepayment_34_date_idx ON tx_prepayment_34 (transaction_date);
CREATE INDEX tx_prepayment_34_source_idx ON tx_prepayment_34 (source_folder);
CREATE INDEX tx_prepayment_84_date_idx ON tx_prepayment_84 (transaction_date);
CREATE INDEX tx_prepayment_84_source_idx ON tx_prepayment_84 (source_folder);
