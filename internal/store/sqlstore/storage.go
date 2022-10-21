package sqlstore

import (
	"context"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	pglogrus "github.com/jackc/pgx-logrus"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type storage struct {
	db     *pgxpool.Pool
	logger logger.Logger

	// repositories
	user     store.UserRepository
	order    store.OrderRepository
	withdraw store.WithdrawRepository
}

// New ...
func New(ctx context.Context, l logger.Logger, c *config.Config) (store.Storage, error) {
	cfg, err := pgxpool.ParseConfig(c.DBURI)

	// logger for pgx
	t := &tracelog.TraceLog{
		Logger:   pglogrus.NewLogger(l.GetEntry()),
		LogLevel: tracelog.LogLevel(l.GetLevel()),
	}
	cfg.ConnConfig.Tracer = t

	db, err := pgxpool.NewWithConfig(ctx, cfg)

	if err != nil {
		return nil, fmt.Errorf("sql open: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %v", err)
	}

	s := &storage{
		db:     db,
		logger: l,
	}
	s.user = &userRepository{s}
	s.order = &orderRepository{s}
	s.withdraw = &withdrawRepository{s}

	s.user.Migrate(context.Background())
	s.order.Migrate(context.Background())
	s.withdraw.Migrate(context.Background())

	return s, nil
}

// User ...
func (s *storage) User() store.UserRepository {
	return s.user
}

// Order ...
func (s *storage) Order() store.OrderRepository {
	return s.order
}

// Withdraws ...
func (s *storage) Withdraws() store.WithdrawRepository {
	return s.withdraw
}

// Close ...
func (s *storage) Close() {
	s.db.Close()
}
