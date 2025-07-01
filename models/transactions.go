package models

import "time"

type Transaction struct {
	Id        int
	AccountId int
	Type      string
	Amount    float64
	Time      time.Time
	Fee       float64
}

type TransactionTransfer struct {
	Id            int
	TransId       int
	DestAccountId int
}
