package dto

type UserCreateRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"omitempty,e164"`
}

type UserLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}
