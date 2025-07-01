package dto

import "time"

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
