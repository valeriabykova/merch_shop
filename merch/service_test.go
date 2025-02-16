package merch_test

import (
	"bytes"
	"encoding/json"
	"merch_avito/merch"
	"merch_avito/merch/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetInfo_Existing(t *testing.T) {
	testBalance := int64(1000)
	testInventory := []merch.InventoryItem{{Name: "t-shirt", Quantity: 2}}
	testRecieved := []merch.RecievedCoins{{From: "test2", Amount: 100}}
	testSent := []merch.SentCoins{{To: "test2", Amount: 100}}
	testCoinHistory := merch.CoinHistory{Recieved: testRecieved, Sent: testSent}
	testResponse := merch.InfoResponse{
		Coins:       testBalance,
		Inventory:   testInventory,
		CoinHistory: testCoinHistory,
	}

	merchRepo := mocks.NewRepository(t)
	merchRepo.On("GetUserID", mock.Anything, "test").Return(int64(1), nil)
	merchRepo.On("GetBalance", mock.Anything, int64(1)).Return(int64(1000), nil)
	merchRepo.On("GetInventory", mock.Anything, int64(1)).Return(testInventory, nil)
	merchRepo.On("GetRecievedCoins", mock.Anything, int64(1)).Return(testRecieved, nil)
	merchRepo.On("GetSentCoins", mock.Anything, int64(1)).Return(testSent, nil)

	gin.SetMode(gin.TestMode)
	service := merch.NewService(merchRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/info", bytes.NewBufferString(""))
	c.Set("login", "test")
	service.GetUserInfo(c)

	var response merch.InfoResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, testResponse, response)
}

func TestGetInfo_NonExisting(t *testing.T) {
	merchRepo := mocks.NewRepository(t)
	merchRepo.On("GetUserID", mock.Anything, "test").Return(int64(0), merch.ErrNotFound)

	gin.SetMode(gin.TestMode)
	service := merch.NewService(merchRepo)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/info", bytes.NewBufferString(""))
	c.Set("login", "test")
	service.GetUserInfo(c)

	var response merch.InfoResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	// Пользователь не может просто не существовать, так как он прошел auth
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
