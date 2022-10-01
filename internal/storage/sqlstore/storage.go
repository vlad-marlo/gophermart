package sqlstore

import (
	"database/sql"
	"fmt"

	"github.com/vlad-marlo/gophermart/internal/config"
)

type storage struct {
	db   *sql.DB
	user *userRepository
}

// New ...
func New(c *config.Config) (*storage, error) {
	db, err := sql.Open("", c.DBURI)
	if err != nil {
		return nil, fmt.Errorf("sql open: %v", err)
	}
	ur := newUserRepository(db)
	return &storage{db: db, user: ur}, nil
}

// User ...
func (s *storage) User() *userRepository {
	return s.user
}

// Close
func (s *storage) Close() error {
	return s.db.Close()
}
