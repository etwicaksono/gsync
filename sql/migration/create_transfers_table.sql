CREATE TABLE transfers (
    id BIGINT PRIMARY KEY,
    date TIMESTAMP NOT NULL,
    from_account VARCHAR(64) NOT NULL,
    to_account VARCHAR(64) NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    note TEXT NOT NULL,
    position JSONB DEFAULT NULL,
    created_at TIMESTAMP NOT NULL,
    created_by VARCHAR(64) DEFAULT NULL,
    updated_at TIMESTAMP DEFAULT NULL,
    updated_by VARCHAR(64) DEFAULT NULL
);