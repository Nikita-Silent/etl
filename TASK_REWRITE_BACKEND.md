# TASK_REWRITE_BACKEND

## Goal
Rewrite Frontol response parsing + DB loaders so they match the **current DB schema** (tx_* tables in `pkg/migrate/migrations/000001_init_schema.up.sql`) and keep **DB export/unload** functionality working (webhook reads, response generation, etc.).

## Sources of truth
- Transaction format: `docs/frontol_6_integration.md`
- DB schema: `pkg/migrate/migrations/000001_init_schema.up.sql`
- Existing parser/loader code: `pkg/parser`, `pkg/repository/loader.go`, `pkg/db/postgres.go`
- Mapping doc: `PARSERS_TO_TABLES_MAPPING.md` (currently references old tables)

## Current state (mismatch)
- Parsers produce models like `TransactionRegistration`, `SpecialPrice`, etc. and loaders insert into legacy tables (`transaction_registrations`, `special_prices`, ...).
- The **current schema** uses tx_* tables (e.g. `tx_item_registration_1_11`, `tx_document_discount_35`, `tx_fiscal_payment_40`) and different column naming.
- Export/unload endpoints rely on legacy tables (see `pkg/repository/loader.go` and `cmd/webhook-server/main.go`).

## Required outcomes
1. Parsers map file fields to **tx_* table columns** (not legacy tables).
2. DB loaders insert into **tx_* tables** with correct column lists/types.
3. DB unload/export queries read from **tx_* tables** and return correct raw rows for webhook/response use.
4. Update documentation under `docs/` if schema or business logic changes are made.

## Test policy (fresh start)
- Every block of changes must include unit tests.
- Workflow: write code for a block → write unit tests for that block → run tests → only after success proceed to next block.

## Export performance decision
- Chosen approach: **(1) single UNION ALL view/query over tx_* filtered by `source_folder` + `transaction_date`**
- Reason: lowest risk, uses existing per-table indexes (`transaction_date`, `source_folder`), no schema changes required.

---

## Task plan

### 1) Inventory transaction types and tables (done)
- Source schema: `pkg/migrate/migrations/000001_init_schema.up.sql`.
- Verified the full list of tx_* tables and mapped to transaction types below.
- Field position mapping still comes from `docs/frontol_6_integration.md`.

## Inventory results: transaction type -> tx_* table
| Transaction type(s) | tx_* table |
|---|---|
| 1, 11 | `tx_item_registration_1_11` |
| 2, 12 | `tx_item_storno_2_12` |
| 4, 14 | `tx_item_tax_4_14` |
| 6, 16 | `tx_item_kkt_6_16` |
| 3 | `tx_special_price_3` |
| 9 | `tx_bonus_accrual_9` |
| 10 | `tx_bonus_refund_10` |
| 15 | `tx_position_discount_15` |
| 17 | `tx_position_discount_17` |
| 21, 23 | `tx_bill_registration_21_23` |
| 22, 24 | `tx_bill_storno_22_24` |
| 25 | `tx_employee_registration_25` |
| 26 | `tx_employee_accounting_doc_26` |
| 29 | `tx_employee_accounting_pos_29` |
| 27 | `tx_card_status_change_27` |
| 30 | `tx_modifier_registration_30` |
| 31 | `tx_modifier_storno_31` |
| 32 | `tx_bonus_payment_32` |
| 33 | `tx_bonus_payment_33` |
| 82 | `tx_bonus_payment_82` |
| 83 | `tx_bonus_payment_83` |
| 34 | `tx_prepayment_34` |
| 84 | `tx_prepayment_84` |
| 35 | `tx_document_discount_35` |
| 37 | `tx_document_discount_37` |
| 38 | `tx_document_rounding_38` |
| 85 | `tx_document_discount_85` |
| 87 | `tx_document_discount_87` |
| 36 | `tx_non_fiscal_payment_36` |
| 86 | `tx_non_fiscal_payment_86` |
| 40 | `tx_fiscal_payment_40` |
| 43 | `tx_fiscal_payment_43` |
| 42 | `tx_document_open_42` |
| 45 | `tx_document_close_kkt_45` |
| 49 | `tx_document_close_gp_49` |
| 55 | `tx_document_close_55` |
| 56 | `tx_document_cancel_56` |
| 58 | `tx_document_non_fin_close_58` |
| 65 | `tx_document_clients_65` |
| 120 | `tx_document_egais_120` |
| 88 | `tx_vat_kkt_88` |
| 50 | `tx_cash_in_50` |
| 51 | `tx_cash_out_51` |
| 57 | `tx_counter_change_57` |
| 60 | `tx_report_zless_60` |
| 61 | `tx_shift_close_61` |
| 62 | `tx_shift_open_62` |
| 63 | `tx_report_z_63` |
| 64 | `tx_shift_open_doc_64` |
| 121 | `tx_mark_unit_121` |

