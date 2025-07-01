package service

import (
	"context"
	"uniback/models"
	"uniback/repository"
)

type TransacrionServiceConfig struct {
	globalFee float64
}

type TransactionService struct {
	userRepo repository.UserRepository
	cfg      TransacrionServiceConfig
}

func NewTransactionService(u repository.UserRepository) *TransactionService {
	return &TransactionService{
		userRepo: u,
		cfg: TransacrionServiceConfig{
			globalFee: 0.,
		},
	}
}

func (s *TransactionService) DepositTransaction(ctx context.Context, acc models.Account, amount float64) (*models.Account, error) {
	return s.userRepo.DepositToAccountTransaction(ctx, acc, amount, s.cfg.globalFee)
}
