package models

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type Account struct {
	Id            int
	UserId        int
	AccountNumber string
	AccountType   string
	Balance       float64
	OpeningDate   time.Time
	Status        string
}

func GenerateAccount() string {
	// simulate ZZZ control key
	controlKey := fmt.Sprintf("%03d", rand.Intn(1000))
	// simulate check K
	checkDigit := strconv.Itoa(rand.Intn(10))

	// Unique random part 10 digit
	uniquePart := ""
	for i := 0; i < 10; i++ {
		uniquePart += strconv.Itoa(rand.Intn(10))
	}
	// accountType always for Private Person
	accountType := "408"
	// currencyCode always for RUB
	currencyCode := "810"
	// Generate number only for Private and RUB
	return accountType + currencyCode + controlKey + checkDigit + uniquePart
}
