package user

import (
	"context"
	"log"

	"github.com/Gurveer1510/urlshortner/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	Repo domain.UserStore
}

func NewUserUsecase(repo domain.UserStore) *UserUsecase  {
	return &UserUsecase{
		Repo: repo,
	}
}

func (u *UserUsecase) CreateUser(ctx context.Context, user domain.User) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost) 
	if err != nil {
		return "", err
	}
	user.Password = string(bytes)
	return u.Repo.CreateUser(ctx, user)
}

func (u *UserUsecase) VerifyPassword(ctx context.Context, user domain.LoginReq) (string, error) {
	userBody, err := u.Repo.GetUser(ctx, user.Email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(userBody.Password), []byte(user.Password))
	if err != nil {
		log.Println("HASH:", userBody.Password)
		log.Println("PASS:", user.Password)
		log.Println(err)
		return "", err
	}
	return userBody.Id, nil
}