package sqlstore_test

import (
	"os"
	"testing"
)

var (
	conStr string
)

func TestMain(m *testing.M) {
	conStr = os.Getenv("TEST_DB_URI")
	if conStr == "" {
		conStr = "host=localhost dbname=gophermart_test"
	}
	os.Exit(m.Run())
}
