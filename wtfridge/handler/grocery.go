package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// type Grocery struct {
// 	Repo *item.FirebaseRepo
// }

func (db *DB) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create a grocery item")

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := db.getCollectionFromHeader(authHeader, GROCERY)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

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

	item := map[string]interface{}{
		"ItemID":   body.ItemID,
		"Name":     body.Name,
		"IsActive": body.IsActive,
		"Index":    body.Index,
		"Quantity": body.Quantity,
		"Notes":    body.Notes,
	}

	err = db.Repo.Insert(r.Context(), groceryCollection, item)
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

func (db *DB) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List all grocery items")

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := db.getCollectionFromHeader(authHeader, GROCERY)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	items, err := db.Repo.FetchAll(r.Context(), groceryCollection)
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

func (db *DB) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an item by ID")

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := db.getCollectionFromHeader(authHeader, GROCERY)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Println("failed to convert id to integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.Repo.DeleteByID(r.Context(), groceryCollection, id)
	if err != nil {
		fmt.Println("failed to delete:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (db *DB) SetActiveByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Change active state")

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := db.getCollectionFromHeader(authHeader, GROCERY)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println("failed to convert id to integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.Repo.ToggleActiveByID(r.Context(), groceryCollection, id)
	if err != nil {
		log.Println("failed to toggle active state")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (db *DB) UpdateByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Update by ID")

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := db.getCollectionFromHeader(authHeader, GROCERY)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

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

	err = db.Repo.UpdateItemByID(r.Context(), groceryCollection, body.ItemID, new_values)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (db *DB) MoveToFridge(w http.ResponseWriter, r *http.Request) {
	log.Println("Move items to fridge")

	authHeader := r.Header.Get("Authorization")
	var err error
	if authHeader == "" {
		err = db.Repo.MoveToFridge(r.Context(), nil)
		if err != nil {
			log.Printf("failed to move grocery items to fridge: %v", err)
		}
	} else {
		userClaims, err := getUserClaimsFromHeader(authHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userDocRef := db.Repo.GetDocRef(db.Repo.GetCollectionRef(USER, nil), userClaims.Username)
		log.Println("THIS IS THE USER:", userDocRef.ID)

		err = db.Repo.MoveToFridge(r.Context(), userDocRef)
		if err != nil {
			log.Printf("failed to move grocery items to fridge: %v", err)
		}
	}

}

func (db *DB) RearrageItems(w http.ResponseWriter, r *http.Request) {
	log.Println("Rearrage items")

	authHeader := r.Header.Get("Authorization")
	groceryCollection, err := db.getCollectionFromHeader(authHeader, GROCERY)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

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

	err = db.Repo.RearrageItems(r.Context(), groceryCollection, body.OldIndex, body.NewIndex)
	if err != nil {
		log.Printf("failed to rearrage items: %v", err)
	}
}
