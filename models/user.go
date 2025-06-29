package models

type User struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Password string `json:"-" db:"password"`
	Email    string `json:"email" db:"email" validate:"required,email"`
	Phone    string `json:"phone" db:"phone" validate:"omitempty,e164"`
}
