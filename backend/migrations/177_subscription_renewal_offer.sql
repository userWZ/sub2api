ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS renewal_discount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS renewal_rollover_amount DECIMAL(20,10) NOT NULL DEFAULT 0,
	ADD COLUMN IF NOT EXISTS renewal_source_subscription_id BIGINT,
	ADD COLUMN IF NOT EXISTS renewal_source_expires_at TIMESTAMPTZ;

COMMENT ON COLUMN payment_orders.renewal_discount IS '下单时固化的续订优惠金额（支付币种）';
COMMENT ON COLUMN payment_orders.renewal_rollover_amount IS '支付完成后转入余额账户的未用月额度结余积分';
COMMENT ON COLUMN payment_orders.renewal_source_subscription_id IS '本次续订权益来源订阅 ID';
COMMENT ON COLUMN payment_orders.renewal_source_expires_at IS '本次续订权益来源订阅在下单时的到期时间';

CREATE INDEX IF NOT EXISTS idx_payment_orders_renewal_source_subscription_id
    ON payment_orders(renewal_source_subscription_id)
    WHERE renewal_source_subscription_id IS NOT NULL;
