CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status VARCHAR(32) NOT NULL,
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
