package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/vlad-marlo/gophermart/pkg/logger"
	"strings"
	"testing"

	"github.com/vlad-marlo/gophermart/internal/store"
)

func TestStore(t *testing.T, con string) (store.Storage, func(...string)) {
	t.Helper()

	l := logger.GetLogger()
	logger.DeleteLogFolderAndFile()

	db, err := sql.Open("postgres", con)
	if err != nil {
		t.Fatalf("test store: sql open: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("test store: db ping: %v", err)
	}

	s := &storage{
		db:   db,
		user: &userRepository{db, l},
		l:    l,
	}
	source := "file://../../../migrations"
	if err := s.migrate(source); err != nil {
		t.Fatalf("test store: sql migrate: %v", err)
	}
	return s, func(tables ...string) {
		if len(tables) > 0 {
			_, _ = db.Exec(fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", ")))
		}
		_ = db.Close()
	}
}
