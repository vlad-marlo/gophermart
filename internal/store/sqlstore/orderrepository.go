package sqlstore

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type orderRepository struct {
	db *pgxpool.Pool
	l  logger.Logger
}

func (o *orderRepository) Register(ctx context.Context, user int, number int) error {
	//TODO implement me
	panic("implement me")
}

func (o *orderRepository) GetAllByUser(ctx context.Context, user int) ([]*model.Order, error) {
	//TODO implement me
	panic("implement me")
}
