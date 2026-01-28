package notifications

import "time"

type LargeTransferEvent struct {
	UserID    int64     `json:"user_id"`
	Amount    float32   `json:"amount"`
	Currency  string    `json:"currency"`
	Timestamp time.Time `json:"timestamp"`
}
