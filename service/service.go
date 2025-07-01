package service

import (
	"context"
	"uniback/models"
)

type Service interface {
	DepositTransaction(ctx context.Context, acc models.Account, amount float64) (*models.Account, error)
}
