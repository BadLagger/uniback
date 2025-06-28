package repository

import (
	"context"
	"uniback/dto"
	"uniback/models"
)

type Repository interface {
	// User methods
	CreateUser(ctx context.Context, user dto.UserCreateRequest) (int, error)
	FindUserByUsername(ctx context.Context, username string) (*models.User, error)

	// Balance methods
	GetBalanceByUsername(ctx context.Context, username string) (*models.Balance, error)
	UpdateBalance(ctx context.Context, username string, amount float64) error

	// Connection managment
	Close() error
	Ping(ctx context.Context) error
}
