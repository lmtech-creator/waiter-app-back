package main

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"os"
	"strings"

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
//
// @securityDefinitions.apikey  SessionToken
// @in                          header
// @name                        Authorization
// @description                 JWT de sesión de cliente (QR). Formato: "Bearer <token>"
//
// @securityDefinitions.apikey  AdminToken
// @in                          header
// @name                        Authorization
// @description                 JWT de administrador. Formato: "Bearer <token>"
func main() {
	_ = godotenv.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync() //nolint:errcheck
	zap.ReplaceGlobals(logger)

	cfg := config.Load()

	secret := resolveSecret("SESSION_SECRET", cfg.SessionSecret)
	adminSecret := resolveSecret("ADMIN_SECRET", cfg.AdminSecret)

	db, err := persistence.NewDatabase(cfg.DSN)
	if err != nil {
		zap.L().Fatal("failed to initialize database", zap.Error(err))
	}

	// Repositories
	restaurantRepo := persistence.NewRestaurantRepo(db)
	tableRepo := persistence.NewTableRepo(db)
	requestRepo := persistence.NewRequestRepo(db)
	feedbackRepo := persistence.NewFeedbackRepo(db)
	adminRepo := persistence.NewAdminRepo(db)

	// WebSocket hub
	hub := ws.NewHub(parseOrigins(cfg.AllowedOrigins))

	// Use cases
	requestUC := usecase.NewRequestUseCase(requestRepo, tableRepo, hub)
	feedbackUC := usecase.NewFeedbackUseCase(feedbackRepo, tableRepo)
	restaurantUC := usecase.NewRestaurantUseCase(restaurantRepo, tableRepo)
	adminUC := usecase.NewAdminUseCase(adminRepo, adminSecret)
	statsUC := usecase.NewStatsUseCase(requestRepo, feedbackRepo)

	// Seed superadmin on first run.
	if seeded, username, password, err := usecase.SeedAdminIfNeeded(adminRepo); err != nil {
		zap.L().Error("failed to seed admin user", zap.Error(err))
	} else if seeded {
		zap.L().Info("superadmin created — change this password immediately",
			zap.String("username", username),
			zap.String("password", password),
		)
	}

	// Handlers
	requestHandler := handler.NewRequestHandler(requestUC)
	feedbackHandler := handler.NewFeedbackHandler(feedbackUC)
	restaurantHandler := handler.NewRestaurantHandler(restaurantUC)
	sessionHandler := handler.NewSessionHandler(tableRepo, secret)
	adminHandler := handler.NewAdminHandler(adminUC)
	statsHandler := handler.NewStatsHandler(statsUC)

	// Background tasks
	go mw.StartLimiterCleanup()

	// Router
	router := handler.SetupRouter(requestHandler, feedbackHandler, restaurantHandler, sessionHandler, adminHandler, statsHandler, hub, secret, adminSecret)

	zap.L().Info("server starting", zap.String("port", cfg.Port))
	if err := router.Run(":" + cfg.Port); err != nil {
		zap.L().Fatal("server failed", zap.Error(err))
	}
}

// parseOrigins splits a comma-separated origin string into a slice, trimming whitespace.
func parseOrigins(s string) []string {
	var out []string
	for _, o := range strings.Split(s, ",") {
		if o = strings.TrimSpace(o); o != "" {
			out = append(out, o)
		}
	}
	return out
}

// resolveSecret decodes a hex secret from env, or generates a random one for development.
func resolveSecret(envKey, hexStr string) []byte {
	if hexStr != "" {
		b, err := hex.DecodeString(hexStr)
		if err == nil && len(b) >= 32 {
			return b
		}
		zap.L().Warn(envKey + " is invalid or too short; using random secret (not for production)")
	} else {
		zap.L().Warn(envKey + " not set; using random secret (not for production)")
	}
	b := make([]byte, 32)
	if _, err := cryptorand.Read(b); err != nil {
		zap.L().Fatal("failed to generate random secret", zap.Error(err))
	}
	return b
}
