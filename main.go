package main

import (
	"context"
	"fmt"
	"log"
	"merch_avito/auth"
	"merch_avito/merch"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	serverPort = "8080"
	dbHost     = "postgres"
	dbUser     = "user"
	dbPassword = "password"
	dbName     = "merch_db"
	dbPort     = "5432"
	jwtSecret  = []byte("secret")
)

func loadEnv() {
	if port := os.Getenv("PORT"); port != "" {
		serverPort = port
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		dbHost = host
	}
	if user := os.Getenv("DB_USER"); user != "" {
		dbUser = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		dbPassword = password
	}
	if db := os.Getenv("DB_NAME"); db != "" {
		dbName = db
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		dbPort = port
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		jwtSecret = []byte(secret)
	}
}

func main() {
	loadEnv()

	router := gin.Default()
	err := router.SetTrustedProxies(nil) // sorry
	if err != nil {
		log.Panic(err)
	}

	fulldDBHost := net.JoinHostPort(dbHost, dbPort)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s", dbUser, dbPassword, fulldDBHost, dbName)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()
	authRepo, err := auth.NewRepo(ctx, pool)
	if err != nil {
		log.Panic(err)
	}

	authService := auth.NewService(authRepo, jwtSecret)

	merchRepo, err := merch.NewMerchRepo(ctx, pool)
	if err != nil {
		log.Panic(err)
	}

	merchService := merch.NewService(merchRepo)

	router.POST("/api/auth", authService.LoginHandler)

	authed := router.Group("/api", authService.AuthMiddleware)
	authed.GET("/info", merchService.GetUserInfo)
	authed.POST("/sendCoin", merchService.SendCoin)
	authed.POST("/buy/:item", merchService.BuyItem)

	serverHost := net.JoinHostPort("0.0.0.0", serverPort)
	err = router.Run(serverHost)
	if err != nil {
		log.Panic(err)
	}
}
