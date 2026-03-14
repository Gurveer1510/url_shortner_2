package usecase

import (
	"github.com/Gurveer1510/urlshortner/internal/store"
	"github.com/Gurveer1510/urlshortner/internal/utils"
)

type Usecase struct {
	UrlStore	store.Store
}

func NewUseCase(urlStore store.Store) *Usecase {
	return &Usecase{UrlStore: urlStore}
}

func (uc *Usecase) Shorten(url string) (string, error) { 
	code, err := utils.Generate(url)
	if err != nil {
		return "", err
	}
	err = uc.UrlStore.Save(code, url)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (uc *Usecase) Get(code string)( string, error) {
	url, err := uc.UrlStore.Get(code)
	if err != nil {
		return "", err
	}

	return url, nil
}