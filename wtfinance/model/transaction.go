package model

import "time"

type Transaction struct {
	DateCreated *time.Time `json:"date"`
	Name        string     `json:"name"`
	Amount      float32    `json:"amount"`
	Category    string     `json:"category"`
	SheetRef    string     `json:"sheet"`
}
