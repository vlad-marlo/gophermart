package sqlstore

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/vlad-marlo/gophermart/pkg/logger"

	"github.com/vlad-marlo/gophermart/internal/store"
)

// TestStore ...
func TestStore(t *testing.T, con string) (store.Storage, func(...string)) {
	t.Helper()

	l := logger.GetLogger()

	db, err := pgxpool.Connect(context.Background(), con)
	if err != nil {
		t.Fatalf("test store: sql open: %v", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("test store: db ping: %v", err)
	}

	s := &storage{
		db:     db,
		logger: l,
	}
	s.user = &userRepository{s}
	s.order = &orderRepository{s}
	s.withdraw = &withdrawRepository{s}

	if err := s.user.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if err := s.order.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if err := s.withdraw.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return s, func(tables ...string) {
		if len(tables) > 0 {
			if _, err = db.Exec(context.Background(), fmt.Sprintf("TRUNCATE %s CASCADE;", strings.Join(tables, ", "))); err != nil {
				t.Fatalf("%v", pgError("defer func: truncate test db: %v", err))
			}
		}
		logger.DeleteLogFolderAndFile(t)
		db.Close()
	}
}