## Full mapping table (field index -> column)
Notes:
- Field positions come from `docs/frontol_6_integration.md`.
- `source_folder` is a **system field** (not in the file); it must be injected for every row.
- All tables share the first 7 fields unless stated otherwise:
  1 `transaction_id_unique`, 2 `transaction_date`, 3 `transaction_time`,
  4 `transaction_type`, 5 `cash_register_code`, 6 `document_number`, 7 `cashier_code`.

## Go types ↔ PostgreSQL types (final)
- `BIGINT` -> `int64`
- `NUMERIC(18,6)` -> `float64`
- `TEXT` -> `string`
- `DATE` -> `time.Time` (date-only)
- `TIME` -> `time.Time` (time-only)

### Item registration / storno / tax / kkt
`tx_item_registration_1_11` (types 1, 11)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 item_identifier, 9 dimension_value_codes, 10 price_without_discounts,
11 quantity, 12 position_amount_with_rounding, 13 operation_type, 14 shift_number, 15 final_price,
16 position_total_amount, 17 print_group_code, 18 article_sku, 19 registration_barcode,
20 position_amount_base, 21 kkt_section, 22 reserved_22, 23 document_type_code, 24 comment_code,
25 reserved_25, 26 document_info, 27 enterprise_id, 28 employee_code, 29 split_pack_quantity,
30 gift_card_external_number, 31 pack_quantity, 32 item_type_code, 33 marking_code, 34 excise_marks,
35 personal_modifier_group_code, 36 stoloto_registration_time, 37 stoloto_ticket_id, 38 reserved_38,
39 alc_code, 40 reserved_40, 41 prescription_data_1, 42 prescription_data_2, 43 position_coupons,
44 reserved_44.

`tx_item_storno_2_12` (types 2, 12)
Same columns and field mapping as `tx_item_registration_1_11`.

`tx_item_tax_4_14` (types 4, 14)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 reserved_8, 9 dimension_value_codes, 10 tax_group_code,
11 tax_rate_code, 12 tax_amount_base, 13 operation_type, 14 shift_number, 15 reserved_15,
16 total_amount_base_with_discounts, 17 print_group_code, 18 reserved_18, 19 reserved_19,
20 amount_base_without_discounts, 21 reserved_21, 22 reserved_22, 23 document_type_code,
24 reserved_24, 25 reserved_25, 26 document_info, 27 enterprise_id, 28 reserved_28,
29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32, 33 reserved_33, 34 reserved_34,
35 reserved_35, 36 reserved_36.

`tx_item_kkt_6_16` (types 6, 16)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 item_identifier, 9 dimension_value_codes, 10 reserved_10,
11 quantity_kkt, 12 reserved_12, 13 operation_type, 14 shift_number, 15 final_price_kkt_currency,
16 position_total_amount_kkt_currency, 17 print_group_code, 18 article_sku, 19 registration_barcode,
20 reserved_20, 21 kkt_section, 22 reserved_22, 23 document_type_code, 24 comment_code, 25 reserved_25,
26 document_info, 27 enterprise_id, 28 employee_code, 29 split_pack_quantity, 30 gift_card_external_number,
31 pack_quantity, 32 item_type_code, 33 marking_code, 34 excise_marks, 35 personal_modifier_group_code,
36 stoloto_registration_time, 37 stoloto_ticket_id, 38 reserved_38, 39 alc_code, 40 reserved_40,
41 prescription_data_1, 42 prescription_data_2, 43 position_coupons, 44 reserved_44.

### Special price
`tx_special_price_3` (type 3)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 price_list_code, 9 reserved_9, 10 price_type, 11 special_price,
12 product_card_price, 13 operation_type, 14 shift_number, 15 promotion_code, 16 event_code,
17 print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21, 22 reserved_22,
23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info, 27 enterprise_id,
28 reserved_28, 29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32, 33 reserved_33,
34 reserved_34, 35 reserved_35, 36 reserved_36.

