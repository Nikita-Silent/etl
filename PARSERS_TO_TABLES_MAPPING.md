# Соответствие парсеров и таблиц базы данных

**Дата:** 2025-01-XX
**Цель:** Зафиксировать соответствие типов транзакций Frontol и таблиц `tx_*`.

---

## Краткая таблица соответствия

| Типы транзакций | Таблица БД |
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

---

## Примечания

- Маппинг полей для каждого `tx_*` задан в `pkg/migrate/migrations/000001_init_schema.up.sql` и `docs/frontol_6_integration.md`.
- Парсер использует таблицу как ключ результата (например, `tx_item_registration_1_11`).
- `source_folder` добавляется системой и не приходит из файла.
