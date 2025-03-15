package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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
		Index    int    `json:"index"`
		Quantity int    `json:"quantity"`
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("error unmarshaling requst:", err)
		return
	}

	if body.ItemID == 0 || body.Name == "" || body.Index < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("required field is missing")
		return
	}

	if body.Quantity <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("quantity must exceed 0")
		return
	}

	item := map[string]interface{}{
		"ItemID":   body.ItemID,
		"Name":     body.Name,
		"IsActive": body.IsActive,
		"Index":    body.Index,
		"Quantity": body.Quantity,
		"Notes":    body.Notes,
	}

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := g.getCollectionFromHeader(authHeader, "GROCERY")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = g.Repo.Insert(r.Context(), groceryCollection, item)
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
	fmt.Println("List all items - grocery")

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := g.getCollectionFromHeader(authHeader, "GROCERY")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	items, err := g.Repo.FetchAll(r.Context(), groceryCollection)
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

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := g.getCollectionFromHeader(authHeader, "GROCERY")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Println("failed to convert id to integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = g.Repo.DeleteByID(r.Context(), groceryCollection, id)
	if err != nil {
		fmt.Println("failed to delete:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Grocery) SetActiveByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Change active state")

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := g.getCollectionFromHeader(authHeader, "GROCERY")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println("failed to convert id to integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = g.Repo.ToggleActiveByID(r.Context(), groceryCollection, id)
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

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := g.getCollectionFromHeader(authHeader, "GROCERY")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
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

	err = g.Repo.UpdateItemByID(r.Context(), groceryCollection, body.ItemID, new_values)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (g *Grocery) MoveToFridge(w http.ResponseWriter, r *http.Request) {
	log.Println("Move items to fridge")

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := g.getCollectionFromHeader(authHeader, "GROCERY")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	fridgeCollection, err := g.getCollectionFromHeader(authHeader, "FRIDGE")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = g.Repo.MoveToFridge(r.Context(), groceryCollection, fridgeCollection)
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

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := g.getCollectionFromHeader(authHeader, "GROCERY")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
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

	err = g.Repo.RearrageItems(r.Context(), groceryCollection, body.OldIndex, body.NewIndex)
	if err != nil {
		log.Printf("failed to rearrage items: %v", err)
	}
}

func (g *Grocery) getCollectionFromHeader(authHeader string, collection string) (interface{}, error) {
	var grocery interface{}

	if authHeader == "" { // this exists solely to support legacy systems. TODO: remove once legacy system has been migrated
		grocery = g.Repo.GetCollectionRef(strings.ToLower(collection), nil)

	} else {
		// v2 endpoint
		claims, err := getUserClaimsFromHeader(authHeader)
		if err != nil {
			return nil, err
		}

		userCollection := g.Repo.GetCollectionRef("USER", nil)
		userDoc := g.Repo.GetDocRef(userCollection, claims.Username)
		grocery = g.Repo.GetCollectionRef(strings.ToUpper(collection), userDoc)
	}

	return grocery, nil
}
