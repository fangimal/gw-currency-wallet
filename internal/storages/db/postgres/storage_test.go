package postgres

import (
	"context"
	"fmt"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/pkg/logging"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPostgresStorage(t *testing.T) {
	// Подключение к тестовой БД (должна быть запущена)

	logger := logging.GetLogger()

	cfg := &config.StorageConfig{
		Host:     "localhost",
		Port:     "5438",
		User:     "postgres",
		Password: "postgres",
		Name:     "wallet_test_db",
	}

	repo, closeDB := NewPostgresRepository(context.Background(), cfg, logger)
	assert.NotNil(t, repo)
	defer closeDB()

	// Приведем к типу *Postgres для доступа к методам
	storage := repo.(*Postgres)

	// Создание структуры таблиц перед тестированием
	_, err := storage.Client.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users(
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS balances(
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			currency VARCHAR(3) NOT NULL,
			amount DECIMAL(15,2) NOT NULL CHECK ( amount >= 0 ),
			PRIMARY KEY (user_id, currency)
		);

		CREATE INDEX IF NOT EXISTS idx_balances_user_currency ON balances(user_id, currency);
	`)
	assert.NoError(t, err)

	// Используем уникальный email для каждого запуска теста
	email := fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())

	// Создание пользователя
	userID, err := storage.CreateUser(context.Background(), email, "hash")
	assert.NoError(t, err)
	assert.Greater(t, userID, int64(0))

	// Проверка баланса
	balance, err := storage.GetBalance(context.Background(), userID, "USD")
	assert.NoError(t, err)
	assert.Equal(t, float32(0), balance)

	// Обновление баланса
	err = storage.UpdateBalance(context.Background(), userID, "USD", float32(100.5))
	assert.NoError(t, err)

	balance, err = storage.GetBalance(context.Background(), userID, "USD")
	assert.NoError(t, err)
	assert.Equal(t, float32(100.5), balance)

	// Очистка данных после теста
	_, err = storage.Client.Exec(context.Background(), "DELETE FROM balances WHERE user_id = $1", userID)
	assert.NoError(t, err)

	_, err = storage.Client.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	assert.NoError(t, err)
}
