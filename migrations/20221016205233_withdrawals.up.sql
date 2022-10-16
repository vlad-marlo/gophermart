CREATE TABLE IF NOT EXISTS withdrawals(
    id BIGSERIAL UNIQUE PRIMARY KEY,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    order_sum INT NOT NULL
);
