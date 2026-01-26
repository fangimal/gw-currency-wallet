package handlers

import (
	"gw-currency-wallet/internal/auth"
	"gw-currency-wallet/internal/storages"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetBalance(storage storages.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := auth.GetUserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		currency := c.Param("currency")
		if currency == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "currency is required"})
			return
		}

		balance, err := storage.GetBalance(c.Request.Context(), userID, currency)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "balance not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"balance": balance})
	}
}