### Bonus accrual/refund
`tx_bonus_accrual_9` (type 9)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 reserved_8, 9 reserved_9, 10 bonus_type, 11 reserved_11,
12 bonus_amount, 13 operation_type, 14 shift_number, 15 promotion_code, 16 event_code,
17 print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 counter_type_code,
22 counter_code, 23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info,
27 enterprise_id, 28 reserved_28, 29 ps_protocol_number, 30 reserved_30, 31 reserved_31,
32 reserved_32, 33 activation_date, 34 expiration_date, 35 reserved_35, 36 reserved_36.

`tx_bonus_refund_10` (type 10)
Same columns and field mapping as `tx_bonus_accrual_9`.

### Position discount
`tx_position_discount_15` (type 15)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 discount_info, 9 reserved_9, 10 discount_type, 11 discount_value,
12 discount_amount_base, 13 operation_type, 14 shift_number, 15 promotion_code, 16 event_code,
17 print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21, 22 reserved_22,
23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info, 27 enterprise_id,
28 reserved_28, 29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32, 33 reserved_33,
34 reserved_34, 35 reserved_35, 36 reserved_36.

`tx_position_discount_17` (type 17)
Same columns and field mapping as `tx_position_discount_15`.

### Bill registration / storno
`tx_bill_registration_21_23` (types 21, 23)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 bill_code, 9 reserved_9, 10 bill_denomination, 11 bill_quantity,
12 bill_amount_base, 13 operation_type, 14 shift_number, 15 reserved_15, 16 reserved_16, 17 reserved_17,
18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21, 22 reserved_22, 23 document_type_code,
24 reserved_24, 25 reserved_25, 26 document_info, 27 enterprise_id, 28 reserved_28, 29 reserved_29,
30 reserved_30, 31 reserved_31, 32 reserved_32, 33 reserved_33, 34 reserved_34, 35 reserved_35,
36 reserved_36.

`tx_bill_storno_22_24` (types 22, 24)
Same columns and field mapping as `tx_bill_registration_21_23`.

### Employee registration / accounting
`tx_employee_registration_25` (type 25)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 employee_code, 9 reserved_9, 10 reserved_10, 11 reserved_11,
12 reserved_12, 13 operation_type, 14 shift_number, 15 reserved_15, 16 reserved_16, 17 reserved_17,
18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21, 22 reserved_22, 23 document_type_code,
24 reserved_24, 25 reserved_25, 26 document_info, 27 enterprise_id, 28 reserved_28, 29 reserved_29,
30 reserved_30, 31 reserved_31, 32 reserved_32, 33 reserved_33, 34 reserved_34, 35 reserved_35,
36 reserved_36.

`tx_employee_accounting_doc_26` (type 26)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 employee_code, 9 reserved_9, 10 reserved_10, 11 reserved_11,
12 reserved_12, 13 operation_type, 14 shift_number, 15 reserved_15, 16 reserved_16,
17 document_print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21,
22 reserved_22, 23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info,
27 enterprise_id, 28 reserved_28, 29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32,
33 reserved_33, 34 reserved_34, 35 reserved_35, 36 reserved_36.

`tx_employee_accounting_pos_29` (type 29)
Same columns and field mapping as `tx_employee_accounting_doc_26`.

### Card status change
`tx_card_status_change_27` (type 27)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 card_number, 9 card_type_code, 10 card_type, 11 reserved_11,
12 reserved_12, 13 operation_type, 14 shift_number, 15 promotion_code, 16 event_code, 17 reserved_17,
18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21, 22 reserved_22, 23 document_type_code,
24 reserved_24, 25 reserved_25, 26 document_info, 27 enterprise_id, 28 reserved_28, 29 reserved_29,
30 reserved_30, 31 old_card_status, 32 new_card_status, 33 new_valid_from, 34 new_valid_to,
35 reserved_35, 36 reserved_36.

### Modifiers
`tx_modifier_registration_30` (type 30)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 item_identifier, 9 reserved_9, 10 reserved_10, 11 item_quantity,
12 reserved_12, 13 operation_type, 14 shift_number, 15 reserved_15, 16 reserved_16,
17 document_print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21,
22 reserved_22, 23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info,
27 enterprise_id, 28 reserved_28, 29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32,
33 reserved_33, 34 reserved_34, 35 modifier_code, 36 reserved_36.

