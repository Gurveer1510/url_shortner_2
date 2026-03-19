package domain

import "context"

type UserStore interface {
    CreateUser(ctx context.Context, u User) (string, error)
    GetUser(ctx context.Context, email string) (*User, error)
}

