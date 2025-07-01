package models

import "time"

type Balance struct {
	Id             int
	UserId         int
	AcccountNumber string
	AccountType    string
	Balance        float64
	OpeningDate    time.Time
	Status         string
}
