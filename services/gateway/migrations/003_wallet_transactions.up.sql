-- 新增钱包流水表，用于审计 API 扣费成功/失败后的余额变化
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_log_id  BIGINT REFERENCES request_logs(id) ON DELETE SET NULL,
    type            VARCHAR(32) NOT NULL,
    amount          BIGINT NOT NULL,
    balance_before  BIGINT,
    balance_after   BIGINT,
    reference_type  VARCHAR(32),
    reference_id    VARCHAR(128),
    note            TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_user_time ON wallet_transactions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_type ON wallet_transactions(type);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_request_log ON wallet_transactions(request_log_id);

COMMENT ON TABLE wallet_transactions IS '钱包余额变动流水';
COMMENT ON COLUMN wallet_transactions.type IS '流水类型: api_charge/api_charge_failed/admin_adjustment/recharge';
COMMENT ON COLUMN wallet_transactions.amount IS '变动金额,单位:0.001元; 消费为负数';
COMMENT ON COLUMN wallet_transactions.balance_before IS '变动前余额快照,单位:0.001元';
COMMENT ON COLUMN wallet_transactions.balance_after IS '变动后余额快照,单位:0.001元';
