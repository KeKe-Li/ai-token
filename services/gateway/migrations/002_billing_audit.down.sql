ALTER TABLE request_logs
    DROP COLUMN IF EXISTS reserved_amount,
    DROP COLUMN IF EXISTS estimated_tokens,
    DROP COLUMN IF EXISTS billing_note,
    DROP COLUMN IF EXISTS billing_status;
