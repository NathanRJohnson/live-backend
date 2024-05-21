package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/NathanRJohnson/live-backend/wtfridge/model"
	"github.com/NathanRJohnson/live-backend/wtfridge/repository/item"
)

type Grocery struct {
	Repo *item.FirebaseRepo
}

func (g *Grocery) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create a grocery item")

	var body struct {
		ItemID   int    `json:"item_id"`
		Name     string `json:"item_name"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("error unmarshaling requst:", err)
		return
	}

	if body.ItemID == 0 || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("item_name or item_is is missing or 0")
		return
	}

	item := model.GroceryItem{
		ItemID:   body.ItemID,
		Name:     body.Name,
		IsActive: body.IsActive,
	}

	err := g.Repo.Insert(r.Context(), "grocery", item)
	if err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(item)
	if err != nil {
		fmt.Println("failed to marshal grocery:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

func (g *Grocery) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List all grocery items")
	items, err := g.Repo.FetchAll(r.Context(), "grocery")
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

func (g *Grocery) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an item by ID")

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Println("failed to convert id to integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	item := model.GroceryItem{
		ItemID: id,
	}

	err = g.Repo.DeleteByID(r.Context(), "grocery", item)
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
