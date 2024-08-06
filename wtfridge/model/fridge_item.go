package model

import (
	"time"
)

type FridgeItem struct {
	ItemID    int        `json:"item_id"`
	Name      string     `json:"item_name"`
	DateAdded *time.Time `json:"date_added"`
	Quantity  int        `json:"quantity"`
	Notes     string     `json:"notes"`
}

func (f FridgeItem) GetID() int {
	return f.ItemID
}
