DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
      ON tc.constraint_name = kcu.constraint_name
     AND tc.table_schema = kcu.table_schema
    WHERE tc.table_schema = 'public'
      AND tc.table_name = 'tx_document_rounding_38'
      AND tc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
      AND kcu.column_name IN ('transaction_id_unique', 'source_folder')
    GROUP BY tc.constraint_name, tc.constraint_type
    HAVING COUNT(DISTINCT kcu.column_name) = 2
  ) THEN
    ALTER TABLE tx_document_rounding_38
      ADD CONSTRAINT tx_document_rounding_38_pkey
      PRIMARY KEY (transaction_id_unique, source_folder);
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
      ON tc.constraint_name = kcu.constraint_name
     AND tc.table_schema = kcu.table_schema
    WHERE tc.table_schema = 'public'
      AND tc.table_name = 'tx_employee_accounting_pos_29'
      AND tc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
      AND kcu.column_name IN ('transaction_id_unique', 'source_folder')
    GROUP BY tc.constraint_name, tc.constraint_type
    HAVING COUNT(DISTINCT kcu.column_name) = 2
  ) THEN
    ALTER TABLE tx_employee_accounting_pos_29
      ADD CONSTRAINT tx_employee_accounting_pos_29_pkey
      PRIMARY KEY (transaction_id_unique, source_folder);
  END IF;
END $$;
