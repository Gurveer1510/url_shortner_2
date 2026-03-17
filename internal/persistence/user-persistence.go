package persistence

import (
	"context"

	"github.com/Gurveer1510/urlshortner/internal/apptypes"
)

func (p *Persistence) CreateUser(ctx context.Context, u apptypes.User) error {
	query := `
		INSERT INTO users (name, email, password) VALUES ($1, $2, $3)
	`
	_, err := p.db.Pool.Exec(ctx, query, u.Name, u.Email, u.Password)
	if err != nil {
		return err
	}
	return nil
}

func (p *Persistence) GetUser(ctx context.Context, email string) (*apptypes.User, error) {
	query := `
		SELECT name, email, password FROM users WHERE email = $1
	`
	var user apptypes.User
	err := p.db.Pool.QueryRow(ctx, query, email).Scan(&user.Name, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}