package storages

import "context"

type Repository interface {
	//Users
	CreateUser(ctx context.Context, email, passwordHash string) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)

	//Currencies
	GetBalance(ctx context.Context, userID int64, currency string) (float32, error)
	GetAllBalances(ctx context.Context, userID int64) (map[string]float32, error)
	UpdateBalance(ctx context.Context, userID int64, currency string, amount float32) error
}
