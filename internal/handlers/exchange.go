package handlers

import (
	"context"
	"gw-currency-wallet/internal/auth"
	"gw-currency-wallet/internal/notifications"
	"gw-currency-wallet/internal/storages"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ExchangeRequest struct {
	FromCurrency string  `json:"from_currency" binding:"required"`
	ToCurrency   string  `json:"to_currency" binding:"required"`
	Amount       float32 `json:"amount" binding:"required,gt=0"`
}

// @Summary Exchange currencies
// @Tags exchange
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body ExchangeRequest true "Exchange request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/exchange [post]
func Exchange(storage storages.Repository, authService *auth.Service, notificationService *notifications.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := auth.GetUserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		var req ExchangeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. Проверяем баланс
		balance, err := storage.GetBalance(c.Request.Context(), userID, req.FromCurrency)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "balance not found"})
			return
		}
		if balance < req.Amount {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "insufficient funds"})
			return
		}

		// 2. Получаем курс
		rate, err := authService.GetExchangeRateWithCache(req.FromCurrency, req.ToCurrency)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to get exchange rate"})
			return
		}

		receivedAmount := req.Amount * rate

		// 3. Атомарное обновление балансов
		err = storage.UpdateBalance(c.Request.Context(), userID, req.FromCurrency, -req.Amount)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to deduct balance"})
			return
		}

		err = storage.UpdateBalance(c.Request.Context(), userID, req.ToCurrency, receivedAmount)
		if err != nil {
			// Откат в реальном проекте требует транзакций!
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to add balance"})
			return
		}

		// 4. Проверка на крупный перевод (≥30 000)
		if req.Amount >= 30000 {
			//отправить в Kafka
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err = notificationService.SendLargeTransfer(ctx, userID, req.Amount, req.FromCurrency)
				if err != nil {
					return
				}
			}()
		}

		c.JSON(http.StatusOK, gin.H{
			"from_currency":   req.FromCurrency,
			"to_currency":     req.ToCurrency,
			"sent_amount":     req.Amount,
			"received_amount": receivedAmount,
			"rate":            rate,
		})
	}
}

// @Summary Get current exchange rates
// @Tags exchange
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/exchange/rates [get]
func GetExchangeRates(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		rates, err := authService.FetchAndCacheAllRates()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rates"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"rates": rates})
	}
}
