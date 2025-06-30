CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    account_number VARCHAR(20) NOT NULL UNIQUE,
    account_type VARCHAR(10) NOT NULL CHECK (account_type IN ('debit', 'credit')),
    balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    opening_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status VARCHAR(10) NOT NULL CHECK (status IN ('active', 'blocked', 'closed'))
);