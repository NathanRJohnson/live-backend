package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	item := map[string]interface{}{
		"ItemID":    body.ItemID,
		"Name":      body.Name,
		"Quantity":  body.Quantity,
		"Notes":     body.Notes,
		"DateAdded": &now,
	}

	authHeader := r.Header.Get("Authorization")
	fridgeCollection, err := i.getCollectionFromHeader(authHeader, "FRIDGE")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = i.Repo.Insert(r.Context(), fridgeCollection, item)
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
}

func (i *Item) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List all items - fridge")

	authHeader := r.Header.Get("Authorization")
	fridgeCollection, err := i.getCollectionFromHeader(authHeader, "FRIDGE")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	items, err := i.Repo.FetchAll(r.Context(), fridgeCollection)

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

// func (i *Item) GetByID(w http.ResponseWriter, r *http.Request) {
// 	id := r.PathValue("id")
// 	fmt.Println("Get an item by ID: " + id)
// }

func (i *Item) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an item by ID")

	authHeader := r.Header.Get("Authorization")
	fridgeCollection, err := i.getCollectionFromHeader(authHeader, "FRIDGE")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

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

	err = i.Repo.UpdateItemByID(r.Context(), fridgeCollection, body.ItemID, new_values)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (i *Item) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an item by ID")

	authHeader := r.Header.Get("Authorization")
	fridgeCollection, err := i.getCollectionFromHeader(authHeader, "FRIDGE")
	if err != nil {
		log.Println("could not get claims for user:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get the doc id
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Println("failed to convert id to integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = i.Repo.DeleteByID(r.Context(), fridgeCollection, id)
	if err != nil {
		fmt.Println("failed to delete:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (i *Item) getCollectionFromHeader(authHeader string, collection string) (interface{}, error) {
	var grocery interface{}

	if authHeader == "" { // this exists solely to support legacy systems. TODO: remove once legacy system has been migrated
		grocery = i.Repo.GetCollectionRef(strings.ToLower(collection), nil)

	} else {
		// v2 endpoint
		claims, err := getUserClaimsFromHeader(authHeader)
		if err != nil {
			return nil, err
		}

		userCollection := i.Repo.GetCollectionRef("USER", nil)
		userDoc := i.Repo.GetDocRef(userCollection, claims.Username)
		grocery = i.Repo.GetCollectionRef(strings.ToUpper(collection), userDoc)
	}

	return grocery, nil
}
