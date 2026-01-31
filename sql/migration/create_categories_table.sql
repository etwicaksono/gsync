CREATE TABLE `categories` (
CREATE TABLE categories (
    id BIGINT PRIMARY KEY,
    parent VARCHAR(64) NOT NULL,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    nature VARCHAR(64) NOT NULL,
    visibility VARCHAR(64) NOT NULL,
    position JSONB DEFAULT NULL,
    created_at TIMESTAMP NOT NULL,
    created_by VARCHAR(64) DEFAULT NULL,
    updated_at TIMESTAMP DEFAULT NULL,
    updated_by VARCHAR(64) DEFAULT NULL
);