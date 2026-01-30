package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"gw-currency-wallet/internal/auth"
	"gw-currency-wallet/internal/auth/mocks"
	"gw-currency-wallet/internal/storages"
	"gw-currency-wallet/pkg/logging"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockAuthService struct{}

func (m *MockAuthService) GetExchangeRateWithCache(from, to string) (float32, error) {
	return 90.0, nil
}

type MockStorage struct{}

func (m *MockStorage) CreateUser(ctx context.Context, email, passwordHash string) (int64, error) {
	return 1, nil
}

func (m *MockStorage) GetUserByEmail(ctx context.Context, email string) (storages.User, error) {
	return storages.User{ID: 1, Email: email}, nil
}

func (m *MockStorage) GetBalance(ctx context.Context, userID int64, currency string) (float32, error) {
	return 1000.0, nil
}
func (m *MockStorage) GetAllBalances(ctx context.Context, userID int64) (map[string]float32, error) {
	return make(map[string]float32), nil
}

func (m *MockStorage) UpdateBalance(ctx context.Context, userID int64, currency string, amount float32) error {
	return nil
}

func TestExchangeHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	reqBody := ExchangeRequest{
		FromCurrency: "USD",
		ToCurrency:   "RUB",
		Amount:       100,
	}
	jsonBody, _ := json.Marshal(reqBody)

	router.POST("/exchange", func(c *gin.Context) {
		c.Set("userID", int64(1))
		mockStorage := &MockStorage{} // с реализованными методами
		mockClient := &mocks.MockExchangerClient{}
		authService := auth.NewService(mockStorage, "test-secret", mockClient, logging.GetLogger())

		Exchange(mockStorage, authService, nil)(c)
	})

	req := httptest.NewRequest("POST", "/exchange", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "received_amount")
}