`tx_modifier_storno_31` (type 31)
Same columns and field mapping as `tx_modifier_registration_30`.

### Bonus payments
`tx_bonus_payment_32` (type 32)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 bonus_card_number, 9 reserved_9, 10 bonus_payment_type,
11 counter_change_amount, 12 payment_amount, 13 operation_type, 14 shift_number, 15 promotion_code,
16 event_code, 17 print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20,
21 counter_type_code, 22 counter_code, 23 document_type_code, 24 reserved_24, 25 reserved_25,
26 document_info, 27 enterprise_id, 28 reserved_28, 29 ps_protocol_number, 30 reserved_30,
31 reserved_31, 32 reserved_32, 33 reserved_33, 34 reserved_34, 35 reserved_35, 36 reserved_36.

`tx_bonus_payment_33` (type 33)
Same columns and field mapping as `tx_bonus_payment_32`.

`tx_bonus_payment_82` (type 82)
Same columns and field mapping as `tx_bonus_payment_32`.

`tx_bonus_payment_83` (type 83)
Same columns and field mapping as `tx_bonus_payment_32`.

### Prepayment
`tx_prepayment_34` (type 34)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 reserved_8, 9 reserved_9, 10 prepayment_type, 11 reserved_11,
12 prepayment_amount, 13 operation_type, 14 shift_number, 15 reserved_15, 16 reserved_16,
17 print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21, 22 reserved_22,
23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info, 27 enterprise_id,
28 reserved_28, 29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32, 33 reserved_33,
34 reserved_34, 35 reserved_35, 36 reserved_36.

`tx_prepayment_84` (type 84)
Same columns and field mapping as `tx_prepayment_34`.

### Document discounts / rounding
`tx_document_discount_35` (type 35)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 discount_info, 9 reserved_9, 10 discount_type, 11 discount_value,
12 discount_amount_base, 13 operation_type, 14 shift_number, 15 promotion_code, 16 event_code,
17 print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21, 22 reserved_22,
23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info, 27 enterprise_id,
28 reserved_28, 29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32, 33 reserved_33,
34 reserved_34, 35 reserved_35, 36 reserved_36.

`tx_document_discount_37` (type 37)
Same columns and field mapping as `tx_document_discount_35`.

`tx_document_discount_85` (type 85)
Same columns and field mapping as `tx_document_discount_35`.

`tx_document_discount_87` (type 87)
Same columns and field mapping as `tx_document_discount_35`.

`tx_document_rounding_38` (type 38)
Same columns and field mapping as `tx_document_discount_35`.

### Non-fiscal payment
`tx_non_fiscal_payment_36` (type 36)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 gift_card_number, 9 payment_type_code, 10 payment_type_operation,
11 reserved_11, 12 payment_amount, 13 operation_type, 14 shift_number, 15 promotion_code, 16 event_code,
17 print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 counter_type_code,
22 counter_code, 23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info,
27 enterprise_id, 28 reserved_28, 29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32,
33 reserved_33, 34 reserved_34, 35 reserved_35, 36 reserved_36.

`tx_non_fiscal_payment_86` (type 86)
Same columns and field mapping as `tx_non_fiscal_payment_36`.

### Fiscal payment
`tx_fiscal_payment_40` (type 40)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 card_number, 9 payment_type_code, 10 payment_type_operation,
11 customer_amount_payment_currency, 12 customer_amount_base_currency, 13 operation_type, 14 shift_number,
15 promotion_code, 16 event_code, 17 current_print_group_code, 18 reserved_18, 19 currency_code,
20 cash_out_amount, 21 counter_type_code, 22 counter_code, 23 document_type_code, 24 reserved_24,
25 reserved_25, 26 document_info, 27 enterprise_id, 28 reserved_28, 29 ps_protocol_number,
30 reserved_30, 31 reserved_31, 32 reserved_32, 33 reserved_33, 34 reserved_34, 35 reserved_35,
36 reserved_36.

`tx_fiscal_payment_43` (type 43)
Same columns and field mapping as `tx_fiscal_payment_40`.

