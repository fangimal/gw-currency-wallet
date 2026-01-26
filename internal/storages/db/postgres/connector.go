package postgres

import (
	"context"
	"fmt"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/storages"
	"gw-currency-wallet/pkg/logging"
	"gw-currency-wallet/pkg/repeatable"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Client *pgxpool.Pool
	logger *logging.Logger
	cfg    config.StorageConfig
}

func NewPostgresRepository(ctx context.Context, cfg *config.StorageConfig, log *logging.Logger) (storages.Repository, func()) {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
	maxAttempts := 3
	var pool *pgxpool.Pool
	err := repeatable.DoWithTries(func() error {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var err error
		pool, err = pgxpool.New(ctx, dsn)
		if err != nil {
			return err
		}
		return nil
	}, maxAttempts, 5*time.Second)

	if err != nil {
		log.Fatal("error do with tries postgresql")
	}

	log.Info("connected to PostgreSQL")

	return &Postgres{
		Client: pool,
		logger: log,
		cfg:    *cfg,
	}, pool.Close
}
