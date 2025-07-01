package dto

type TransactionRequestDto struct {
	AccountNumber string  `json:"account_number" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
}
