package handler

import (
	"encoding/json"
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
	log.Println("Create transaction")

	sheetref := r.Header.Get("SheetRef")
	if sheetref == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error: required header SheetRef not present.")
		return
	}

	var body struct {
		DateCreated *time.Time `json:"date"`
		Name        string     `json:"name"`
		Amount      float32    `json:"amount"`
		Category    string     `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error unmarshaling requst:", err)
		return
	}

	if body.Name == "" || body.Category == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("transaction is missing required fields")
		return
	}

	if body.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("transaction value must be positive, non-zero")
		return
	}

	transaction := model.Transaction{
		Name:        body.Name,
		Amount:      body.Amount,
		Category:    body.Category,
		DateCreated: body.DateCreated,
	}

	err := t.Repo.Insert(r.Context(), transaction, sheetref)
	if err != nil {
		log.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(transaction)
	if err != nil {
		log.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

func (t *Transaction) History(w http.ResponseWriter, r *http.Request) {
	log.Println("Transaction history this cycle")

	sheetRef := r.Header.Get("SheetRef")
	if sheetRef == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error: required header SheetRef not present.")
		return
	}

	// decrypt the value here

	transactions, err := t.Repo.FetchTransactions(r.Context(), sheetRef)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error fetching transactions: %v", err)
		return
	}

	res, err := json.Marshal(transactions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error marshalling JSON: %v", err)
		return
	}

	w.Write(res)
}

func (t *Transaction) CircleValues(w http.ResponseWriter, r *http.Request) {
	sheetRef := r.Header.Get("SheetRef")
	if sheetRef == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error: required header SheetRef not present.")
		return
	}

	circleValues, err := t.Repo.FetchCircleAmounts(r.Context(), sheetRef)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error fetching transactions: %v", err)
		return
	}

	res, err := json.Marshal(circleValues)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error marshalling JSON: %v", err)
		return
	}

	w.Write(res)

}
