package dto

type TransactionRequestDto struct {
	AccountNumber string  `json:"account_number" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
}

type TransferRequestDto struct {
	SourceAccountNumber      string  `json:"source_account_number" validate:"required"`
	DestinationAccountNumber string  `json:"destination_account_number" validate:"required"`
	Amount                   float64 `json:"amount" validate:"required,gt=0"`
}
