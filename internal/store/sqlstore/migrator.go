package sqlstore

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// migrate ...
func (s *storage) migrate(source string) error {
	if source == "" {
		source = "file://migrations"
	}
	driver, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("get driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		source,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("with db instance: %v", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("up: %v", err)
	}
	return nil
}
