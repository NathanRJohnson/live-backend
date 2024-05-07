package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/NathanRJohnson/live-backend/wtfridge/model"
	"github.com/NathanRJohnson/live-backend/wtfridge/repository/item"
)

type Item struct {
	Repo *item.FirebaseRepo
}

func (i *Item) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create an items")
	var body struct {
		Name string `json:"item_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()

	item := model.Item{
		ItemID:    rand.Int(),
		Name:      body.Name,
		DateAdded: &now,
	}

	err := i.Repo.Insert(r.Context(), item)
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
	fmt.Println("List all items")
}

func (i *Item) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Println("Get an item by ID: " + id)
}

func (i *Item) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an item by ID")
}

func (i *Item) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an item by ID")
}
