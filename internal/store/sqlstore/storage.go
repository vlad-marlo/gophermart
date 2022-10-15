package sqlstore

import (
	"database/sql"
	"fmt"

	"github.com/vlad-marlo/gophermart/internal/pkg/logger"

	_ "github.com/lib/pq"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type storage struct {
	db *sql.DB
	l  *logger.Logger

	// repositoryes
	user  store.UserRepository
	order store.OrderRepository
}

// New ...
func New(l *logger.Logger, c *config.Config) (store.Storage, error) {
	db, err := sql.Open("postgres", c.DBURI)
	if err != nil {
		return nil, fmt.Errorf("sql open: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %v", err)
	}

	s := &storage{
		db:    db,
		user:  &userRepository{db, l},
		order: &orderRepository{db, l},
		l:     l,
	}

	//TODO hardcoded variable rewrite migrate args
	if err := s.migrate(""); err != nil {
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
func (s *storage) Close() error {
	return s.db.Close()
}
