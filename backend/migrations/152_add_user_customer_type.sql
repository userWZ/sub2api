ALTER TABLE users
    ADD COLUMN IF NOT EXISTS customer_type VARCHAR(20) NOT NULL DEFAULT 'direct';

CREATE INDEX IF NOT EXISTS idx_users_customer_type
    ON users(customer_type)
    WHERE deleted_at IS NULL;
