package auth

import (
	"context"
	"log"
	"net/http"

	"github.com/Gurveer1510/urlshortner/internal/auth/session"
)

type Auth struct {
	store session.Store
}

func NewAuth(s session.Store) *Auth {
	return &Auth{
		store: s,
	}
}

func (a *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("cookie")
		// userCookie, err := r.Cookie("user-id")
		// log.Println(cookie.Value)
		if err != nil {

			log.Println("IN auth 1")
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}
		log.Println("COOKKIE", cookie)
		session, ok := a.store.Get(cookie.Value)
		log.Println(session)
		if !ok {
			log.Println("IN auth 2")
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}
