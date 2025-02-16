package merch

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type service struct {
	repo Repository
}

type CoinHistory struct {
	Recieved []RecievedCoins `json:"recieved"`
	Sent     []SentCoins     `json:"sent"`
}

type InfoResponse struct {
	Coins       int64           `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) GetUserInfo(c *gin.Context) {
	login := c.GetString("login")
	userID, err := s.repo.GetUserID(c, login)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Println(fmt.Errorf("error getting user: %w", err))
		return
	}

	balance, err := s.repo.GetBalance(c, userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Println(fmt.Errorf("error getting balance: %w", err))
		return
	}

	inventory, err := s.repo.GetInventory(c, userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Println(fmt.Errorf("error getting inventory: %w", err))
		return
	}

	recieved, err := s.repo.GetRecievedCoins(c, userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Println(fmt.Errorf("error getting recieved coins: %w", err))
		return
	}

	sent, err := s.repo.GetSentCoins(c, userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Println(fmt.Errorf("error getting sent coins: %w", err))
		return
	}

	response := &InfoResponse{
		Coins:     balance,
		Inventory: inventory,
		CoinHistory: CoinHistory{
			Recieved: recieved,
			Sent:     sent,
		},
	}
	c.JSON(http.StatusOK, response)
}

func (s *service) SendCoin(c *gin.Context) {
	login := c.GetString("login")
	userID, err := s.repo.GetUserID(c, login)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	var req SentCoins
	err = c.Bind(req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	toID, err := s.repo.GetUserID(c, req.To)
	if errors.Is(err, ErrNotFound) {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	err = s.repo.SendCoin(c, userID, toID, req.Amount)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.Status(http.StatusOK)
}

func (s *service) BuyItem(c *gin.Context) {
	login := c.GetString("login")
	userID, err := s.repo.GetUserID(c, login)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	itemName := c.Param("item")
	err = s.repo.BuyItem(c, userID, itemName)
	if errors.Is(err, ErrNotFound) {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.Status(http.StatusOK)
}
