package url

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Gurveer1510/urlshortner/internal/domain"
	"github.com/Gurveer1510/urlshortner/internal/shortid"
)

type Usecase struct {
    UrlStore domain.UrlStore
}

func NewUseCase(urlStore domain.UrlStore) *Usecase {
    return &Usecase{UrlStore: urlStore}
}

func (uc *Usecase) Shorten(ctx context.Context, userId string, urlReq domain.UrlReq) (string, error) {
	if urlReq.ExpiresAt != nil && urlReq.ExpiresAt.UTC().Before(time.Now()) {
		return "", errors.New("expires_at must be in the future")
	}
	if urlReq.Code != "" {
		err := uc.UrlStore.Save(ctx, userId, urlReq)
		if errors.Is(err, domain.ErrConflict) {
			return "", fmt.Errorf("This code is already in use")
		}
		return urlReq.Code, nil
	}
	var newUrl domain.UrlReq
	newUrl.Url = urlReq.Url
	newUrl.ExpiresAt = urlReq.ExpiresAt

	for range 5 {
		code, err := shortid.Generate()
		newUrl.Code = code
		if err != nil {
			return "", err
		}
		err = uc.UrlStore.Save(ctx, userId, newUrl)
		if err == nil {
			return code, err
		}

		if !errors.Is(err, domain.ErrNotFound) {
			return "", err
		}
	}
	return "", errors.New("failed to genrate unique code after retries")
}

func (uc *Usecase) Get(ctx context.Context, ipAddress, code string) (string, error) {
	shortUrl, err := uc.UrlStore.Get(ctx, code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrNotFound
		}
		return "", err
	}

	if shortUrl.ExpiresAt != nil {
		if shortUrl.ExpiresAt.Before(time.Now()) {
			return "", domain.ErrExpiredCode
		}
	}

	err = uc.UrlStore.SaveClick(ctx, ipAddress, code)
	if err != nil {
		log.Printf("Error in usecase from SaveClick(): %w", err)
	}

	return shortUrl.Url, nil
}

func (u *Usecase) GetStats(ctx context.Context, code string) (*domain.StatsResp, error) {
	return u.UrlStore.GetStats(ctx, code)
}
