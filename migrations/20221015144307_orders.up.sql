CREATE TABLE IF NOT EXISTS orders(
    pk BIGSERIAL PRIMARY KEY,
    id VARCHAR(50), -- number in JSON obj
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'NEW',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accrual FLOAT
);
