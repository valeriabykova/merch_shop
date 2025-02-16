package merch

import (
	"context"

	"github.com/gin-gonic/gin"
)

type InventoryItem struct {
	Name     string `json:"type"`
	Quantity int64  `json:"quantity"`
}

type RecievedCoins struct {
	From   string `json:"fromUser"`
	Amount int64  `json:"amount"`
}

type SentCoins struct {
	To     string `json:"toUser"`
	Amount int64  `json:"amount"`
}

type Service interface {
	GetUserInfo(*gin.Context)
	SendCoin(*gin.Context)
	BuyItem(*gin.Context)
}

type Repository interface {
	GetUserID(ctx context.Context, username string) (int64, error)
	GetBalance(ctx context.Context, userID int64) (int64, error)
	GetInventory(ctx context.Context, userID int64) ([]InventoryItem, error)
	GetRecievedCoins(ctx context.Context, userID int64) ([]RecievedCoins, error)
	GetSentCoins(ctx context.Context, userID int64) ([]SentCoins, error)
	SendCoin(ctx context.Context, fromUser, toUser, amount int64) error
	BuyItem(ctx context.Context, userID int64, itemName string) error
}
