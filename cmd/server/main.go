package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/serinew/core/internal/app"
	"github.com/serinew/core/internal/config"
	"github.com/serinew/core/internal/middleware"
	"github.com/serinew/core/internal/repository"
	"github.com/serinew/core/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("config: .env not loaded (%v), using process env only", err)
	}

	cfg := config.Load()
	if len(cfg.TokenSecret) < service.MinTokenSecretLength {
		log.Printf("config: TOKEN_SECRET is empty or shorter than %d chars — POST /sign/login issuance will fail", service.MinTokenSecretLength)
	}
	ctx := context.Background()

	store, shutdownRepo, err := repository.Connect(ctx, cfg.PostgresDSN, cfg.RedisURL)
	if err != nil {
		log.Fatalf("repository: %v", err)
	}
	defer shutdownRepo()
	if store.Redis != nil {
		log.Println("repository: redis client ready")
	}

	if getenvProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	middleware.RegisterRecovery(r)
	tokens := service.NewTokenService(store, cfg.TokenSecret)
	r.Use(middleware.RouteBurstLimiter(cfg, store.Redis, tokens))
	app.AppRoutes(r, store, cfg)
	middleware.RegisterNotFound(r)

	addr := ":" + cfg.HTTPPort
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func getenvProd() bool {
	return os.Getenv("NODE_ENV") == "production" || os.Getenv("GIN_MODE") == "release"
}
