package sqlstore

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type orderRepository struct {
	db *pgxpool.Pool
	l  logger.Logger
}
