-- ============================================================
-- Migration: 013_fix_audit_logs_legacy_id_default
-- Description: Restore legacy_id default to avoid NULL insert failures
-- Created: 2026-01-21
-- ============================================================

DO $$
DECLARE
    seq_name text;
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'audit_logs' AND column_name = 'legacy_id'
    ) THEN
        SELECT pg_get_serial_sequence('audit_logs', 'legacy_id') INTO seq_name;

        IF seq_name IS NOT NULL THEN
            EXECUTE format(
                'ALTER TABLE audit_logs ALTER COLUMN legacy_id SET DEFAULT nextval(%L::regclass)',
                seq_name
            );
        ELSE
            ALTER TABLE audit_logs ALTER COLUMN legacy_id DROP NOT NULL;
        END IF;
    END IF;
END $$;
