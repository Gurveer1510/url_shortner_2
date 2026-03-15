package usecase

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Gurveer1510/urlshortner/internal/store"
	urltype "github.com/Gurveer1510/urlshortner/internal/urlType"
	"github.com/Gurveer1510/urlshortner/internal/utils"
)

type Usecase struct {
	UrlStore store.Store
}

func NewUseCase(urlStore store.Store) *Usecase {
	return &Usecase{UrlStore: urlStore}
}

func (uc *Usecase) Shorten(urlReq urltype.UrlReq) (string, error) {
	if urlReq.Code != "" {
		err := uc.UrlStore.Save(urlReq)
		if errors.Is(err, store.ErrConflict) {
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
		err = uc.UrlStore.Save(newUrl)
		if err == nil {
			return code, err
		}

		if !errors.Is(err, store.ErrConflict) {
			return "", err
		}
	}
	return "", errors.New("failed to genrate unique code after retries")
}

func (uc *Usecase) Get(ipAddress, code string) (string, error) {
	shortUrl, err := uc.UrlStore.Get(code)
	if err != nil && errors.Is(err, store.ErrNotFound) {
		log.Println(err)
		return "", store.ErrNotFound
	}

	if shortUrl.ExpiresAt != nil {
		if shortUrl.ExpiresAt.Before(time.Now()) {
			return "", store.ErrExpiredCode
		}
	}

	err = uc.UrlStore.SaveClick(ipAddress, code)
	if err != nil {
		log.Println("Error in usecase from SaveClick(): %w", err)
	}

	return shortUrl.Url, nil
}
