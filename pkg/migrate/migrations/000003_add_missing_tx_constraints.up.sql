DO $$
DECLARE
  rec record;
  constraint_name text;
BEGIN
  FOR rec IN
    SELECT t.table_name
    FROM information_schema.tables t
    WHERE t.table_schema = 'public'
      AND t.table_type = 'BASE TABLE'
      AND t.table_name LIKE 'tx_%'
      AND EXISTS (
        SELECT 1
        FROM information_schema.columns c
        WHERE c.table_schema = t.table_schema
          AND c.table_name = t.table_name
          AND c.column_name = 'transaction_id_unique'
      )
      AND EXISTS (
        SELECT 1
        FROM information_schema.columns c
        WHERE c.table_schema = t.table_schema
          AND c.table_name = t.table_name
          AND c.column_name = 'source_folder'
      )
  LOOP
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.table_constraints tc
      JOIN information_schema.key_column_usage kcu
        ON tc.constraint_name = kcu.constraint_name
       AND tc.table_schema = kcu.table_schema
      WHERE tc.table_schema = 'public'
        AND tc.table_name = rec.table_name
        AND tc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
        AND kcu.column_name IN ('transaction_id_unique', 'source_folder')
      GROUP BY tc.constraint_name
      HAVING COUNT(DISTINCT kcu.column_name) = 2
    ) THEN
      constraint_name := 'pk_fix__' || rec.table_name;
      EXECUTE format(
        'ALTER TABLE %I ADD CONSTRAINT %I PRIMARY KEY (transaction_id_unique, source_folder)',
        rec.table_name,
        constraint_name
      );
    END IF;
  END LOOP;
END $$;
