package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(handler *Handler) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/api/v1", func(router chi.Router) {
		router.Post("/applications", handler.createApplication)
		router.Get("/applications", handler.listApplications)
		router.Get("/applications/{id}", handler.getApplication)
		router.Post("/applications/{id}/status-transitions", handler.createStatusTransition)
		router.Get("/applications/{id}/status-transitions", handler.listStatusTransitions)
	})

	return router
}
