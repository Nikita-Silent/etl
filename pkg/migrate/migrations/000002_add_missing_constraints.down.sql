DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    WHERE t.relname = 'tx_document_rounding_38'
      AND c.conname = 'tx_document_rounding_38_pkey'
  ) THEN
    ALTER TABLE tx_document_rounding_38
      DROP CONSTRAINT tx_document_rounding_38_pkey;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    WHERE t.relname = 'tx_employee_accounting_pos_29'
      AND c.conname = 'tx_employee_accounting_pos_29_pkey'
  ) THEN
    ALTER TABLE tx_employee_accounting_pos_29
      DROP CONSTRAINT tx_employee_accounting_pos_29_pkey;
  END IF;
END $$;
