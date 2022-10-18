package sqlstore

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vlad-marlo/gophermart/pkg/logger"

	"github.com/vlad-marlo/gophermart/internal/store"
)

// TestStore ...
func TestStore(t *testing.T, con string) (store.Storage, func(...string)) {
	t.Helper()

	l := logger.GetLogger()
	logger.DeleteLogFolderAndFile(t)

	db, err := pgxpool.New(context.TODO(), con)
	if err != nil {
		t.Fatalf("test store: sql open: %v", err)
	}

	if err := db.Ping(context.TODO()); err != nil {
		t.Fatalf("test store: db ping: %v", err)
	}

	s := &storage{
		db:   db,
		user: &userRepository{db, l},
		l:    l,
	}
	source := "file://../../../migrations"
	if err := s.migrate(source, con); err != nil {
		t.Fatalf("test store: sql migrate: %v", err)
	}
	return s, func(tables ...string) {
		if len(tables) > 0 {
			_, _ = db.Exec(context.TODO(), fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", ")))
		}
		db.Close()
	}
}
