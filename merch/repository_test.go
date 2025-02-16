package merch_test

import (
	"context"
	"merch_avito/merch"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func initRepo() (merch.Repository, testcontainers.Container, error) {
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:15.3-alpine",
		postgres.WithInitScripts(filepath.Join(".", "testdata", "init-db.sql")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, nil, err
	}

	merchRepo, err := merch.NewMerchRepo(ctx, pool)
	if err != nil {
		return nil, nil, err
	}
	return merchRepo, container, nil
}

func stopContainer(container testcontainers.Container) {
	if err := container.Terminate(context.Background()); err != nil {
		panic(err)
	}
}

func Test_GetUserId(t *testing.T) {
	merchRepo, container, err := initRepo()
	require.NoError(t, err)
	ctx := context.Background()
	id, err := merchRepo.GetUserID(ctx, "test_user1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	_, err = merchRepo.GetUserID(ctx, "test_user2")
	require.Error(t, err)
	stopContainer(container)
}

func Test_GetBalance(t *testing.T) {
	merchRepo, container, err := initRepo()
	require.NoError(t, err)
	ctx := context.Background()
	balance, err := merchRepo.GetBalance(ctx, int64(1))
	require.NoError(t, err)
	assert.Equal(t, int64(1000), balance)
	stopContainer(container)
}

func Test_GetInventory(t *testing.T) {
	merchRepo, container, err := initRepo()
	require.NoError(t, err)
	ctx := context.Background()
	inventory, err := merchRepo.GetInventory(ctx, int64(1))
	require.NoError(t, err)
	assert.Equal(t, []merch.InventoryItem{
		{
			Name:     "t-shirt",
			Quantity: 2,
		},
	}, inventory)
	stopContainer(container)
}

func Test_SendCoin(t *testing.T) {
	merchRepo, container, err := initRepo()
	require.NoError(t, err)
	ctx := context.Background()
	// перевод самому себе
	err = merchRepo.SendCoin(ctx, 1, 1, 5)
	require.EqualError(t, err, "cannot send coins to yourself")
	// перевод больше, чем есть коинов
	err = merchRepo.SendCoin(ctx, 1, 2, 1001)
	require.ErrorContains(t, err, "insufficient funds:")

	balance, err := merchRepo.GetBalance(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), balance)

	balance, err = merchRepo.GetBalance(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), balance)

	// успешный перевод
	err = merchRepo.SendCoin(ctx, 1, 2, 100)
	require.NoError(t, err)
	// заодно проверим историю переводов
	sent, err := merchRepo.GetSentCoins(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, []merch.SentCoins{{
		To:     "test_user3",
		Amount: 100,
	}}, sent)
	balance, err = merchRepo.GetBalance(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(900), balance)

	balance, err = merchRepo.GetBalance(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(1100), balance)

	recieved, err := merchRepo.GetRecievedCoins(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, []merch.RecievedCoins{{
		From:   "test_user1",
		Amount: 100,
	}}, recieved)
	stopContainer(container)
}
