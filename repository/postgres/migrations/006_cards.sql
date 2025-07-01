CREATE TABLE cards (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL REFERENCES accounts(id),
    number BYTEA NOT NULL,
    expiry BYTEA NOT NULL,
    cvv BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
);