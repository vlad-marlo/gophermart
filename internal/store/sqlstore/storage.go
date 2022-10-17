package sqlstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

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

func (s *storage) migrate(sourceUrl, databaseUrl string) error {
	if sourceUrl == "" {
		sourceUrl = "file://migrations"
	}
	m, err := migrate.New(
		sourceUrl,
		databaseUrl,
	)
	if err != nil {
		return fmt.Errorf("new migrate example: %v", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("up: %v", err)
	}
	return nil
}
