package application

import (
	"net/http"

	"github.com/NathanRJohnson/live-backend/wtfinance/handler"
	"github.com/NathanRJohnson/live-backend/wtfinance/repository/transaction"
)

func (a *App) loadRoutes() {
	router := http.NewServeMux()

	transactionRouter := http.NewServeMux()
	a.loadTransactionRoutes(transactionRouter)

	router.Handle("/transaction/", http.StripPrefix("/transaction", transactionRouter))
	a.router = router
}

func (a *App) loadTransactionRoutes(router *http.ServeMux) {
	transactionHandler := &handler.Transaction{
		Repo: &transaction.GoogleSheetsRepo{
			Service: a.gss,
		},
	}
	router.HandleFunc("POST /", transactionHandler.Create)
}
