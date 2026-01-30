package main

import (
	"context"
	_ "gw-currency-wallet/docs" //для запуска swagger
	"gw-currency-wallet/internal/auth"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/handlers"
	"gw-currency-wallet/internal/notifications"
	"gw-currency-wallet/internal/proto/proto/exchange"
	"gw-currency-wallet/internal/storages/db/postgres"
	"gw-currency-wallet/pkg/logging"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title Currency Wallet API
// @version 1.0
// @description API for currency wallet operations
// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
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

	// Настройка маршрутов
	handlers.SetupRoutes(router, storage, authService, notificationService)

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	//4. Запуск сервера на заданном порту
	logger.Infof("Server started on port %s", cfg.HTTPPort)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	// Запуск в горутине
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	//5. Shutdown. Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited gracefully")
}
