package main

import (
	"context"
	"gw-currency-wallet/internal/auth"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/handlers"
	"gw-currency-wallet/internal/notifications"
	"gw-currency-wallet/internal/proto/proto/exchange"
	"gw-currency-wallet/internal/storages/db/postgres"
	"gw-currency-wallet/pkg/logging"
	"log"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	storage, closeDB := postgres.NewPostgresRepository(ctx, &cfg.Storage, logger)
	defer closeDB()

	exchangerConn, err := grpc.NewClient(cfg.ExchangerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to exchanger: %v", err)
	}
	defer exchangerConn.Close()

	exchangerClient := exchange.NewExchangeServiceClient(exchangerConn)

	authService := auth.NewService(storage, cfg.JWTSecret, exchangerClient, logger)

	notificationService := notifications.NewNotificationService(cfg.KafkaBroker, cfg.KafkaTopic)
	defer notificationService.Close()

	//3. Создание сервера
	router := gin.Default()

	// Публичные ручки
	router.POST("/api/v1/register", handlers.Register(authService))
	router.POST("/api/v1/login", handlers.Login(authService))
	router.GET("/api/v1/exchange/rates", auth.JWTMiddleware(authService), handlers.GetExchangeRates(authService))

	// Защищённые маршруты
	protected := router.Group("/api/v1")
	protected.Use(auth.JWTMiddleware(authService)) // middleware для JWT
	{
		protected.GET("/balance/:currency", handlers.GetBalance(storage))
		protected.POST("/exchange", handlers.Exchange(storage, authService, notificationService))
	}

	//4. Запуск сервера на заданном порту
	logger.Infof("Server started on port %s", cfg.HTTPPort)
	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}

	//5. Ожидание сигнала завершения
}
