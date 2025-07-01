package dto

import (
	"time"
	"uniback/models"
)

type AccountResponseDto struct {
	AccountNumber string    `json:"account_number"`
	AccountType   string    `json:"account_type"`
	Balance       float64   `json:"balance"`
	OpeningDate   time.Time `json:"openin_date"`
	Status        string    `json:"status"`
}

type AccountsResponseDto struct {
	AccountsNum int                  `json:"accounts_num"`
	Acounts     []AccountResponseDto `json:"accounts"`
}

type AccountCreateRequestDto struct {
	AccountType string `json:"account_type" validate:"required,oneof=debit credit"`
}

func AccountToAccountReponseDto(acc *models.Account) *AccountResponseDto {
	return &AccountResponseDto{
		AccountNumber: acc.AccountNumber,
		AccountType:   acc.AccountType,
		Balance:       acc.Balance,
		OpeningDate:   acc.OpeningDate,
		Status:        acc.Status,
	}
}
