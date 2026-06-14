package httpapi

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

//go:embed docs/openapi.yaml
var openAPISpec []byte

//go:embed docs/swagger.html
var swaggerHTML []byte

//go:embed docs/favicon.svg
var swaggerFavicon []byte

func NewRouter(handler *Handler) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(logRequestBody)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://192.168.1.167:*",
			"https://192.168.1.167:*",
			"http://localhost:*",
			"https://localhost:*",
			"http://127.0.0.1:*",
			"https://127.0.0.1:*",
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		MaxAge:         300,
	}))

	router.Get("/openapi.yaml", serveOpenAPISpec)
	router.Get("/swagger", redirectToSwagger)
	router.Get("/swagger/", serveSwaggerUI)
	router.Get("/swagger/favicon.svg", serveSwaggerFavicon)

	router.Route("/api/v1", func(router chi.Router) {
		router.Post("/applications", handler.createApplication)
		router.Get("/applications", handler.listApplications)
		router.Get("/applications/{id}", handler.getApplication)
		router.Put("/applications/{id}", handler.updateApplication)
		router.Post("/applications/{id}/status-transitions", handler.createStatusTransition)
		router.Get("/applications/{id}/status-transitions", handler.listStatusTransitions)
	})

	return router
}

func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpec)
}

func redirectToSwagger(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
}

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(swaggerHTML)
}

func serveSwaggerFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(swaggerFavicon)
}
