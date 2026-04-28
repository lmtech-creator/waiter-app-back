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
	hub *ws.Hub,
	secret []byte,
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

	// Restaurant admin endpoints (TODO: add AdminAuthMiddleware when implemented)
	admin := api.Group("")
	{
		admin.GET("/tables/:tableId/status", requestHandler.GetTableStatus)
		admin.POST("/feedback", feedbackHandler.Create)
		admin.GET("/restaurants/:restaurantId/requests/active", requestHandler.GetActive)
		admin.PATCH("/requests/:requestId", requestHandler.Complete)
		admin.POST("/restaurants", restaurantHandler.Create)
		admin.GET("/restaurants/:restaurantId", restaurantHandler.Get)
		admin.POST("/restaurants/:restaurantId/tables", restaurantHandler.CreateTable)
		admin.GET("/restaurants/:restaurantId/tables", restaurantHandler.GetTables)
		admin.POST("/admin/tables/:id/regenerate-qr", restaurantHandler.RegenerateQR)
	}

	// WebSocket
	wsHandler := NewWSHandler(hub)
	api.GET("/ws/:restaurantId", wsHandler.Connect)

	return r
}
