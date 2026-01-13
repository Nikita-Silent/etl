-- Проверка отсутствующих транзакций 730423 и 2066483
-- Проверяем во всех таблицах, где могут быть эти записи

-- 1. Проверка в bill_registrations (тип операции 21)
SELECT 
    'bill_registrations' as table_name,
    transaction_id_unique,
    source_folder,
    transaction_date,
    transaction_time,
    transaction_type,
    cash_register_code,
    document_number,
    CASE WHEN raw_data IS NULL OR raw_data = '' THEN 'NO RAW DATA' ELSE 'HAS RAW DATA' END as raw_data_status
FROM bill_registrations
WHERE transaction_id_unique IN (730423, 2066483)
ORDER BY transaction_id_unique;

-- 2. Проверка во всех других таблицах (на всякий случай)
SELECT 
    'transaction_registrations' as table_name,
    transaction_id_unique,
    source_folder,
    transaction_date,
    transaction_time,
    transaction_type
FROM transaction_registrations
WHERE transaction_id_unique IN (730423, 2066483)
UNION ALL
SELECT 
    'document_operations' as table_name,
    transaction_id_unique,
    source_folder,
    transaction_date,
    transaction_time,
    transaction_type
FROM document_operations
WHERE transaction_id_unique IN (730423, 2066483)
UNION ALL
SELECT 
    'special_prices' as table_name,
    transaction_id_unique,
    source_folder,
    transaction_date,
    transaction_time,
    transaction_type
FROM special_prices
WHERE transaction_id_unique IN (730423, 2066483)
ORDER BY transaction_id_unique;

-- 3. Проверка всех записей типа 21 за дату 2025-12-21
SELECT 
    'bill_registrations' as table_name,
    COUNT(*) as count,
    COUNT(CASE WHEN raw_data IS NULL OR raw_data = '' THEN 1 END) as without_raw_data,
    COUNT(CASE WHEN raw_data IS NOT NULL AND raw_data != '' THEN 1 END) as with_raw_data
FROM bill_registrations
WHERE transaction_date = '2025-12-21'::date
  AND transaction_type = 21;

-- 4. Проверка source_folder для записей 730423 и 2066483 (если они есть)
-- Проверяем все возможные source_folder для кассы 123
SELECT DISTINCT
    source_folder,
    COUNT(*) as count
FROM bill_registrations
WHERE transaction_id_unique IN (730423, 2066483)
   OR (transaction_date = '2025-12-21'::date 
       AND transaction_type = 21 
       AND source_folder LIKE '123/%')
GROUP BY source_folder
ORDER BY source_folder;

-- 5. Проверка всех записей типа 21 за дату 2025-12-21 с разными source_folder
SELECT 
    source_folder,
    COUNT(*) as count,
    MIN(transaction_id_unique) as min_id,
    MAX(transaction_id_unique) as max_id
FROM bill_registrations
WHERE transaction_date = '2025-12-21'::date
  AND transaction_type = 21
GROUP BY source_folder
ORDER BY source_folder;

