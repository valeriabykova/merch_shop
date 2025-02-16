package auth

import (
	"context"

	"github.com/gin-gonic/gin"
)

type Service interface {
	LoginHandler(*gin.Context)
	AuthMiddleware(*gin.Context)
}

type CredentialsRepository interface {
	GetCredentials(ctx context.Context, login string) (string, bool, error)
	SetCredentials(ctx context.Context, login, password string) error
}
