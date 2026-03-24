package main

import (
	"context"
	"log"
	"time"

	"stratyx/backend/internal/config"
	httpserver "stratyx/backend/internal/platform/http"
	"stratyx/backend/internal/platform/security"
	"stratyx/backend/internal/repository"
	"stratyx/backend/internal/service"
)

func main() {
	cfg := config.Load()

	client, db, err := repository.NewMongo(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		log.Fatalf("mongo connect failed: %v", err)
	}
	defer func() {
		_ = client.Disconnect(nil)
	}()

	jwtManager := security.NewJWTManager(cfg.JWTSecret, cfg.AccessTokenTTLMinutes, cfg.RefreshTokenTTLDays)
	repos := repository.NewRepositories(db)
	indexCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := repos.EnsureIndexes(indexCtx); err != nil {
		log.Fatalf("failed to ensure indexes: %v", err)
	}
	services := service.NewServices(repos, jwtManager)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go services.StartEmailWorker(workerCtx)

	router := httpserver.NewRouter(cfg, services)
	log.Printf("STRATYX API listening on %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
