package service

import (
	"context"
	"uniback/models"
)

type Service interface {
	DepositTransaction(ctx context.Context, acc models.Account, amount float64) (*models.Account, error)
	WithdrawalTransaction(ctx context.Context, acc models.Account, amount float64) (*models.Account, error)
	TransferTransaction(ctx context.Context, source models.Account, dest models.Account, amount float64) (*models.Account, error)
}

type CryptoService interface {
	PgpEncode(data string) []byte
	PgpDecode(data []byte) string
	GenerateCardLuhn() (string, error)
	GenerateCvv() string
}
