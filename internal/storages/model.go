package storages

type User struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

type Balance struct {
	UserID   int64   `json:"user_id"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}
