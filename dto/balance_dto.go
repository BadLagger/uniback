package dto

type BalanceResponse struct {
	UserId  int     `json:"user_id"`
	Balance float64 `json:"balance"`
}

type BalanceUpdateRequest struct {
	Username string  `json:"username" validate:"required"`
	Amount   float64 `json:"amount" validate:"required"`
}
