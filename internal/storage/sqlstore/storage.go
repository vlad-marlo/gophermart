package sqlstore

import "database/sql"

type storage struct {
	db   *sql.DB
	user *userRepository
}

// New ...
func New(db *sql.DB) *storage {
	ur := newUserRepository(db)
	return &storage{db: db, user: ur}
}

// User ...
func (s *storage) User() *userRepository {
	return s.user
}

func (s *storage) Close() error {
	return s.db.Close()
}
