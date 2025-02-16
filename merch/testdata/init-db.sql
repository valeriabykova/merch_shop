CREATE TABLE
    IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        login VARCHAR(50) UNIQUE NOT NULL,
        password_hash VARCHAR(100) NOT NULL,
        balance INTEGER DEFAULT 1000 NOT NULL CHECK (balance >= 0)
    );

INSERT INTO
    users (login, password_hash)
VALUES
    ('test_user1', 'test_password1'),
    ('test_user3', 'test3');

CREATE TABLE
    IF NOT EXISTS items (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50) UNIQUE NOT NULL,
        price INTEGER NOT NULL CHECK (price > 0)
    );

INSERT INTO
    items (name, price)
VALUES
        ('t-shirt', 80),
        ('cup', 20),
        ('book', 50),
        ('pen', 10),
        ('powerbank', 200),
        ('hoody', 300),
        ('umbrella', 200),
        ('socks', 10),
        ('wallet', 50),
        ('pink-hoody', 500);

CREATE TABLE
IF NOT EXISTS
    purchases (
        user_id INTEGER REFERENCES users (id) ON DELETE CASCADE,
        item_name VARCHAR(50) NOT NULL,
        quantity INTEGER DEFAULT 0 CHECK (quantity >= 0),
        PRIMARY KEY (user_id, item_name)
    );

INSERT INTO
    purchases (user_id, item_name, quantity)
VALUES
        (1, 't-shirt', 2);

CREATE TABLE IF NOT EXISTS transactions  (
    id SERIAL PRIMARY KEY,
    from_id INTEGER REFERENCES users(id),
    to_id INTEGER NOT NULL REFERENCES users(id),
    amount INTEGER NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP DEFAULT NOW()
);