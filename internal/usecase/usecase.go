package usecase

import (
	"errors"
	"fmt"

	"github.com/Gurveer1510/urlshortner/internal/store"
	"github.com/Gurveer1510/urlshortner/internal/utils"
)

type Usecase struct {
	UrlStore store.Store
}

func NewUseCase(urlStore store.Store) *Usecase {
	return &Usecase{UrlStore: urlStore}
}

func (uc *Usecase) Shorten(url string, code string ) (string, error) {
	if code != "" {
		err := uc.UrlStore.Save(code, url)
		if errors.Is(err, store.ErrConflict) {
			return "", fmt.Errorf("This code is already in use")
		}
		return code, nil
	}
	for range 5 {
		code, err := utils.Generate(url)
		if err != nil {
			return "", err
		}
		err = uc.UrlStore.Save(code, url)
		if err == nil {
			return code, err
		}
		
		if !errors.Is(err, store.ErrConflict) {
			return "", err
		}
	}
	return "", errors.New("failed to genrate unique code after retries")
}

func (uc *Usecase) Get(code string) (string, error) {
	url, err := uc.UrlStore.Get(code)
	if err != nil {
		return "", err
	}

	return url, nil
}
