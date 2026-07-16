-- 管理员线下提现扣减返利红包。
-- withdraw 动作只扣 user_affiliates.aff_quota，不影响 users.balance 和历史累计返利。

ALTER TABLE user_affiliate_ledger
    ADD COLUMN IF NOT EXISTS operator_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE user_affiliate_ledger
    ADD COLUMN IF NOT EXISTS remark TEXT NULL;

ALTER TABLE user_affiliate_ledger
    ADD COLUMN IF NOT EXISTS external_ref VARCHAR(128) NULL;

COMMENT ON COLUMN user_affiliate_ledger.operator_user_id IS '执行返利提现/人工调整的管理员用户 ID';
COMMENT ON COLUMN user_affiliate_ledger.remark IS '返利提现或人工调整备注';
COMMENT ON COLUMN user_affiliate_ledger.external_ref IS '线下打款单号、微信/支付宝转账号或其它外部凭证';

CREATE INDEX IF NOT EXISTS idx_user_affiliate_ledger_operator_user_id
    ON user_affiliate_ledger(operator_user_id)
    WHERE operator_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_user_affiliate_ledger_withdraw_lookup
    ON user_affiliate_ledger(action, user_id, created_at)
    WHERE action = 'withdraw';
