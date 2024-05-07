package model

import "time"

type Item struct {
	ItemID    int        `json:"item_id"`
	Name      string     `json:"item_name"`
	DateAdded *time.Time `json:"date_added"`
}
