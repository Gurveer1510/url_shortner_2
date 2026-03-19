package postgres

import (
    "context"

    "github.com/Gurveer1510/urlshortner/internal/db"
    "github.com/Gurveer1510/urlshortner/internal/domain"
)

type UserRepository struct {
    db *db.DB
}

func NewUserRepository(db *db.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, u domain.User) (string, error) {
    query := `
        INSERT INTO users (name, email, hashpass) VALUES ($1, $2, $3) RETURNING id
    `
    var id string
    err := r.db.Pool.QueryRow(ctx, query, u.Name, u.Email, u.Password).Scan(&id)
    if err != nil {
        return "", err
    }
    return id, nil
}

func (r *UserRepository) GetUser(ctx context.Context, email string) (*domain.User, error) {
    query := `
        SELECT id, name, email, hashpass FROM users WHERE email = $1
    `
    var user domain.User
    err := r.db.Pool.QueryRow(ctx, query, email).Scan(&user.Id, &user.Name, &user.Email, &user.Password)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

