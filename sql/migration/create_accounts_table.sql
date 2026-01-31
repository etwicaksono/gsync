CREATE TABLE `accounts` (
CREATE TABLE accounts (
    id BIGINT PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(64) NOT NULL,
    active VARCHAR(64) NOT NULL,
    usability VARCHAR(64) NOT NULL,
    ownership VARCHAR(64) NOT NULL,
    position JSONB DEFAULT NULL,
    created_at TIMESTAMP NOT NULL,
    created_by VARCHAR(64) DEFAULT NULL,
    updated_at TIMESTAMP DEFAULT NULL,
    updated_by VARCHAR(64) DEFAULT NULL
);