CREATE TABLE IF NOT EXISTS withdrawals(
    id BIGSERIAL UNIQUE PRIMARY KEY,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id BIGINT,
    order_id BIGINT,
    order_sum INT NOT NULL
);

ALTER TABLE IF EXISTS 
    withdrawals
ADD CONSTRAINT
    fk_user_withdraw
FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE IF EXISTS 
    withdrawals
ADD CONSTRAINT
    fk_order_withdraw
FOREIGN KEY (order_id) REFERENCES orders(id);
