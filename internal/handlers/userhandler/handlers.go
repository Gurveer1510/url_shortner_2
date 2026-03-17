package user

import (
	"github.com/Gurveer1510/urlshortner/internal/api/session"
)

type UserHandler struct {
	sessionStore *session.SessionStore
	// userUsecase
}
