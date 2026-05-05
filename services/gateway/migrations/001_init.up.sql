-- AI Token 数据库初始化迁移
-- 创建时间: 2026-04-30

-- 用户表
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(64) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          SMALLINT NOT NULL DEFAULT 1,
    status        SMALLINT NOT NULL DEFAULT 1,
    balance       BIGINT NOT NULL DEFAULT 0,
    reserved_balance BIGINT NOT NULL DEFAULT 0,
    used_amount   BIGINT NOT NULL DEFAULT 0,
    request_count BIGINT NOT NULL DEFAULT 0,
    group_name    VARCHAR(64) NOT NULL DEFAULT 'default',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN users.role IS '1:普通用户 10:管理员';
COMMENT ON COLUMN users.status IS '1:正常 2:禁用';
COMMENT ON COLUMN users.balance IS '余额,单位:0.001元';
COMMENT ON COLUMN users.reserved_balance IS '已预授权但尚未结算的冻结余额,单位:0.001元';

-- API 密钥表
CREATE TABLE api_keys (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(128) NOT NULL,
    key_hash    VARCHAR(64) NOT NULL UNIQUE,
    key_prefix  VARCHAR(12) NOT NULL,
    status      SMALLINT NOT NULL DEFAULT 1,
    models      TEXT[],
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);

COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 哈希,明文仅创建时返回一次';
COMMENT ON COLUMN api_keys.key_prefix IS 'sk- + 前8位,用于列表展示';
COMMENT ON COLUMN api_keys.status IS '1:正常 2:禁用';

