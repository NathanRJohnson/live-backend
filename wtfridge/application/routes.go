package application

import (
	"net/http"

	handler "github.com/NathanRJohnson/live-backend/wtfridge/handler"
	"github.com/NathanRJohnson/live-backend/wtfridge/repository/item"
)

func (a *App) loadRoutes() {
	router := http.NewServeMux()

	fridgeRouter := http.NewServeMux()
	a.loadFridgeRoutes(fridgeRouter)

	router.Handle("/fridge/", http.StripPrefix("/fridge", fridgeRouter))

	a.router = router
}

func (a *App) loadFridgeRoutes(router *http.ServeMux) {
	fridgeHandler := &handler.Item{
		Repo: &item.FirebaseRepo{
			Client: a.fdb,
		},
	}
	router.HandleFunc("POST /", fridgeHandler.Create)
	router.HandleFunc("GET /", fridgeHandler.List)
	// router.HandleFunc("GET /{id}", fridgeHandler.GetByID)
	// router.HandleFunc("PUT /{id}", fridgeHandler.UpdateByID)
	router.HandleFunc("DELETE /{id}", fridgeHandler.DeleteByID)
}
