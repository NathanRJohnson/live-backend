package model

import "time"

type Transaction struct {
	DateCreated *time.Time `json:"date"`
	Name        string     `json:"name"`
	Category    string     `json:"category"`
	Amount      float32    `json:"amount"`
}