-- 供应商渠道表
CREATE TABLE channels (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    provider        VARCHAR(32) NOT NULL,
    base_url        VARCHAR(512) NOT NULL,
    api_key_enc     TEXT NOT NULL,
    models          TEXT[] NOT NULL,
    status          SMALLINT NOT NULL DEFAULT 1,
    priority        INT NOT NULL DEFAULT 0,
    weight          INT NOT NULL DEFAULT 1,
    rate_limit      INT NOT NULL DEFAULT 0,
    cooldown_until  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channels_provider ON channels(provider);
CREATE INDEX idx_channels_status ON channels(status);

COMMENT ON COLUMN channels.provider IS 'openai/anthropic/google/deepseek';
COMMENT ON COLUMN channels.api_key_enc IS 'AES-256-GCM 加密的供应商密钥';
COMMENT ON COLUMN channels.status IS '1:启用 2:禁用 3:冷却中';
COMMENT ON COLUMN channels.priority IS '优先级,数字越大越优先';
COMMENT ON COLUMN channels.weight IS '同优先级内的加权随机权重';

-- 模型配置表
CREATE TABLE models (
    id              BIGSERIAL PRIMARY KEY,
    model_id        VARCHAR(128) UNIQUE NOT NULL,
    display_name    VARCHAR(128) NOT NULL,
    provider        VARCHAR(32) NOT NULL,
    category        VARCHAR(32) NOT NULL,
    context_length  INT NOT NULL DEFAULT 0,
    input_price     BIGINT NOT NULL DEFAULT 0,
    output_price    BIGINT NOT NULL DEFAULT 0,
    price_unit      VARCHAR(16) NOT NULL DEFAULT '1K',
    capabilities    TEXT[],
    status          SMALLINT NOT NULL DEFAULT 1,
    icon_url        VARCHAR(512),
    description     TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_models_provider ON models(provider);
CREATE INDEX idx_models_category ON models(category);
CREATE INDEX idx_models_status ON models(status);

COMMENT ON COLUMN models.model_id IS '模型标识,如 gpt-4o, claude-sonnet-4-6';
COMMENT ON COLUMN models.category IS 'llm/embedding/image/audio/video';
COMMENT ON COLUMN models.input_price IS '输入价格,单位:0.001元/price_unit tokens';
COMMENT ON COLUMN models.output_price IS '输出价格,单位:0.001元/price_unit tokens';
COMMENT ON COLUMN models.status IS '1:可用 2:下线';

-- 请求日志表
CREATE TABLE request_logs (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    api_key_id      BIGINT NOT NULL,
    channel_id      BIGINT,
    wallet_hold_id  BIGINT,
    model           VARCHAR(128) NOT NULL,
    request_method  VARCHAR(16) NOT NULL,
    request_path    VARCHAR(256) NOT NULL,
    status_code     INT,
    input_tokens    INT NOT NULL DEFAULT 0,
    output_tokens   INT NOT NULL DEFAULT 0,
    cost            BIGINT NOT NULL DEFAULT 0,
    reserved_amount BIGINT NOT NULL DEFAULT 0,
    latency_ms      INT NOT NULL DEFAULT 0,
    error_message   TEXT,
    ip_address      VARCHAR(45),
    billing_status  VARCHAR(32) NOT NULL DEFAULT 'pending',
    billing_note    TEXT,
    estimated_tokens BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_request_logs_user_time ON request_logs(user_id, created_at DESC);
CREATE INDEX idx_request_logs_created ON request_logs(created_at DESC);
CREATE INDEX idx_request_logs_model ON request_logs(model);

COMMENT ON COLUMN request_logs.cost IS '本次费用,单位:0.001元';
COMMENT ON COLUMN request_logs.wallet_hold_id IS '关联的钱包预授权 hold ID';
COMMENT ON COLUMN request_logs.reserved_amount IS '请求前预授权估算金额,单位:0.001元';
COMMENT ON COLUMN request_logs.latency_ms IS '响应延迟,毫秒';
COMMENT ON COLUMN request_logs.billing_status IS '计费状态: charged/charge_failed/unpriced/zero_cost/no_user';
COMMENT ON COLUMN request_logs.billing_note IS '计费说明或失败原因';
COMMENT ON COLUMN request_logs.estimated_tokens IS '是否使用估算 token 计费';

-- 钱包预授权 hold 表
CREATE TABLE wallet_holds (
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

CREATE INDEX idx_wallet_holds_user_time ON wallet_holds(user_id, created_at DESC);
CREATE INDEX idx_wallet_holds_status ON wallet_holds(status);
CREATE INDEX idx_wallet_holds_request_log ON wallet_holds(request_log_id);

COMMENT ON TABLE wallet_holds IS '钱包预授权 hold,用于请求前冻结可用余额并在结算时 capture/release';
COMMENT ON COLUMN wallet_holds.amount IS '预授权金额,单位:0.001元';
COMMENT ON COLUMN wallet_holds.status IS 'held/captured/released/failed';

-- 账务事件表
CREATE TABLE billing_events (
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

CREATE INDEX idx_billing_events_user_time ON billing_events(user_id, created_at DESC);
CREATE INDEX idx_billing_events_type ON billing_events(event_type);
CREATE INDEX idx_billing_events_hold ON billing_events(wallet_hold_id);

COMMENT ON TABLE billing_events IS '不可变账务事件,记录预授权、capture、release、失败等非余额流水事件';

-- 钱包流水表
CREATE TABLE wallet_transactions (
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

CREATE INDEX idx_wallet_transactions_user_time ON wallet_transactions(user_id, created_at DESC);
CREATE INDEX idx_wallet_transactions_type ON wallet_transactions(type);
CREATE INDEX idx_wallet_transactions_request_log ON wallet_transactions(request_log_id);

COMMENT ON TABLE wallet_transactions IS '钱包余额变动流水';
COMMENT ON COLUMN wallet_transactions.type IS '流水类型: api_charge/api_charge_failed/admin_adjustment/recharge';
COMMENT ON COLUMN wallet_transactions.amount IS '变动金额,单位:0.001元; 消费为负数';
COMMENT ON COLUMN wallet_transactions.balance_before IS '变动前余额快照,单位:0.001元';
COMMENT ON COLUMN wallet_transactions.balance_after IS '变动后余额快照,单位:0.001元';

-- 插入默认管理员账户 (密码: admin123, bcrypt哈希)
-- 默认管理员 密码: Admin@2026!
INSERT INTO users (username, email, password_hash, role, balance)
VALUES ('admin', 'admin@aitoken.dev', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 10, 999999000);

-- 插入初始模型数据
INSERT INTO models (model_id, display_name, provider, category, context_length, input_price, output_price, price_unit, capabilities, description) VALUES
-- OpenAI GPT-5 系列
('gpt-5.5', 'GPT-5.5', 'openai', 'llm', 1000000, 2500, 10000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'OpenAI 最新旗舰模型'),
('gpt-5.4', 'GPT-5.4', 'openai', 'llm', 1000000, 2500, 10000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'OpenAI 旗舰模型'),
('gpt-5.3-codex', 'GPT-5.3 Codex', 'openai', 'llm', 256000, 2000, 8000, '1M', ARRAY['function_call','streaming','thinking'], 'OpenAI Codex 编程模型'),
('gpt-5.2', 'GPT-5.2', 'openai', 'llm', 256000, 2000, 8000, '1M', ARRAY['vision','function_call','streaming'], 'OpenAI 高级模型'),
('gpt-5.1', 'GPT-5.1', 'openai', 'llm', 256000, 2000, 8000, '1M', ARRAY['vision','function_call','streaming'], 'OpenAI 高级模型'),
('gpt-5', 'GPT-5', 'openai', 'llm', 256000, 2000, 8000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'OpenAI GPT-5'),
('gpt-5-codex', 'GPT-5 Codex', 'openai', 'llm', 256000, 2000, 8000, '1M', ARRAY['function_call','streaming','thinking'], 'GPT-5 Codex 编程专用'),
('gpt-5-mini', 'GPT-5 Mini', 'openai', 'llm', 256000, 300, 1200, '1M', ARRAY['vision','function_call','streaming'], 'GPT-5 轻量版'),
('gpt-5-nano', 'GPT-5 Nano', 'openai', 'llm', 128000, 100, 400, '1M', ARRAY['function_call','streaming'], 'GPT-5 超轻量版'),
-- OpenAI GPT-4 系列
('gpt-4o', 'GPT-4o', 'openai', 'llm', 128000, 2500, 10000, '1M', ARRAY['vision','function_call','streaming','json_mode'], 'GPT-4o 多模态模型'),
('gpt-4o-mini', 'GPT-4o Mini', 'openai', 'llm', 128000, 150, 600, '1M', ARRAY['vision','function_call','streaming','json_mode'], '轻量级 GPT-4o'),
('gpt-4.1', 'GPT-4.1', 'openai', 'llm', 1000000, 2000, 8000, '1M', ARRAY['vision','function_call','streaming'], 'GPT-4.1 长上下文'),
('gpt-4.1-mini', 'GPT-4.1 Mini', 'openai', 'llm', 1000000, 400, 1600, '1M', ARRAY['vision','function_call','streaming'], 'GPT-4.1 轻量版'),
('gpt-4.1-nano', 'GPT-4.1 Nano', 'openai', 'llm', 1000000, 100, 400, '1M', ARRAY['function_call','streaming'], 'GPT-4.1 超轻量版'),
('gpt-4', 'GPT-4', 'openai', 'llm', 128000, 30000, 60000, '1M', ARRAY['vision','function_call','streaming'], '经典 GPT-4'),
('gpt-4-turbo', 'GPT-4 Turbo', 'openai', 'llm', 128000, 10000, 30000, '1M', ARRAY['vision','function_call','streaming','json_mode'], 'GPT-4 Turbo'),
('gpt-3.5-turbo', 'GPT-3.5 Turbo', 'openai', 'llm', 16385, 500, 1500, '1M', ARRAY['function_call','streaming'], '经典 GPT-3.5'),
-- OpenAI 推理系列
('o3', 'o3', 'openai', 'llm', 200000, 10000, 40000, '1M', ARRAY['streaming','thinking'], 'OpenAI o3 推理模型'),
('o3-mini', 'o3 Mini', 'openai', 'llm', 200000, 1100, 4400, '1M', ARRAY['streaming','thinking'], 'o3 轻量推理'),
('o4-mini', 'o4 Mini', 'openai', 'llm', 200000, 1100, 4400, '1M', ARRAY['streaming','thinking'], 'o4 轻量推理'),
-- Anthropic Claude Opus 系列
('claude-opus-4-7', 'Claude Opus 4.7', 'anthropic', 'llm', 200000, 15000, 75000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Anthropic 最强旗舰模型'),
('claude-opus-4-6', 'Claude Opus 4.6', 'anthropic', 'llm', 200000, 15000, 75000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Claude Opus 4.6'),
('claude-opus-4-5-20251101', 'Claude Opus 4.5', 'anthropic', 'llm', 200000, 15000, 75000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Claude Opus 4.5'),
('claude-opus-4-20250514', 'Claude Opus 4', 'anthropic', 'llm', 200000, 15000, 75000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Claude Opus 4'),
('claude-opus-4-1-20250805', 'Claude Opus 4.1', 'anthropic', 'llm', 200000, 15000, 75000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Claude Opus 4.1'),
-- Anthropic Claude Sonnet 系列
('claude-sonnet-4-6', 'Claude Sonnet 4.6', 'anthropic', 'llm', 200000, 3000, 15000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Anthropic 最新编程模型'),
('claude-sonnet-4-5-20250929', 'Claude Sonnet 4.5', 'anthropic', 'llm', 200000, 3000, 15000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Claude Sonnet 4.5'),
('claude-sonnet-4-20250514', 'Claude Sonnet 4', 'anthropic', 'llm', 200000, 3000, 15000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Claude Sonnet 4'),
('claude-3-7-sonnet-20250219', 'Claude 3.7 Sonnet', 'anthropic', 'llm', 200000, 3000, 15000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Claude 3.7 Sonnet'),
('claude-3-5-sonnet-20241022', 'Claude 3.5 Sonnet', 'anthropic', 'llm', 200000, 3000, 15000, '1M', ARRAY['vision','function_call','streaming'], 'Claude 3.5 Sonnet'),
-- Anthropic Claude Haiku 系列
('claude-haiku-4-5', 'Claude Haiku 4.5', 'anthropic', 'llm', 200000, 800, 4000, '1M', ARRAY['vision','function_call','streaming'], 'Anthropic 轻量快速模型'),
('claude-haiku-4-5-20251001', 'Claude Haiku 4.5 (Oct)', 'anthropic', 'llm', 200000, 800, 4000, '1M', ARRAY['vision','function_call','streaming'], 'Claude Haiku 4.5 稳定版'),
('claude-3-5-haiku-20241022', 'Claude 3.5 Haiku', 'anthropic', 'llm', 200000, 250, 1250, '1M', ARRAY['vision','streaming'], 'Claude 3.5 Haiku'),
-- DeepSeek
('deepseek-chat', 'DeepSeek V3', 'deepseek', 'llm', 64000, 270, 1100, '1M', ARRAY['function_call','streaming','json_mode'], 'DeepSeek 通用对话模型'),
('deepseek-reasoner', 'DeepSeek R1', 'deepseek', 'llm', 64000, 550, 2190, '1M', ARRAY['streaming','thinking'], 'DeepSeek 深度推理模型'),
-- Google
('gemini-2.5-pro', 'Gemini 2.5 Pro', 'google', 'llm', 1000000, 1250, 10000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Google 旗舰模型,100万上下文'),
('gemini-2.5-flash', 'Gemini 2.5 Flash', 'google', 'llm', 1000000, 150, 600, '1M', ARRAY['vision','function_call','streaming'], 'Google 快速模型');
