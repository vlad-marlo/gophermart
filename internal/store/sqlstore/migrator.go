package sqlstore

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// migrate ...
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
