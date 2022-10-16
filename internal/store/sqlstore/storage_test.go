package sqlstore_test

import (
	"os"
	"testing"
)

var (
	conStr string
)

// TestMain ...
func TestMain(m *testing.M) {
	conStr = os.Getenv("TEST_DB_URI")
	if conStr == "" {
		conStr = "postgres://marlo:819819@localhost:5432/gophermart_test?sslmode=disable"
	}
	os.Exit(m.Run())
}
