package service

import (
	"context"
	"fmt"
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

	if (amount - s.cfg.globalFee) <= 0 {
		return nil, fmt.Errorf("not enought money for fee on this transaction")
	}

	acc.Balance += (amount - s.cfg.globalFee)

	return s.userRepo.UpdateAccountTransaction(ctx, acc, amount, s.cfg.globalFee, "deposit")
}

func (s *TransactionService) WithdrawalTransaction(ctx context.Context, acc models.Account, amount float64) (*models.Account, error) {

	if (acc.Balance - (amount + s.cfg.globalFee)) < 0 {
		return nil, fmt.Errorf("not enought money for transaction")
	}

	acc.Balance -= (amount + s.cfg.globalFee)

	return s.userRepo.UpdateAccountTransaction(ctx, acc, amount, s.cfg.globalFee, "withdrawal")
}
