package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/NathanRJohnson/live-backend/wtfridge/model"
	"github.com/NathanRJohnson/live-backend/wtfridge/repository/item"
)

type Item struct {
	Repo *item.FirebaseRepo
}

func (i *Item) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create an item")
	var body struct {
		ItemID   int    `json:"item_id"`
		Name     string `json:"item_name"`
		Quantity int    `json:"quantity"`
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("error unmarshaling requst:", err)
		return
	}

	// missing id
	if body.ItemID == 0 || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("item_name or item_is is missing or 0")
		return
	}

	if body.Quantity <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("quantity must exceed 0")
		return
	}

	now := time.Now().UTC()

	item := model.FridgeItem{
		ItemID:    body.ItemID,
		Name:      body.Name,
		Quantity:  body.Quantity,
		Notes:     body.Notes,
		DateAdded: &now,
	}

	// TODO: can I make the collection a const?
	err := i.Repo.Insert(r.Context(), "fridge", item)
	if err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(item)
	if err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
	w.WriteHeader(http.StatusCreated)

}

func (i *Item) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List all items - fridge")
	items, err := i.Repo.FetchAll(r.Context(), "fridge")
	if err != nil {
		fmt.Println("failed to fetch all:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	res, err := json.Marshal(items)
	if err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write(res)
}

func (i *Item) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Println("Get an item by ID: " + id)
}

func (i *Item) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an item by ID")

	var body struct {
		ItemID       int        `json:"item_id"`
		NewName      *string    `json:"new_name,omitempty"`
		NewQuantity  *int       `json:"new_quantity,omitempty"`
		NewNotes     *string    `json:"new_notes,omitempty"`
		NewDateAdded *time.Time `json:"new_date,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("error unmarshaling requst:", err)
		return
	}

	if body.ItemID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	new_values := make(map[string]interface{})

	if body.NewName != nil {
		if *body.NewName != "" {
			new_values["Name"] = *body.NewName
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if body.NewQuantity != nil {
		if *body.NewQuantity > 0 {
			new_values["Quantity"] = *body.NewQuantity
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if body.NewNotes != nil {
		new_values["Notes"] = *body.NewNotes
	}

	if body.NewDateAdded != nil {
		new_values["DateAdded"] = body.NewDateAdded
	}

	err := i.Repo.UpdateItemByID(r.Context(), "fridge", body.ItemID, new_values)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (i *Item) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an item by ID")

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Println("failed to convert id to integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = i.Repo.DeleteByID(r.Context(), "fridge", id)
	if err != nil {
		fmt.Println("failed to delete:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
