-- 返利钱包购买抵扣。
-- original_amount: 订单折扣前的购买原价（余额包为付款面额，订阅为套餐价格）。
-- affiliate_discount: 本单使用的返利钱包抵扣金额。

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS original_amount DECIMAL(20,2) NOT NULL DEFAULT 0;

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS affiliate_discount DECIMAL(20,2) NOT NULL DEFAULT 0;

COMMENT ON COLUMN payment_orders.original_amount IS '订单折扣前购买原价；历史订单为 0';
COMMENT ON COLUMN payment_orders.affiliate_discount IS '本单使用的返利钱包抵扣金额';

CREATE INDEX IF NOT EXISTS idx_payment_orders_affiliate_discount
    ON payment_orders(affiliate_discount)
    WHERE affiliate_discount > 0;

CREATE INDEX IF NOT EXISTS idx_user_affiliate_ledger_discount_order
    ON user_affiliate_ledger(action, source_order_id, user_id)
    WHERE action IN ('discount', 'discount_restore', 'rebate_reversal');

INSERT INTO settings (key, value)
VALUES
    ('affiliate_discount_enabled', 'true'),
    ('affiliate_discount_max_percent', '50'),
    ('affiliate_discount_min_pay_amount', '1')
ON CONFLICT (key) DO NOTHING;