### Document operations (open/close/etc.)
`tx_document_open_42` (type 42)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 customer_card_numbers, 9 dimension_value_codes, 10 reserved_10,
11 reserved_11, 12 reserved_12, 13 operation_type, 14 shift_number, 15 customer_code, 16 reserved_16,
17 document_print_group_code, 18 reserved_18, 19 order_id, 20 document_amount_without_discounts,
21 visitor_count, 22 reserved_22, 23 document_type_code, 24 comment_code, 25 base_document_number,
26 document_info, 27 enterprise_id, 28 employee_code, 29 employee_edit_document_number,
30 department_code, 31 hall_code, 32 service_point_code, 33 reservation_id, 34 user_variables,
35 external_comment, 36 revaluation_datetime, 37 contractor_code, 38 subdivision_id, 39 reserved_39,
40 document_coupons, 44 reserved_44.

`tx_document_close_55` (type 55) -> same columns as `tx_document_open_42`.
`tx_document_cancel_56` (type 56) -> same columns as `tx_document_open_42`.
`tx_document_non_fin_close_58` (type 58) -> same columns as `tx_document_open_42`.
`tx_document_clients_65` (type 65) -> same columns as `tx_document_open_42`.
`tx_document_close_kkt_45` (type 45) -> same columns as `tx_document_open_42`.
`tx_document_close_gp_49` (type 49) -> same columns as `tx_document_open_42`.
`tx_document_egais_120` (type 120) -> same columns as `tx_document_open_42`.

### VAT KKT
`tx_vat_kkt_88` (type 88)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 reserved_8, 9 reserved_9, 10 reserved_10, 11 reserved_11,
12 reserved_12, 13 operation_type, 14 shift_number, 15 reserved_15, 16 reserved_16,
17 print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21, 22 reserved_22,
23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info, 27 enterprise_id,
28 vat_0_amount, 29 vat_10_amount, 30 vat_20_amount, 31 no_vat_amount, 32 vat_10_110_amount,
33 vat_20_120_amount, 34 reserved_34, 35 reserved_35, 36 reserved_36, 37 reserved_37, 38 reserved_38,
39 reserved_39, 43 reserved_43.

### Cash in/out
`tx_cash_in_50` (type 50)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 reserved_8, 9 reserved_9, 10 reserved_10, 11 reserved_11,
12 amount_base, 13 operation_type, 14 shift_number, 15 reserved_15, 16 reserved_16,
17 print_group_code, 18 reserved_18, 19 order_id, 20 reserved_20, 21 reserved_21, 22 reserved_22,
23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info, 27 enterprise_id,
28 reserved_28, 29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32, 33 reserved_33,
34 reserved_34, 35 reserved_35, 36 reserved_36.

`tx_cash_out_51` (type 51)
Same columns and field mapping as `tx_cash_in_50`.

### Counter change
`tx_counter_change_57` (type 57)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 card_or_client_code, 9 card_type_code, 10 binding_type,
11 value_after_changes, 12 change_amount, 13 operation_type, 14 shift_number, 15 promotion_code,
16 event_code, 17 reserved_17, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 counter_type_code,
22 counter_code, 23 document_type_code, 24 reserved_24, 25 reserved_25, 26 document_info,
27 enterprise_id, 28 reserved_28, 29 reserved_29, 30 counter_valid_from, 31 reserved_31,
32 reserved_32, 33 card_valid_from, 34 card_valid_to, 35 counter_valid_to, 36 reserved_36.

### Reports / shifts
`tx_report_zless_60` (type 60)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 reserved_8, 9 reserved_9, 10 shift_revenue, 11 cash_in_drawer,
12 shift_income_total, 13 reserved_13, 14 shift_number, 15 reserved_15, 16 reserved_16,
17 print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21,
22 reserved_22, 23 reserved_23, 24 reserved_24, 25 reserved_25, 26 cash_document_number,
27 enterprise_id, 28 reserved_28, 29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32,
33 reserved_33, 34 unreported_docs_count, 35 exchange_error_codes, 36 earliest_unreported_doc_datetime,
44 reserved_44.

`tx_report_z_63` (type 63) -> same columns as `tx_report_zless_60`.
`tx_shift_open_doc_64` (type 64) -> same columns as `tx_report_zless_60`.
`tx_shift_close_61` (type 61) -> same columns as `tx_report_zless_60`.
`tx_shift_open_62` (type 62) -> same columns as `tx_report_zless_60`.

