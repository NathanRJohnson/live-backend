package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NathanRJohnson/live-backend/wtfinance/model"
	"github.com/NathanRJohnson/live-backend/wtfinance/repository/transaction"
)

type Transaction struct {
	Repo *transaction.GoogleSheetsRepo
}

func (t *Transaction) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create transaction")
	var body struct {
		DateCreated *time.Time `json:"date"`
		Name        string     `json:"name"`
		Amount      float32    `json:"amount"`
		Category    string     `json:"category"`
		SheetRef    string     `json:"sheet"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("error unmarshaling requst:", err)
		return
	}

	// missing id
	if body.Name == "" || body.Category == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("item_name or item_is is missing or 0")
		return
	}

	transaction := model.Transaction{
		Name:        body.Name,
		Amount:      body.Amount,
		Category:    body.Category,
		DateCreated: body.DateCreated,
		SheetRef:    body.SheetRef,
	}

	err := t.Repo.Insert(r.Context(), transaction)
	if err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(transaction)
	if err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
	w.WriteHeader(http.StatusCreated)
}
