package sqlstore

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type storage struct {
	db     *pgxpool.Pool
	logger logger.Logger
	cfg    *pgxpool.Config

	// repositories
	user     store.UserRepository
	order    store.OrderRepository
	withdraw store.WithdrawRepository
}

// New ...
func New(ctx context.Context, l logger.Logger, c *config.Config) (store.Storage, error) {
	cfg, err := pgxpool.ParseConfig(c.DBURI)
	if err != nil {
		return nil, pgError("parse config: %v", err)
	}

	db, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, pgError("sql open: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, pgError("ping db: %v", err)
	}

	s := &storage{
		db:     db,
		logger: l,
		cfg:    cfg,
	}
	s.user = &userRepository{s}
	s.order = &orderRepository{s}
	s.withdraw = &withdrawRepository{s}

	if err := s.user.Migrate(context.Background()); err != nil {
		return nil, pgError("user: migrate: %v", err)
	}
	if err := s.order.Migrate(context.Background()); err != nil {
		return nil, pgError("orders: migrate: %v", err)
	}
	if err := s.withdraw.Migrate(context.Background()); err != nil {
		return nil, pgError("withdraws: migrate: %v", err)
	}

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
