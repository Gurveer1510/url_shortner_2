package main

import (
	"log"
	"net/http"

	"github.com/Gurveer1510/urlshortner/internal/handlers"
	"github.com/Gurveer1510/urlshortner/internal/store"
	"github.com/Gurveer1510/urlshortner/internal/usecase"
)

func main() {
	s := store.NewInMemory()
	uc := usecase.NewUseCase(s)
	h := handlers.NewHandler(uc, "http://localhost:8080")

	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", h.Shorten)
	mux.HandleFunc("GET /{code}", h.Redirect)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}