package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Gurveer1510/urlshortner/internal/apperrors"
	"github.com/Gurveer1510/urlshortner/internal/store"
	urltype "github.com/Gurveer1510/urlshortner/internal/apptypes"
	"github.com/Gurveer1510/urlshortner/internal/utils"
)

type Usecase struct {
	UrlStore store.Store
}

func NewUseCase(urlStore store.Store) *Usecase {
	return &Usecase{UrlStore: urlStore}
}

func (uc *Usecase) Shorten(ctx context.Context, urlReq urltype.UrlReq) (string, error) {
	if urlReq.ExpiresAt != nil && urlReq.ExpiresAt.UTC().Before(time.Now()) {
		return "", errors.New("expires_at must be in the future")
	}
	if urlReq.Code != "" {
		err := uc.UrlStore.Save(ctx, urlReq)
		if errors.Is(err, apperrors.ErrConflict) {
			return "", fmt.Errorf("This code is already in use")
		}
		return urlReq.Code, nil
	}
	var newUrl urltype.UrlReq
	newUrl.Url = urlReq.Url
	newUrl.ExpiresAt = urlReq.ExpiresAt

	for range 5 {
		code, err := utils.Generate()
		newUrl.Code = code
		if err != nil {
			return "", err
		}
		err = uc.UrlStore.Save(ctx, newUrl)
		if err == nil {
			return code, err
		}

		if !errors.Is(err, apperrors.ErrNotFound) {
			return "", err
		}
	}
	return "", errors.New("failed to genrate unique code after retries")
}

func (uc *Usecase) Get(ctx context.Context, ipAddress, code string) (string, error) {
	shortUrl, err := uc.UrlStore.Get(ctx, code)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return "", apperrors.ErrNotFound
		}
		return "", err
	}

	if shortUrl.ExpiresAt != nil {
		if shortUrl.ExpiresAt.Before(time.Now()) {
			return "", apperrors.ErrExpiredCode
		}
	}

	err = uc.UrlStore.SaveClick(ctx, ipAddress, code)
	if err != nil {
		log.Printf("Error in usecase from SaveClick(): %w", err)
	}

	return shortUrl.Url, nil
}

func (u *Usecase) GetStats(ctx context.Context, code string) (*urltype.StatsResp, error) {
	return u.UrlStore.GetStats(ctx, code)
}
