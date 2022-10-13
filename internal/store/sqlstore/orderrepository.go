package sqlstore

import (
	"database/sql"

	"github.com/vlad-marlo/gophermart/internal/pkg/logger"
)

type orderRepository struct {
	db *sql.DB
	l  logger.Logger
}