### Mark unit
`tx_mark_unit_121` (type 121)
1 transaction_id_unique, 2 transaction_date, 3 transaction_time, 4 transaction_type, 5 cash_register_code,
6 document_number, 7 cashier_code, 8 reserved_8, 9 reserved_9, 10 reserved_10, 11 reserved_11,
12 reserved_12, 13 operation_type, 14 shift_number, 15 reserved_15, 16 reserved_16,
17 print_group_code, 18 reserved_18, 19 reserved_19, 20 reserved_20, 21 reserved_21, 22 reserved_22,
23 document_type_code, 24 reserved_24, 25 reserved_25, 26 reserved_26, 27 enterprise_id,
28 reserved_28, 29 reserved_29, 30 reserved_30, 31 reserved_31, 32 reserved_32, 33 reserved_33,
34 reserved_34, 35 reserved_35, 36 reserved_36, 39 reserved_39, 40 reserved_40, 41 reserved_41,
42 reserved_42, 43 reserved_43.

### 2) Update models to match new schema
Create new model structs or replace existing ones so that field names match **tx_* schema**.
- Example: replace `TransactionRegistration` with a model for `tx_item_registration_1_11` and similar for `tx_item_storno_2_12`, `tx_item_tax_4_14`, `tx_item_kkt_6_16`.
- Keep `RawData` if needed for audits and export (add column in schema if missing).
- Ensure types match SQL schema: `NUMERIC(18,6)` -> `float64` or `pgtype.Numeric`, `DATE` -> `time.Time`, `TIME` -> `time.Time`.

### 3) Rewrite parsers to emit new models
- Update `pkg/parser/dispatcher.go` and `pkg/parser/mappers.go` to return **tx_* models**.
- For types with multiple tables (1/11, 2/12, 4/14, 6/16), route to the correct model/table based on transaction type.
- Keep `source_folder` behavior intact.
- Parse date/time as `time.Time` for DB insert (avoid string unless schema requires it).
- Preserve `raw_data` (full line) for debugging/export.

### 4) Rewrite DB loaders
- Update `pkg/db/postgres.go` load functions to insert into **tx_* tables**.
- Create `LoadTxItemRegistration_1_11`, `LoadTxItemStorno_2_12`, etc. with proper column lists.
- Update `pkg/repository/loader.go` switch cases to use new tables and new model types.
- Maintain retry logic and batching.

### 5) Update DB export/unload logic
- Update `GetAllTransactionsBySourceFolderAndDate` to UNION the **tx_* tables**.
- If webhook expects old column names, provide a new `TransactionRow` view that reads from tx_* with aliases:
  - `transaction_id_unique`, `source_folder`, `transaction_date`, `transaction_time`, `transaction_type`, `raw_data`.
- Update other read APIs (`GetTransactionRegistrationsBySourceFolderAndDate`, etc.) or remove if obsolete.

### 6) Update mapping docs/tests
- Replace `PARSERS_TO_TABLES_MAPPING.md` with tx_* mapping.
- Update or add tests to cover:
  - Parsing `data/response.txt` samples for at least 1-2 transaction types per category.
  - Loader inserts for tx_* tables.
  - Export/unload query returns rows sorted by `transaction_time` and `transaction_id_unique`.

---

## Suggested mapping skeleton (to expand)
- 1/11: `tx_item_registration_1_11`
- 2/12: `tx_item_storno_2_12`
- 4/14: `tx_item_tax_4_14`
- 6/16: `tx_item_kkt_6_16`
- 3: (special price) -> find tx_* table in migration and map fields
- 9/10: (bonus) -> find tx_* table in migration and map fields
- 15/17: (discount position) -> find tx_* table in migration and map fields
- 21/23, 22/24: (bill registration) -> find tx_* table in migration and map fields
- 25, 26, 29, 27, 30/31, 32/33/82/83, 34/84, 35/37/38/85/87, 36/86, 40/43, 42/45/49/55/56/58/65/120, 50/51, 57, 60/61/62/63/64, 88, 121: map each to its tx_* table in migration.

---

## Acceptance criteria
- All parser outputs align with **tx_* table columns**.
- DB loads succeed against current schema with no missing columns.
- Export/unload endpoints read from tx_* tables without SQL errors.
- Updated docs and tests reflect the new schema.
