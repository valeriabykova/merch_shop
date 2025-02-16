CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    balance INTEGER NOT NULL DEFAULT 1000 CHECK (balance >= 0)

);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0)
);

CREATE TABLE purchases (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    item_id INTEGER REFERENCES items(id) ON DELETE CASCADE,
    quantity INTEGER CHECK (quantity > 0),
    PRIMARY KEY (user_id, item_id)
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    from_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    to_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_coin_transactions_from ON coin_transactions(from_id);

CREATE INDEX idx_coin_transactions_to ON coin_transactions(to_id);