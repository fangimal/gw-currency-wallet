package handlers

import (
	"gw-currency-wallet/internal/auth"
	"gw-currency-wallet/internal/notifications"
	"gw-currency-wallet/internal/storages"

	"github.com/gin-gonic/gin"
)

// SetupRoutes настраивает все маршруты приложения
func SetupRoutes(router *gin.Engine, storage storages.Repository, authService *auth.Service, notificationService *notifications.NotificationService) {
	// Публичные маршруты
	router.POST("/api/v1/register", Register(authService))
	router.POST("/api/v1/login", Login(authService))

	// Swagger
	// Роуты Swagger остаются в main.go, так как они специфичны для запуска сервера

	// Защищённые маршруты
	protected := router.Group("/api/v1")
	protected.Use(auth.JWTMiddleware(authService)) // middleware для JWT
	{
		protected.GET("/balance/:currency", GetBalance(storage))
		protected.GET("/balance", GetTotalBalance(storage))
		protected.POST("/exchange", Exchange(storage, authService, notificationService))
		protected.GET("/exchange/rates", GetExchangeRates(authService))
		protected.POST("/wallet/deposit", Deposit(storage))
		protected.POST("/wallet/withdraw", Withdraw(storage))
	}
}
