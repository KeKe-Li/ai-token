-- 增加强一致性预授权 hold 与不可变账务事件
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS reserved_balance BIGINT NOT NULL DEFAULT 0;

COMMENT ON COLUMN users.reserved_balance IS '已预授权但尚未结算的冻结余额,单位:0.001元';

ALTER TABLE request_logs
    ADD COLUMN IF NOT EXISTS wallet_hold_id BIGINT;

COMMENT ON COLUMN request_logs.wallet_hold_id IS '关联的钱包预授权 hold ID';

CREATE TABLE IF NOT EXISTS wallet_holds (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id      BIGINT,
    request_log_id  BIGINT REFERENCES request_logs(id) ON DELETE SET NULL,
    model           VARCHAR(128),
    amount          BIGINT NOT NULL,
    captured_amount BIGINT NOT NULL DEFAULT 0,
    released_amount BIGINT NOT NULL DEFAULT 0,
    status          VARCHAR(32) NOT NULL,
    reason          VARCHAR(64) NOT NULL DEFAULT 'api_request',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_wallet_holds_user_time ON wallet_holds(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_holds_status ON wallet_holds(status);
CREATE INDEX IF NOT EXISTS idx_wallet_holds_request_log ON wallet_holds(request_log_id);

COMMENT ON TABLE wallet_holds IS '钱包预授权 hold,用于请求前冻结可用余额并在结算时 capture/release';
COMMENT ON COLUMN wallet_holds.amount IS '预授权金额,单位:0.001元';
COMMENT ON COLUMN wallet_holds.status IS 'held/captured/released/failed';

CREATE TABLE IF NOT EXISTS billing_events (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id        BIGINT,
    request_log_id    BIGINT REFERENCES request_logs(id) ON DELETE SET NULL,
    wallet_hold_id    BIGINT REFERENCES wallet_holds(id) ON DELETE SET NULL,
    event_type        VARCHAR(64) NOT NULL,
    model             VARCHAR(128),
    amount            BIGINT NOT NULL DEFAULT 0,
    balance           BIGINT NOT NULL DEFAULT 0,
    reserved_balance  BIGINT NOT NULL DEFAULT 0,
    available_balance BIGINT NOT NULL DEFAULT 0,
    status            VARCHAR(32) NOT NULL,
    note              TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_billing_events_user_time ON billing_events(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_billing_events_type ON billing_events(event_type);
CREATE INDEX IF NOT EXISTS idx_billing_events_hold ON billing_events(wallet_hold_id);

COMMENT ON TABLE billing_events IS '不可变账务事件,记录预授权、capture、release、失败等非余额流水事件';
