package main

import (
	"log"
	"net/http"

	"github.com/JulioB1t/job-tracker-backend/internal/config"
	"github.com/JulioB1t/job-tracker-backend/internal/httpapi"
	"github.com/JulioB1t/job-tracker-backend/internal/store"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	cfg := config.Load()
	applicationStore := store.NewMemoryApplicationStore()
	handler := httpapi.NewHandler(applicationStore)
	router := httpapi.NewRouter(handler)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	log.Printf("server running addr=%s", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
