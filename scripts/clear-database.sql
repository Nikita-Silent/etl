-- Скрипт очистки базы данных
-- Удаляет все данные из таблиц транзакций, сохраняя справочники
-- Использование: psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f scripts/clear-database.sql

-- Отключаем проверку внешних ключей для ускорения
SET session_replication_role = 'replica';

-- Очистка всех таблиц транзакций (в порядке зависимостей)
TRUNCATE TABLE transaction_registrations CASCADE;
TRUNCATE TABLE special_prices CASCADE;
TRUNCATE TABLE bonus_transactions CASCADE;
TRUNCATE TABLE discount_transactions CASCADE;
TRUNCATE TABLE bill_registrations CASCADE;
TRUNCATE TABLE employee_edits CASCADE;
TRUNCATE TABLE employee_accounting CASCADE;
TRUNCATE TABLE card_status_changes CASCADE;
TRUNCATE TABLE modifier_transactions CASCADE;
TRUNCATE TABLE bonus_payments CASCADE;
TRUNCATE TABLE prepayment_transactions CASCADE;
TRUNCATE TABLE document_discounts CASCADE;
TRUNCATE TABLE non_fiscal_payments CASCADE;
TRUNCATE TABLE fiscal_payments CASCADE;
TRUNCATE TABLE document_operations CASCADE;
TRUNCATE TABLE vat_kkt_transactions CASCADE;
TRUNCATE TABLE additional_transactions CASCADE;
TRUNCATE TABLE astu_exchange_transactions CASCADE;
TRUNCATE TABLE counter_change_transactions CASCADE;
TRUNCATE TABLE kkt_shift_reports CASCADE;
TRUNCATE TABLE frontol_mark_unit_transactions CASCADE;

-- Включаем обратно проверку внешних ключей
SET session_replication_role = 'origin';

-- Выводим статистику
SELECT 
    'transaction_registrations' as table_name, 
    COUNT(*) as remaining_rows 
FROM transaction_registrations
UNION ALL
SELECT 'fiscal_payments', COUNT(*) FROM fiscal_payments
UNION ALL
SELECT 'document_operations', COUNT(*) FROM document_operations;
