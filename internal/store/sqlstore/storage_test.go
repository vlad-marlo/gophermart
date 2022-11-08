package sqlstore_test

import (
	"os"
)

var (
	conStr               = os.Getenv("TEST_DB_URI")
	userLogin1           = "first"
	userLogin2           = "second"
	userTableName        = "users"
	ordersTableName      = "orders"
	withdrawalsTableName = "withdrawals"
	orderNum1            = 79927398713
	orderNum2            = 4929972884676289
	orderNum3            = 4532733309529845
	orderNum4            = 4539088167512356
)
