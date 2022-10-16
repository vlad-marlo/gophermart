package sqlstore

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type storage struct {
	db *pgxpool.Pool
	l  logger.Logger

	// repositories
	user  store.UserRepository
	order store.OrderRepository
}

// New ...
func New(ctx context.Context, l logger.Logger, c *config.Config) (store.Storage, error) {
	cfg, err := pgxpool.ParseConfig(c.DBURI)
	db, err := pgxpool.NewWithConfig(ctx, cfg)

	if err != nil {
		return nil, fmt.Errorf("sql open: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %v", err)
	}

	s := &storage{
		db:    db,
		user:  &userRepository{db, l},
		order: &orderRepository{db, l},
		l:     l,
	}

	//TODO hardcoded variable rewrite migrate args
	if err := s.migrate("", c.DBURI); err != nil {
		return nil, fmt.Errorf("migrate: %v", err)
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

// Close ...
func (s *storage) Close() {
	s.db.Close()
}
