DROP TABLE IF EXISTS billing_events;
DROP TABLE IF EXISTS wallet_holds;

ALTER TABLE request_logs
    DROP COLUMN IF EXISTS wallet_hold_id;

ALTER TABLE users
    DROP COLUMN IF EXISTS reserved_balance;
