CREATE TABLE `transactions` (
CREATE TABLE transactions (
    id BIGINT NOT NULL,
    date TIMESTAMP NOT NULL,
    account VARCHAR(64) NOT NULL,
    category VARCHAR(64) NOT NULL,
    amount NUMERIC(18, 2) NOT NULL,
    note TEXT NOT NULL,
    type VARCHAR(16) NOT NULL,
    position JSONB DEFAULT NULL,
    created_at TIMESTAMP NOT NULL,
    created_by VARCHAR(64) DEFAULT NULL,
    updated_at TIMESTAMP DEFAULT NULL,
    updated_by VARCHAR(64) DEFAULT NULL,
    PRIMARY KEY (id)
);