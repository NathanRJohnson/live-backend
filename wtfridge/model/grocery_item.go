package model

type GroceryItem struct {
	ItemID   int    `json:"item_id"`
	Name     string `json:"item_name"`
	IsActive bool   `json:"is_active"`
}

func (g GroceryItem) GetID() int {
	return g.ItemID
}
