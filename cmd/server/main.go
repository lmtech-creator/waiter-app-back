package main

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"os"

	"github.com/joho/godotenv"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/config"
	_ "github.com/waiter/back/docs"
	"github.com/waiter/back/infrastructure/persistence"
	ws "github.com/waiter/back/infrastructure/websocket"
	handler "github.com/waiter/back/interfaces/http"
	mw "github.com/waiter/back/interfaces/http/middleware"
	"go.uber.org/zap"
)

// @title          Waiter API
// @version        1.0
// @description    Sistema de llamado de mozos para restaurantes
// @host           localhost:8080
// @BasePath       /api/v1
func main() {
	_ = godotenv.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync() //nolint:errcheck
	zap.ReplaceGlobals(logger)

	cfg := config.Load()

	secret := resolveSecret(cfg.SessionSecret)

	db, err := persistence.NewDatabase(cfg.DSN)
	if err != nil {
		zap.L().Fatal("failed to initialize database", zap.Error(err))
	}

	// Repositories
	restaurantRepo := persistence.NewRestaurantRepo(db)
	tableRepo := persistence.NewTableRepo(db)
	requestRepo := persistence.NewRequestRepo(db)
	feedbackRepo := persistence.NewFeedbackRepo(db)

	// WebSocket hub
	hub := ws.NewHub()

	// Use cases
	requestUC := usecase.NewRequestUseCase(requestRepo, tableRepo, hub)
	feedbackUC := usecase.NewFeedbackUseCase(feedbackRepo, tableRepo)
	restaurantUC := usecase.NewRestaurantUseCase(restaurantRepo, tableRepo)

	// Handlers
	requestHandler := handler.NewRequestHandler(requestUC)
	feedbackHandler := handler.NewFeedbackHandler(feedbackUC)
	restaurantHandler := handler.NewRestaurantHandler(restaurantUC)
	sessionHandler := handler.NewSessionHandler(tableRepo, secret)

	// Background tasks
	go mw.StartLimiterCleanup()

	// Router
	router := handler.SetupRouter(requestHandler, feedbackHandler, restaurantHandler, sessionHandler, hub, secret)

	zap.L().Info("server starting", zap.String("port", cfg.Port))
	if err := router.Run(":" + cfg.Port); err != nil {
		zap.L().Fatal("server failed", zap.Error(err))
	}
}

// resolveSecret decodes SESSION_SECRET from hex, or generates a random one for development.
func resolveSecret(hexStr string) []byte {
	if hexStr != "" {
		b, err := hex.DecodeString(hexStr)
		if err == nil && len(b) >= 32 {
			return b
		}
		zap.L().Warn("SESSION_SECRET is invalid or too short; using random secret (not for production)")
	} else {
		zap.L().Warn("SESSION_SECRET not set; using random secret (not for production)")
	}
	b := make([]byte, 32)
	if _, err := cryptorand.Read(b); err != nil {
		zap.L().Fatal("failed to generate random secret", zap.Error(err))
	}
	return b
}
