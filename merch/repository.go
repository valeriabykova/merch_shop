package merch

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type merchRepository struct {
	pool *pgxpool.Pool
}

func NewMerchRepo(ctx context.Context, pool *pgxpool.Pool) (Repository, error) {
	_, err := pool.Exec(ctx, `CREATE TABLE
    IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        login VARCHAR(50) UNIQUE NOT NULL,
        password_hash VARCHAR(100) NOT NULL,
        balance INTEGER DEFAULT 1000 NOT NULL CHECK (balance >= 0)
    );`)
	if err != nil {
		return nil, fmt.Errorf("error creating table users: %w", err)
	}

	_, err = pool.Exec(ctx, `CREATE TABLE
    IF NOT EXISTS items (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50) UNIQUE NOT NULL,
        price INTEGER NOT NULL CHECK (price > 0)
    );`)
	if err != nil {
		return nil, fmt.Errorf("error creating table items: %w", err)
	}

	_, err = pool.Exec(ctx, `CREATE TABLE
	IF NOT EXISTS
    purchases (
        user_id INTEGER REFERENCES users (id) ON DELETE CASCADE,
        item_name VARCHAR(50) NOT NULL,
        quantity INTEGER DEFAULT 0 CHECK (quantity >= 0),
        PRIMARY KEY (user_id, item_name)
    );`)
	if err != nil {
		return nil, fmt.Errorf("error creating table purchases: %w", err)
	}

	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS 
	transactions (
    id SERIAL PRIMARY KEY,
    from_id INTEGER REFERENCES users(id),
    to_id INTEGER NOT NULL REFERENCES users(id),
    amount INTEGER NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP DEFAULT NOW()
	);`)
	if err != nil {
		return nil, fmt.Errorf("error creating table transactions: %w", err)
	}

	return &merchRepository{pool}, nil
}

var (
	ErrNotFound = errors.New("not found")
)

func (mr *merchRepository) GetUserID(ctx context.Context, username string) (int64, error) {
	row := mr.pool.QueryRow(ctx, "SELECT id FROM users WHERE login = $1", username)
	var id int64
	err := row.Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	return id, err
}

func (mr *merchRepository) GetBalance(ctx context.Context, userID int64) (int64, error) {
	row := mr.pool.QueryRow(ctx, "SELECT balance FROM users WHERE id = $1", userID)
	var amount int64
	err := row.Scan(&amount)
	return amount, err
}

func (mr *merchRepository) GetInventory(ctx context.Context, userID int64) ([]InventoryItem, error) {
	rows, err := mr.pool.Query(ctx, "SELECT item_name, quantity from purchases WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	var items []InventoryItem
	for rows.Next() {
		var item InventoryItem
		err = rows.Scan(&item.Name, &item.Quantity)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (mr *merchRepository) GetRecievedCoins(ctx context.Context, userID int64) ([]RecievedCoins, error) {
	rows, err := mr.pool.Query(
		ctx,
		"SELECT login, amount from transactions JOIN users on from_id = users.id WHERE to_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	var recieved []RecievedCoins
	for rows.Next() {
		var rc RecievedCoins
		err = rows.Scan(&rc.From, &rc.Amount)
		if err != nil {
			return nil, err
		}
		recieved = append(recieved, rc)
	}
	return recieved, nil
}

func (mr *merchRepository) GetSentCoins(ctx context.Context, userID int64) ([]SentCoins, error) {
	rows, err := mr.pool.Query(
		ctx,
		"SELECT login, amount from transactions JOIN users ON to_id = users.id WHERE from_id = $1 ",
		userID,
	)
	if err != nil {
		return nil, err
	}
	var sent []SentCoins
	for rows.Next() {
		var sc SentCoins
		err = rows.Scan(&sc.To, &sc.Amount)
		if err != nil {
			return nil, err
		}
		sent = append(sent, sc)
	}
	return sent, nil
}

func (mr *merchRepository) SendCoin(ctx context.Context, fromID, toID, amount int64) error {
	if fromID == toID {
		return errors.New("cannot send coins to yourself")
	}
	tx, err := mr.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		err = tx.Rollback(ctx)
	}()

	var currentBalance int64
	err = tx.QueryRow(ctx,
		"SELECT balance FROM users WHERE id = $1 FOR UPDATE",
		fromID,
	).Scan(&currentBalance)

	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	if currentBalance < amount {
		return fmt.Errorf("insufficient funds: %d available, %d requested",
			currentBalance, amount)
	}

	_, err = tx.Exec(ctx,
		"UPDATE users SET balance = balance - $1 WHERE id = $2",
		amount, fromID,
	)
	if err != nil {
		return fmt.Errorf("failed to subtract balance: %w", err)
	}

	_, err = tx.Exec(ctx,
		"UPDATE users SET balance = balance + $1 WHERE id = $2",
		amount, toID,
	)
	if err != nil {
		return fmt.Errorf("failed to add balance: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO transactions 
			(from_id, to_id, amount) 
			VALUES ($1, $2, $3)`,
		fromID, toID, amount,
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}
	return nil
}

func (mr *merchRepository) BuyItem(ctx context.Context, userID int64, itemName string) (err error) {
	tx, err := mr.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		err = tx.Rollback(ctx)
	}()

	var currentBalance int64
	err = tx.QueryRow(ctx,
		"SELECT balance FROM users WHERE id = $1 FOR UPDATE",
		userID,
	).Scan(&currentBalance)

	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	var price int64
	err = tx.QueryRow(ctx,
		"SELECT price FROM items WHERE name = $1",
		itemName,
	).Scan(&price)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get the price of %s: %w", itemName, err)
	}

	if currentBalance < price {
		return fmt.Errorf("insufficient funds: %d available, %d required",
			currentBalance, price)
	}
	_, err = tx.Exec(ctx,
		"UPDATE users SET balance = balance - $1 WHERE id = $2",
		price, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to subtract balance: %w", err)
	}
	var count int64
	err = tx.QueryRow(ctx,
		"SELECT COUNT(*) FROM purchases WHERE item_name = $1 AND user_id = $2",
		itemName, userID,
	).Scan(&count)

	if err != nil {
		return err
	}

	if count == 0 {
		_, err = tx.Exec(ctx,
			`INSERT INTO purchases 
				(user_id, item_name, quantity) 
				VALUES ($1, $2, $3)`,
			userID, itemName, 1,
		)
		if err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}
	} else {
		_, err = tx.Exec(ctx,
			"UPDATE purchases SET quantity = quantity + 1 WHERE user_id = $1 AND item_name = $2",
			userID, itemName,
		)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}
	return nil
}
