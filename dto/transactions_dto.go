package dto

type DepositRequestDto struct {
	AccountNumber string  `json:"account_number" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
}
