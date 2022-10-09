package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/pkg/logger"
	"github.com/vlad-marlo/gophermart/internal/store"
	"strings"
	"testing"
)

func TestDB(t *testing.T, databaseURL string) (*sql.DB, func(...string)) {
	t.Helper()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	return db, func(tables ...string) {
		if len(tables) > 0 {
			_, _ = db.Exec(fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", ")))
		}

		_ = db.Close()
	}
}

func TestStore(con string) (store.Storage, error) {
	l := logger.Logger{Entry: logrus.NewEntry(logrus.New())}
	db, err := sql.Open("postgres", con)
	if err != nil {
		return nil, fmt.Errorf("sql open: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %v", err)
	}

	s := &storage{
		db:   db,
		user: &userRepository{db},
		l:    l,
	}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %v", err)
	}
	l.Info("successfully migrated")
	return s, nil
}
