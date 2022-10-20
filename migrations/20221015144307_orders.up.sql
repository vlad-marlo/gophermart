CREATE TABLE IF NOT EXISTS orders(
    pk BIGSERIAL PRIMARY KEY,
    id INT UNIQUE, -- number in JSON obj
    user_id BIGINT,
    status VARCHAR(50) DEFAULT 'NEW',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accrual FLOAT
);

ALTER TABLE IF EXISTS
    orders
ADD CONSTRAINT
    fk_user_order
FOREIGN KEY (user_id) REFERENCES users(id);

CREATE UNIQUE INDEX IF NOT EXISTS
    index_user_id_orders
ON orders(user_id);
