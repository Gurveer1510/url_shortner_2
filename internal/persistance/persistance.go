package persistance

import (
	"context"
	"errors"

	"github.com/Gurveer1510/urlshortner/internal/db"
	"github.com/jackc/pgx/v5"
)

type Persistance struct {
	db *db.DB
}

func NewPersistance(db *db.DB) *Persistance {
	return &Persistance{db: db}
}

func (p *Persistance) Save(code, url string) error {
	query := `
		INSERT INTO links (code, url) VALUES ($1, $2) ON CONFLICT (code) DO NOTHING
	`
	_, err := p.db.Pool.Exec(context.Background(), query, code, url)
	return err
}

func (p *Persistance) Get(code string) (string, error) {
	var url string
	query := `
		SELECT url from links WHERE code=$1
	`
	err := p.db.Pool.QueryRow(context.Background(), query, code).Scan(&url)

	if errors.Is(err, pgx.ErrNoRows) {
		return  "", pgx.ErrNoRows
	}

	return url, nil

}
