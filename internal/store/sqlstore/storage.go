package sqlstore

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type storage struct {
	db   *sql.DB
	user store.UserRepository
	l    *logrus.Logger
}

// New ...
func New(l *logrus.Logger, c *config.Config) (store.Storage, error) {
	db, err := sql.Open("postgres", c.DBURI)
	if err != nil {
		return nil, fmt.Errorf("sql open: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %v", err)
	}

	s := &storage{
		db:   db,
		user: &userRepository{db, l},
		l:    l,
	}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %v", err)
	}
	l.Debug("successfully migrated")
	return s, nil
}

// User ...
func (s *storage) User() store.UserRepository {
	return s.user
}

// Close ...
func (s *storage) Close() error {
	return s.db.Close()
}
