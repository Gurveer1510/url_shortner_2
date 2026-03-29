package user

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Gurveer1510/urlshortner/internal/auth/session"
	"github.com/Gurveer1510/urlshortner/internal/domain"
	userusecase "github.com/Gurveer1510/urlshortner/internal/usecase/user"
)

type UserHandler struct {
	sessionStore session.Store
	userUsecase  *userusecase.UserUsecase
}

func NewUserHandler(sessionStore session.Store, userUsecase *userusecase.UserUsecase) *UserHandler {
	return &UserHandler{
		sessionStore: sessionStore,
		userUsecase:  userUsecase,
	}
}

func (uh *UserHandler) CreateUserHandler(rw http.ResponseWriter, r *http.Request) {
	var body domain.User

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		json.NewEncoder(rw).Encode(map[string]string{
			"error": "invalid json",
		})
		return
	}

	id, err := uh.userUsecase.CreateUser(r.Context(), body)
	if err != nil {
		log.Println("ERROR WHILE CREATING USER:", err)
		http.Error(rw, "Something went wrong", http.StatusInternalServerError)
		return
	}

	cookie, err := uh.sessionStore.Create(id, body.Email)
	if err != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Println(err)
			json.NewEncoder(rw).Encode(map[string]string{
				"error": "something went wrong",
			})
			return
		}
	}
	c := http.Cookie{
		Name:     "cookie",
		Value:    cookie.Id,
		Expires:  cookie.ExpiresAt,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	userId := http.Cookie{
		Name:  "user-id",
		Value: id,
	}
	http.SetCookie(rw, &c)
	http.SetCookie(rw, &userId)
}

func (uh *UserHandler) LoginHandler(rw http.ResponseWriter, r *http.Request) {
	var body domain.LoginReq

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		json.NewEncoder(rw).Encode(map[string]string{
			"error": "invalid json",
		})
		return
	}

	id, err := uh.userUsecase.VerifyPassword(r.Context(), body)
	if err != nil {
		log.Println("here cuz of err")
		log.Println("ERROR WHILE VERIFYING PASSWORD:", err)
		json.NewEncoder(rw).Encode(map[string]string{
			"error": "Incorrect credentials",
		})
		return
	}

	cookie, err := uh.sessionStore.Create(id, body.Email)
	log.Println(cookie)
	if err != nil {
		json.NewEncoder(rw).Encode(map[string]string{
			"error": "something went wrong",
		})
		return
	}

	c := http.Cookie{
		Name:     "cookie",
		Value:    cookie.Id,
		Expires:  cookie.ExpiresAt,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	userId := http.Cookie{
		Name:  "user-id",
		Value: id,
	}
	http.SetCookie(rw, &c)
	http.SetCookie(rw, &userId)
}
