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
    used_amount   BIGINT NOT NULL DEFAULT 0,
    request_count BIGINT NOT NULL DEFAULT 0,
    group_name    VARCHAR(64) NOT NULL DEFAULT 'default',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN users.role IS '1:普通用户 10:管理员';
COMMENT ON COLUMN users.status IS '1:正常 2:禁用';
COMMENT ON COLUMN users.balance IS '余额,单位:0.001元';

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
    model           VARCHAR(128) NOT NULL,
    request_method  VARCHAR(16) NOT NULL,
    request_path    VARCHAR(256) NOT NULL,
    status_code     INT,
    input_tokens    INT NOT NULL DEFAULT 0,
    output_tokens   INT NOT NULL DEFAULT 0,
    cost            BIGINT NOT NULL DEFAULT 0,
    latency_ms      INT NOT NULL DEFAULT 0,
    error_message   TEXT,
    ip_address      VARCHAR(45),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_request_logs_user_time ON request_logs(user_id, created_at DESC);
CREATE INDEX idx_request_logs_created ON request_logs(created_at DESC);
CREATE INDEX idx_request_logs_model ON request_logs(model);

COMMENT ON COLUMN request_logs.cost IS '本次费用,单位:0.001元';
COMMENT ON COLUMN request_logs.latency_ms IS '响应延迟,毫秒';

-- 插入默认管理员账户 (密码: admin123, bcrypt哈希)
-- 默认管理员 密码: Admin@2026!
INSERT INTO users (username, email, password_hash, role, balance)
VALUES ('admin', 'admin@aitoken.dev', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 10, 999999000);

-- 插入初始模型数据
INSERT INTO models (model_id, display_name, provider, category, context_length, input_price, output_price, price_unit, capabilities, description) VALUES
('gpt-4o', 'GPT-4o', 'openai', 'llm', 128000, 2500, 10000, '1M', ARRAY['vision','function_call','streaming','json_mode'], '最新 GPT-4o 多模态模型'),
('gpt-4o-mini', 'GPT-4o Mini', 'openai', 'llm', 128000, 150, 600, '1M', ARRAY['vision','function_call','streaming','json_mode'], '轻量级 GPT-4o,性价比高'),
('claude-sonnet-4-6', 'Claude Sonnet 4.6', 'anthropic', 'llm', 200000, 3000, 15000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Anthropic 最新编程模型'),
('claude-haiku-4-5', 'Claude Haiku 4.5', 'anthropic', 'llm', 200000, 800, 4000, '1M', ARRAY['vision','function_call','streaming'], 'Anthropic 轻量快速模型'),
('gemini-2.5-pro', 'Gemini 2.5 Pro', 'google', 'llm', 1000000, 1250, 10000, '1M', ARRAY['vision','function_call','streaming','thinking'], 'Google 最新旗舰模型,100万上下文'),
('gemini-2.5-flash', 'Gemini 2.5 Flash', 'google', 'llm', 1000000, 150, 600, '1M', ARRAY['vision','function_call','streaming'], 'Google 快速模型'),
('deepseek-chat', 'DeepSeek V3', 'deepseek', 'llm', 64000, 270, 1100, '1M', ARRAY['function_call','streaming','json_mode'], 'DeepSeek 通用对话模型'),
('deepseek-reasoner', 'DeepSeek R1', 'deepseek', 'llm', 64000, 550, 2190, '1M', ARRAY['streaming','thinking'], 'DeepSeek 深度推理模型');
