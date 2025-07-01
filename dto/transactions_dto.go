package dto

type DepositRequestDto struct {
	AccountId string  `json:"id" validate:"required"`
	Amount    float64 `json:"amount" validate:"required,gt=0,decimal=2"`
}
