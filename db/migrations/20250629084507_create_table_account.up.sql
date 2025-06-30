CREATE TABLE IF NOT EXISTS accounts (
    id bigint PRIMARY KEY,
    balance decimal(10, 5) NOT NULL DEFAULT 0,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);