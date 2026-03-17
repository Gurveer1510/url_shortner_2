package auth

import (
	"context"
	"net/http"

	"github.com/Gurveer1510/urlshortner/internal/api/session"
)

type Auth struct {
	store *session.SessionStore
}

func NewAuth(s *session.SessionStore) *Auth {
	return &Auth{
		store: s,
	}
}

func (a *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request){
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		session, ok := a.store.Get(cookie.Value)
		if !ok {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return 
		}

		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}
