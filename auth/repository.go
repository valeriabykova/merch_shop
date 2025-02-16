package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type credentialsRepository struct {
	pool *pgxpool.Pool
}

var _ CredentialsRepository = (*credentialsRepository)(nil)

func NewRepo(ctx context.Context, pool *pgxpool.Pool) (CredentialsRepository, error) {
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		login VARCHAR(50) UNIQUE NOT NULL,
		password_hash VARCHAR(100) NOT NULL,
		balance INTEGER NOT NULL DEFAULT 1000 CHECK (balance >= 0)
	);`)
	if err != nil {
		return nil, fmt.Errorf("error creating table users: %w", err)
	}

	return &credentialsRepository{
		pool: pool,
	}, nil
}

func (cr *credentialsRepository) GetCredentials(ctx context.Context, login string) (string, bool, error) {
	row := cr.pool.QueryRow(ctx, "SELECT password_hash FROM users WHERE login = $1", login)
	var pass string
	err := row.Scan(&pass)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return pass, true, err
}

func (cr *credentialsRepository) SetCredentials(ctx context.Context, login, password string) error {
	_, err := cr.pool.Exec(ctx, "INSERT INTO users (login, password_hash) VALUES ($1, $2)", login, password)
	return err
}
