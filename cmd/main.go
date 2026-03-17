package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Gurveer1510/urlshortner/internal/config"
	"github.com/Gurveer1510/urlshortner/internal/db"
	"github.com/Gurveer1510/urlshortner/internal/handlers"
	"github.com/Gurveer1510/urlshortner/internal/middlewares"
	"github.com/Gurveer1510/urlshortner/internal/persistence"
	"github.com/Gurveer1510/urlshortner/internal/usecase"
)

func main() {
	conf, err := config.LoadConfig()

	if err != nil {
		log.Println("ERROR in loading config file:", err.Error())
		return
	}
	dsn := db.DSN(conf)
	pool, err := db.NewPool(context.Background(), dsn)
	if err != nil {
		log.Println("ERROR in creating pool:", err.Error())
		return
	}
	p := persistence.NewPersistence(pool)
	// s := store.NewInMemory()
	uc := usecase.NewUseCase(p)
	h := handlers.NewHandler(uc, "http://localhost:8080")

	limiter := middlewares.NewRateLimiter(5,10)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten",  h.Shorten)
	mux.HandleFunc("GET /{code}", h.Redirect)
	mux.HandleFunc("GET /stats/{code}", h.GetStats)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", limiter.Middleware(mux)))
}
