package main

import (
	"context"
	"gw-currency-wallet/internal/auth"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/handlers"
	"gw-currency-wallet/internal/storages/db/postgres"
	"gw-currency-wallet/pkg/logging"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	//0. Подготовка логгера
	logger := logging.GetLogger()
	logger.Info("Start...")

	//1. Загрузка конфига
	cfg := config.GetConfig()
	logger.Info("Config loaded")

	//2. Подключение к бд
	//repo, closeDB := repository.NewRepository(ctx, &cfg.Storage, logger)
	repo, closeDB := postgres.NewPostgresRepository(ctx, &cfg.Storage, logger)
	defer closeDB()

	authService := auth.NewService(repo, cfg.JWTSecret)

	//3. Создание сервера
	router := gin.Default()

	// Публичные ручки
	router.POST("/api/v1/register", handlers.Register(authService))
	router.POST("/api/v1/login", handlers.Login(authService))

	// Защищённые ручки
	protected := router.Group("/api/v1")
	protected.Use(auth.JWTMiddleware(authService)) // middleware для JWT
	{
		protected.GET("/balance/:currency", handlers.GetBalance(repo))
		//protected.POST("/exchange", handlers.Exchange(repo, authService.ExchangerClient()))
	}

	//4. Запуск сервера на заданном порту
	logger.Infof("Server started on port %s", cfg.HTTPPort)
	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}

	//5. Ожидание сигнала завершения
}
