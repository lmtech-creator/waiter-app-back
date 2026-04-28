package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ws "github.com/waiter/back/infrastructure/websocket"
	mw "github.com/waiter/back/interfaces/http/middleware"
)

func SetupRouter(
	requestHandler *RequestHandler,
	feedbackHandler *FeedbackHandler,
	restaurantHandler *RestaurantHandler,
	sessionHandler *SessionHandler,
	adminHandler *AdminHandler,
	hub *ws.Hub,
	secret []byte,
	adminSecret []byte,
) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	r.Use(mw.RateLimitMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	// Public: session creation via QR code
	api.POST("/session", sessionHandler.Create)

	// Customer endpoints — require valid session JWT
	customer := api.Group("")
	customer.Use(mw.AuthMiddleware(secret))
	{
		customer.POST("/requests", requestHandler.Create)
	}

	// Public admin auth
	api.POST("/admin/login", adminHandler.Login)

	// Public: customer-facing read routes
	api.GET("/tables/:tableId/status", requestHandler.GetTableStatus)
	api.POST("/feedback", feedbackHandler.Create)

	// Admin protected routes
	adminGroup := api.Group("")
	adminGroup.Use(mw.AdminAuthMiddleware(adminSecret))
	{
		adminGroup.GET("/restaurants/:restaurantId/requests/active", requestHandler.GetActive)
		adminGroup.PATCH("/requests/:requestId", requestHandler.Complete)
		adminGroup.POST("/restaurants", restaurantHandler.Create)
		adminGroup.GET("/restaurants/:restaurantId", restaurantHandler.Get)
		adminGroup.POST("/restaurants/:restaurantId/tables", restaurantHandler.CreateTable)
		adminGroup.GET("/restaurants/:restaurantId/tables", restaurantHandler.GetTables)
		adminGroup.POST("/admin/tables/:id/regenerate-qr", restaurantHandler.RegenerateQR)
	}

	// WebSocket
	wsHandler := NewWSHandler(hub, secret, adminSecret)
	api.GET("/ws/:restaurantId", wsHandler.Connect)

	return r
}
