package repository

import (
	"context"
	"uniback/dto"
	"uniback/models"
	//"uniback/models"
)

type Repository interface {
	// User methods
	IsUserExistsByUsernameEmailPhone(ctx context.Context)
	//CreateUser(ctx context.Context, user dto.UserCreateRequest) (int, error)
	//FindUserByUsername(ctx context.Context, username string) (*models.User, error)

	// Balance methods
	//GetBalanceByUsername(ctx context.Context, username string) (*models.Balance, error)
	//UpdateBalance(ctx context.Context, username string, amount float64) error

	// Connection managment
	Close() error
	Ping(ctx context.Context) error
}

type UserRepository interface {
	IsUserExistsByUsernameEmailPhone(ctx context.Context, userDto dto.UserCreateRequest) (username, email, phone bool, err error)
	CreateUser(ctx context.Context, userDto dto.UserCreateRequest) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	IsUserExists(ctx context.Context, username string) (bool, error)
	GetUserId(ctx context.Context, username string) (int, error)

	GetAccountsByUsername(ctx context.Context, username string) (*dto.AccountsResponseDto, error)
	IsAccountExits(ctx context.Context, accountNumber string) (bool, error)
	CreateAccount(ctx context.Context, acc models.Account) (*dto.AccountResponseDto, error)
	GetAccountByUsername(ctx context.Context, account string, username string) (*models.Account, error)

	UpdateAccountTransaction(ctx context.Context, acc models.Account, amount float64, fee float64, trsType string) (*models.Account, error)
}
