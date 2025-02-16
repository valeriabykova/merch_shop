package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type userCredentials struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}

type authService struct {
	credRepo      CredentialsRepository
	encryptSecret []byte
}

func NewService(repo CredentialsRepository, secret []byte) Service {
	return &authService{
		credRepo:      repo,
		encryptSecret: secret,
	}
}

func (s *authService) LoginHandler(c *gin.Context) {
	var user userCredentials
	if err := c.BindJSON(&user); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		log.Println(fmt.Errorf("error binding credentials: %w", err))
		return
	}
	password, ok, err := s.credRepo.GetCredentials(c, user.Login)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Println(fmt.Errorf("error getting credentials: %w", err))
		return
	}
	if ok {
		if password != user.Password {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	} else {
		// Я знаю, что пароли надо хешировать, просто не успела
		err = s.credRepo.SetCredentials(c, user.Login, user.Password)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Println(fmt.Errorf("error creating user: %w", err))
			return
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"login": user.Login})
	signedToken, err := token.SignedString(s.encryptSecret)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Println(fmt.Errorf("error creating jwt token: %w", err))
		return
	}
	c.IndentedJSON(http.StatusOK, map[string]string{
		"token": signedToken,
	})
}

func (s *authService) AuthMiddleware(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (interface{}, error) {
		return s.encryptSecret, nil
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "invalid token")
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var loginClaim interface{}
		loginClaim, ok = claims["login"]
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, "invalid token")
			return
		}
		var login string
		login, ok = loginClaim.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, "invalid claims")
			return
		}
		c.Set("login", login)
		c.Next()
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, "invalid token")
		return
	}
}
