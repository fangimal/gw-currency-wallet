package auth

import (
	"context"
	"gw-currency-wallet/internal/storages"
	"gw-currency-wallet/pkg/logging"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) CreateUser(ctx context.Context, email, passwordHash string) (int64, error) {
	args := m.Called(ctx, email, passwordHash)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorage) GetUserByEmail(ctx context.Context, email string) (storages.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(storages.User), args.Error(1)
}

func (m *MockStorage) GetBalance(ctx context.Context, userID int64, currency string) (float32, error) {
	args := m.Called(ctx, userID, currency)
	return args.Get(0).(float32), args.Error(1)
}

func (m *MockStorage) GetAllBalances(ctx context.Context, userID int64) (map[string]float32, error) {
	return make(map[string]float32), nil
}

func (m *MockStorage) UpdateBalance(ctx context.Context, userID int64, currency string, amount float32) error {
	args := m.Called(ctx, userID, currency, amount)
	return args.Error(0)
}

func TestAuth_Register(t *testing.T) {
	storage := new(MockStorage)
	storage.On("CreateUser", mock.Anything, "test2@example.com", mock.Anything).Return(int64(1), nil)
	logger := logging.GetLogger()
	service := NewService(storage, "secret", nil, logger)
	err := service.Register(context.Background(), "test2@example.com", "password")

	assert.NoError(t, err)
	storage.AssertExpectations(t)
}

func TestAuth_Login(t *testing.T) {
	storage := new(MockStorage)

	// Создаем корректный хеш для пароля "password"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	user := storages.User{ID: 1, Email: "test2@example.com", PasswordHash: string(passwordHash)}
	storage.On("GetUserByEmail", mock.Anything, "test2@example.com").Return(user, nil)
	logger := logging.GetLogger()
	service := NewService(storage, "secret", nil, logger)
	token, err := service.Login(context.Background(), "test2@example.com", "password")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
