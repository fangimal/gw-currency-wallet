CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS balances(
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency VARCHAR(3) NOT NULL,
    amount DECIMAL(15,2) NOT NULL CHECK ( amount >= 0 ),
    PRIMARY KEY (user_id, currency)
);

CREATE INDEX IF NOT EXISTS idx_balances_user_currency ON balances(user_id, currency);


CREATE DATABASE wallet_test_db;