-- 首次实际用量邀请奖励。
-- 当前线上没有邀请记录，因此无需历史数据回填；新记录只从本迁移上线后产生。
CREATE TABLE IF NOT EXISTS affiliate_usage_rewards (
    id BIGSERIAL PRIMARY KEY,
    inviter_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    invitee_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount DECIMAL(20,8) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    risk_reason VARCHAR(64) NOT NULL DEFAULT '',
    trigger_request_id VARCHAR(255) NOT NULL DEFAULT '',
    trigger_api_key_id BIGINT NULL REFERENCES api_keys(id) ON DELETE SET NULL,
    trigger_ip VARCHAR(45) NULL,
    trigger_actual_cost DECIMAL(20,8) NOT NULL DEFAULT 0,
    reviewer_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    review_remark VARCHAR(500) NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_affiliate_usage_rewards_invitee UNIQUE (invitee_user_id),
    CONSTRAINT chk_affiliate_usage_rewards_status CHECK (status IN ('pending', 'granted', 'rejected')),
    CONSTRAINT chk_affiliate_usage_rewards_amount CHECK (amount > 0),
    CONSTRAINT chk_affiliate_usage_rewards_cost CHECK (trigger_actual_cost > 0)
);

CREATE INDEX IF NOT EXISTS idx_affiliate_usage_rewards_inviter_created
    ON affiliate_usage_rewards(inviter_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_affiliate_usage_rewards_status_created
    ON affiliate_usage_rewards(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_affiliate_usage_rewards_ip_created
    ON affiliate_usage_rewards(trigger_ip, created_at DESC)
    WHERE trigger_ip IS NOT NULL AND trigger_ip <> '';

COMMENT ON TABLE affiliate_usage_rewards IS '被邀请人首次产生正计费用量后的邀请奖励记录';
COMMENT ON COLUMN affiliate_usage_rewards.amount IS '系统内计费额度积分，不代表实际货币';
COMMENT ON COLUMN affiliate_usage_rewards.risk_reason IS 'auto_granted|missing_ip|same_ip|inviter_daily_limit|inviter_30d_limit|ip_daily_limit|manual_approved|manual_rejected';

INSERT INTO settings (key, value, updated_at) VALUES
    ('affiliate_usage_reward_enabled', 'false', NOW()),
    ('affiliate_usage_reward_amount', '20', NOW()),
    ('affiliate_reward_inviter_daily_limit', '5', NOW()),
    ('affiliate_reward_inviter_30d_limit', '30', NOW()),
    ('affiliate_reward_ip_daily_limit', '2', NOW()),
    ('affiliate_reward_review_same_ip', 'true', NOW()),
    ('affiliate_reward_review_missing_ip', 'true', NOW())
ON CONFLICT (key) DO NOTHING;
