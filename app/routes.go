package application

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/thiagobgarc/orders-api/handler"
)

func loadRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/orders", loadOrderRoutes)

	return router
}

func loadOrderRoutes(router chi.Router) {
	orderhandler := &handler.Order{}

	router.Post("/", orderhandler.Create)
	router.Get("/", orderhandler.List)
	router.Get("/{id}", orderhandler.GetByID)
	router.Put("/{id}", orderhandler.UpdateByID)
	router.Delete("/{id}", orderhandler.DeleteByID)
}
