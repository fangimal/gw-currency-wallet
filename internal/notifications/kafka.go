package notifications

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type NotificationService struct {
	writer *kafka.Writer
}

type TransferEvent struct {
	UserID    int64     `json:"user_id"`
	Amount    float32   `json:"amount"`
	Currency  string    `json:"currency"`
	Timestamp time.Time `json:"timestamp"`
}

func NewNotificationService(broker, topic string) *NotificationService {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(broker),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true, // создаёт топик, если не существует
	}
	return &NotificationService{writer: writer}
}

func (ns *NotificationService) SendLargeTransfer(ctx context.Context, userID int64, amount float32, currency string) error {
	event := TransferEvent{
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ns.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
}

func (ns *NotificationService) Close() error {
	return ns.writer.Close()
}
