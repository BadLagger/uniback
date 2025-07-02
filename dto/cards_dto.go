package dto

type CardCreateRequestDto struct {
	AccountNumber string `json:"account_number" validate:"required"`
}
