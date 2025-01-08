package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/NathanRJohnson/live-backend/wtfridge/repository/item"
)

type Grocery struct {
	Repo *item.FirebaseRepo
}

// func (g *Grocery) Create(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("Create a grocery item")

// 	var body struct {
// 		ItemID   int    `json:"item_id"`
// 		Name     string `json:"item_name"`
// 		IsActive bool   `json:"is_active"`
// 		Index    int    `json:"index"`
// 		Quantity int    `json:"quantity"`
// 		Notes    string `json:"notes"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		fmt.Println("error unmarshaling requst:", err)
// 		return
// 	}

// 	if body.ItemID == 0 || body.Name == "" || body.Index < 1 {
// 		w.WriteHeader(http.StatusBadRequest)
// 		fmt.Println("required field is missing")
// 		return
// 	}

// 	item := model.GroceryItem{
// 		ItemID:   body.ItemID,
// 		Name:     body.Name,
// 		IsActive: body.IsActive,
// 		Index:    body.Index,
// 		Quantity: body.Quantity,
// 		Notes:    body.Notes,
// 	}

// 	err := g.Repo.Insert(r.Context(), "grocery", item)
// 	if err != nil {
// 		fmt.Println("failed to insert:", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	res, err := json.Marshal(item)
// 	if err != nil {
// 		fmt.Println("failed to marshal grocery:", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	w.Write(res)
// }

// func (g *Grocery) List(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("List all grocery items")
// 	items, err := g.Repo.FetchAll(r.Context(), "grocery")
// 	if err != nil {
// 		fmt.Println("failed to fetch all:", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 	}

// 	res, err := json.Marshal(items)
// 	if err != nil {
// 		fmt.Println("failed to marshal:", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 	}

// 	w.Write(res)
// }

func (g *Grocery) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an item by ID")

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Println("failed to convert id to integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = g.Repo.DeleteByID(r.Context(), "grocery", id)
	if err != nil {
		fmt.Println("failed to delete:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Grocery) SetActiveByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Change active state")

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println("failed to convert id to integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = g.Repo.ToggleActiveByID(r.Context(), "grocery", id)
	if err != nil {
		log.Println("failed to toggle active state")
		return
	}
}

func (g *Grocery) UpdateByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Update by ID")

	var body struct {
		ItemID      int    `json:"item_id"`
		NewName     string `json:"new_name"`
		NewQuantity int    `json:"new_quantity"`
		NewNotes    string `json:"new_notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("error unmarshaling requst:", err)
		return
	}

	if body.ItemID <= 0 || body.NewName == "" || body.NewQuantity <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	new_values := map[string]interface{}{
		"Name":     body.NewName,
		"Quantity": body.NewQuantity,
		"Notes":    body.NewNotes,
	}

	err := g.Repo.UpdateItemByID(r.Context(), "grocery", body.ItemID, new_values)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (g *Grocery) MoveToFridge(w http.ResponseWriter, r *http.Request) {
	log.Println("Move items to fridge")

	err := g.Repo.MoveToFridge(r.Context())
	if err != nil {
		log.Printf("failed to move grocery items to fridge: %v", err)
	}
}

func (g *Grocery) RearrageItems(w http.ResponseWriter, r *http.Request) {
	log.Println("Rearrage items")

	var body struct {
		OldIndex int64 `json:"old_index"`
		NewIndex int64 `json:"new_index"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("error unmarshaling requst:", err)
		return
	}

	if body.OldIndex == body.NewIndex || body.OldIndex <= 0 || body.NewIndex <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("indicies must be > 0 and not equal")
		return
	}

	err := g.Repo.RearrageItems(r.Context(), "grocery", body.OldIndex, body.NewIndex)
	if err != nil {
		log.Printf("failed to rearrage items: %v", err)
	}
}
