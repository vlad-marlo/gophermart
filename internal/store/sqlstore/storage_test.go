package sqlstore_test

import (
	"os"
	"testing"
)

var (
	conStr               string
	userLogin1           = "first"
	userLogin2           = "second"
	userTableName        = "users"
	ordersTableName      = "orders"
	withdrawalsTableName = "withdrawals"
	orderNum1            = 1
	orderNum2            = 2
	orderNum3            = 3
	orderNum4            = 4
)

// TestMain ...
func TestMain(m *testing.M) {
	conStr = os.Getenv("TEST_DB_URI")
	os.Exit(m.Run())
}
