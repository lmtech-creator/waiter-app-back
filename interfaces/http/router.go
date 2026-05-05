package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/waiter/back/domain/entity"
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

	// All admin-authenticated routes share this base middleware.
	adminBase := api.Group("")
	adminBase.Use(mw.AdminAuthMiddleware(adminSecret))
	{
		// ── Superadmin only ──────────────────────────────────────────────────
		superadmin := adminBase.Group("")
		superadmin.Use(mw.RequireRole(entity.RoleSuperAdmin))
		{
			superadmin.GET("/restaurants", restaurantHandler.GetAll)
			superadmin.POST("/restaurants", restaurantHandler.Create)
		}

		// ── User management: superadmin can do everything; owner can manage
		// employees of their own restaurant. The handler enforces the distinction.
		userMgmt := adminBase.Group("")
		userMgmt.Use(mw.RequireRole(entity.RoleSuperAdmin, entity.RoleOwner))
		{
			userMgmt.POST("/admin/users", adminHandler.CreateUser)
			userMgmt.GET("/admin/users", adminHandler.ListUsers)
			userMgmt.DELETE("/admin/users/:id", adminHandler.DeleteUser)
		}

		// ── Owner + Superadmin (with restaurant scope) ───────────────────────
		ownerOrAbove := adminBase.Group("")
		ownerOrAbove.Use(mw.RequireRole(entity.RoleOwner, entity.RoleSuperAdmin))
		ownerOrAbove.Use(mw.RequireRestaurantScope())
		{
			ownerOrAbove.GET("/restaurants/:restaurantId", restaurantHandler.Get)
			ownerOrAbove.POST("/restaurants/:restaurantId/tables", restaurantHandler.CreateTable)
			ownerOrAbove.GET("/restaurants/:restaurantId/tables", restaurantHandler.GetTables)
			ownerOrAbove.POST("/admin/tables/:id/regenerate-qr", restaurantHandler.RegenerateQR)
		}

		// ── Employee + Owner + Superadmin (operational routes) ───────────────
		operational := adminBase.Group("")
		operational.Use(mw.RequireRole(entity.RoleEmployee, entity.RoleOwner, entity.RoleSuperAdmin))
		operational.Use(mw.RequireRestaurantScope())
		{
			operational.GET("/restaurants/:restaurantId/requests/active", requestHandler.GetActive)
		}

		// Complete request: available to all admin roles
		adminBase.PATCH("/requests/:requestId", requestHandler.Complete)
	}

	// WebSocket
	wsHandler := NewWSHandler(hub, secret, adminSecret)
	api.GET("/ws/:restaurantId", wsHandler.Connect)

	return r
}
