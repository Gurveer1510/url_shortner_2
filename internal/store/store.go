package store

import (
	"context"
	"errors"

	urltype "github.com/Gurveer1510/urlshortner/internal/urlType"
)

var ErrNotFound = errors.New("short code not found")
var ErrConflict = errors.New("Duplicate code found")
var ErrExpiredCode = errors.New("Code is expired")

type Store interface {
	Save(ctx context.Context, urlReq urltype.UrlReq) error
	Get(ctx context.Context, code string) (*urltype.UrlReq, error)
	SaveClick(ctx context.Context, ipAddress, code string) error
}