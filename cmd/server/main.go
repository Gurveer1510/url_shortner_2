package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Gurveer1510/urlshortner/internal/auth/session"
	"github.com/Gurveer1510/urlshortner/internal/config"
	"github.com/Gurveer1510/urlshortner/internal/db"
	urlhandler "github.com/Gurveer1510/urlshortner/internal/http/handlers/url"
	"github.com/Gurveer1510/urlshortner/internal/http/handlers/user"
	"github.com/Gurveer1510/urlshortner/internal/http/middleware"
	"github.com/Gurveer1510/urlshortner/internal/http/middleware/auth"
	"github.com/Gurveer1510/urlshortner/internal/repository/postgres"
	usecase "github.com/Gurveer1510/urlshortner/internal/usecase/url"
	userusecase "github.com/Gurveer1510/urlshortner/internal/usecase/user"
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

	sessStore := session.NewSessionStore()
	authMiddleware := auth.NewAuth(sessStore)

	urlRepo := postgres.NewURLRepository(pool)
	userRepo := postgres.NewUserRepository(pool)

	uc := usecase.NewUseCase(urlRepo)
	userUsecase := userusecase.NewUserUsecase(userRepo)

	baseURL := conf.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	
	h := urlhandler.NewHandler(uc, baseURL)
	uh := user.NewUserHandler(sessStore, userUsecase)

	limiter := middleware.NewRateLimiter(5, 10)

	mux := http.NewServeMux()
	mux.Handle("POST /shorten", authMiddleware.AuthMiddleware(http.HandlerFunc(h.Shorten)))
	mux.HandleFunc("GET /{code}", h.Redirect)
	mux.HandleFunc("GET /stats/{code}", h.GetStats)
	mux.HandleFunc("POST /signup", uh.CreateUserHandler)
	mux.HandleFunc("POST /login", uh.LoginHandler)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", limiter.Middleware(mux)))
}
