package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
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

func TestStore(t *testing.T, con string) (store.Storage, func(...string)) {
	t.Helper()

	l := logger.Logger{Entry: logrus.NewEntry(logrus.New())}

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
	if err := s.migrate(); err != nil {
		t.Fatalf("test store: sql migrate: %v", err)
	}
	return s, func(tables ...string) {
		if len(tables) > 0 {
			_, _ = db.Exec("TRUNCATE $1 CASCADE", pq.Array(tables))
		}
		_ = db.Close()
	}
}
