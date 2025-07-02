package models

import "time"

type Card struct {
	Id        int
	AccountId int
	Number    []byte
	Expiry    []byte
	Cvv       []byte
	CreatedAt time.Time
}
