package auth_test

import (
	"bytes"
	"encoding/json"
	"merch_avito/auth"
	"merch_avito/auth/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthService(t *testing.T) {
	authRepo := mocks.NewCredentialsRepo(t)
	authRepo.On("GetCredentials", mock.Anything, "existing").Return("password", true, nil)
	authRepo.On("GetCredentials", mock.Anything, "nonexisting").Return("", false, nil)
	authRepo.On("SetCredentials", mock.Anything, "nonexisting", "password").Return(nil)

	gin.SetMode(gin.TestMode)
	service := auth.NewService(authRepo, []byte("secret"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(
		`{"username":"existing", "password": "wrong"}`,
	))

	service.LoginHandler(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(
		`{"username":"nonexisting", "password": "password"}`,
	))

	service.LoginHandler(c)
	assert.Equal(t, 200, w.Code)

	tokenJSON := w.Body.Bytes()
	var data map[string]interface{}
	err := json.Unmarshal(tokenJSON, &data)
	require.NoError(t, err)

	token := data["token"].(string)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(""))
	c.Request.Header.Add("Authorization", "hrpp")
	service.AuthMiddleware(c)

	assert.Equal(t, "\"invalid token\"", w.Body.String())

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(""))
	c.Request.Header.Add("Authorization", token)
	service.AuthMiddleware(c)

	assert.Equal(t, "nonexisting", c.GetString("login"))
}
