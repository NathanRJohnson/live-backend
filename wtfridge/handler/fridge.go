package handler

import (
	"encoding/json"
	"fmt"
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
	fmt.Println("Create an items")
	var body struct {
		ItemID int    `json:"item_id"`
		Name   string `json:"item_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// missing id
	if body.ItemID == 0 || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()

	item := model.Item{
		ItemID:    body.ItemID,
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
	fmt.Println("List all items - test")
	items, err := i.Repo.FetchAll(r.Context())
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
	w.WriteHeader(http.StatusOK)
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

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Println("failed to convert id to integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	item := model.Item{
		ItemID: id,
	}

	err = i.Repo.DeleteByID(r.Context(), item)
	if err != nil {
		fmt.Println("failed to delete:", err)
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
}
