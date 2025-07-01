CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL REFERENCES accounts(id),
    type VARCHAR(20) NOT NULL CHECK (type IN ('deposit', 'withdrawal', 'transfer')),
    amount DECIMAL(15,2) NOT NULL,
    time TIMESTAMP NOT NULL DEFAULT NOW(),
    fee DECIMAL(15,2) NULL
);

CREATE TABLE transaction_trasfers (
    id SERIAL PRIMARY KEY,
    trans_id INT NOT NULL REFERENCES transactions(id),
    dest_account_id INT NOT NULL REFERENCES accounts(id)
)