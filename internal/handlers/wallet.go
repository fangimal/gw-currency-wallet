package handlers

import (
	"gw-currency-wallet/internal/auth"
	"gw-currency-wallet/internal/storages"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WalletOperation struct {
	Amount   float32 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency" binding:"required,oneof=USD RUB EUR"`
}

// @Summary Deposit funds to wallet
// @Tags wallet
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body WalletOperation true "Deposit request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /wallet/deposit [post]
func Deposit(storage storages.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := auth.GetUserID(c)

		var req WalletOperation
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid amount or currency"})
			return
		}

		err := storage.UpdateBalance(c.Request.Context(), userID, req.Currency, req.Amount)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
			return
		}

		balances, _ := storage.GetAllBalances(c.Request.Context(), userID)
		c.JSON(http.StatusOK, gin.H{
			"message":     "Account topped up successfully",
			"new_balance": balances,
		})
	}
}

// @Summary Withdraw funds from wallet
// @Tags wallet
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body WalletOperation true "Withdraw request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /wallet/withdraw [post]
func Withdraw(storage storages.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := auth.GetUserID(c)

		var req WalletOperation
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid amount or currency"})
			return
		}

		// Проверяем баланс
		current, err := storage.GetBalance(c.Request.Context(), userID, req.Currency)
		if err != nil || current < req.Amount {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds or invalid amount"})
			return
		}

		err = storage.UpdateBalance(c.Request.Context(), userID, req.Currency, -req.Amount)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
			return
		}

		balances, _ := storage.GetAllBalances(c.Request.Context(), userID)
		c.JSON(http.StatusOK, gin.H{
			"message":     "Withdrawal successful",
			"new_balance": balances,
		})
	}
}
