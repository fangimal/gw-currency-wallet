package postgres

import (
	"context"
	"errors"
	"fmt"
	"gw-currency-wallet/internal/storages"

	"github.com/jackc/pgx/v5"
)

// Users
func (p *Postgres) CreateUser(ctx context.Context, email, passwordHash string) (int64, error) {
	var userID int64
	err := p.Client.QueryRow(ctx,
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		email, passwordHash,
	).Scan(&userID)

	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	// Инициализируем балансы по умолчанию (0 для всех валют)
	currencies := []string{"USD", "RUB", "EUR"}
	for _, currency := range currencies {
		_, err := p.Client.Exec(ctx,
			"INSERT INTO balances (user_id, currency, amount) VALUES ($1, $2, 0)",
			userID, currency,
		)
		if err != nil {
			p.logger.Warn("failed to initialize balance for currency", "currency", currency, "error", err)
		}
	}

	return userID, nil
}
func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (storages.User, error) {
	var user storages.User
	err := p.Client.QueryRow(ctx,
		"SELECT id, email, password_hash FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, fmt.Errorf("user not found")
		}
		return user, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (p *Postgres) GetBalance(ctx context.Context, userID int64, currency string) (float32, error) {
	var amount float32
	err := p.Client.QueryRow(ctx,
		"SELECT amount FROM balances WHERE user_id = $1 AND currency = $2",
		userID, currency,
	).Scan(&amount)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("balance not found for currency %s", currency)
		}
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return amount, nil
}

func (p *Postgres) GetAllBalances(ctx context.Context, userID int64) (map[string]float32, error) {
	sql := "SELECT currency, amount FROM balances WHERE user_id = $1"
	rows, err := p.Client.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := make(map[string]float32)
	for rows.Next() {
		var currency string
		var amount float32
		if err = rows.Scan(&currency, &amount); err != nil {
			return nil, err
		}
		balances[currency] = amount
	}
	return balances, nil
}

func (p *Postgres) UpdateBalance(ctx context.Context, userID int64, currency string, amount float32) error {
	result, err := p.Client.Exec(ctx,
		"UPDATE balances SET amount = amount + $1 WHERE user_id = $2 AND currency = $3",
		amount, userID, currency,
	)

	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("balance record not found for user %d and currency %s", userID, currency)
	}

	return nil
}
