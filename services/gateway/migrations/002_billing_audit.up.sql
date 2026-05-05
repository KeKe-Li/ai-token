-- 为请求日志增加计费审计字段
ALTER TABLE request_logs
    ADD COLUMN IF NOT EXISTS billing_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS billing_note TEXT,
    ADD COLUMN IF NOT EXISTS estimated_tokens BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS reserved_amount BIGINT NOT NULL DEFAULT 0;

COMMENT ON COLUMN request_logs.billing_status IS '计费状态: charged/charge_failed/unpriced/zero_cost/no_user';
COMMENT ON COLUMN request_logs.billing_note IS '计费说明或失败原因';
COMMENT ON COLUMN request_logs.estimated_tokens IS '是否使用估算 token 计费';
COMMENT ON COLUMN request_logs.reserved_amount IS '请求前预授权估算金额,单位:0.001元';
