package persistance

import (
	"context"
	"errors"
	"fmt"

	"github.com/Gurveer1510/urlshortner/internal/db"
	"github.com/Gurveer1510/urlshortner/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Persistance struct {
	db *db.DB
}

func NewPersistance(db *db.DB) *Persistance {
	return &Persistance{db: db}
}

func (p *Persistance) Save(code, url string) error {
	query := `
		INSERT INTO links (code, url) VALUES ($1, $2) 
	`	
	_, err := p.db.Pool.Exec(context.Background(), query, code, url)

	if err != nil {
		var pgErr *pgconn.PgError
		// Use errors.As to check the underlying error type
		if errors.As(err, &pgErr) {
			// Check for the specific PostgreSQL error code for unique_violation
			if pgErr.Code == "23505" {
				return  store.ErrConflict
			}
		}
		// Handle other types of errors or re-wrap the original error
		return  fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (p *Persistance) Get(code string) (string, error) {
	var url string
	
	query := `
		UPDATE links SET clicks=clicks+1 WHERE code=$1 RETURNING url
	`
	err := p.db.Pool.QueryRow(context.Background(), query, code).Scan(&url)

	if err != nil {
		var pgErr *pgconn.PgError
		// Use errors.As to check the underlying error type
		if errors.As(err, &pgErr) {
			// Check for the specific PostgreSQL error code for unique_violation
			if errors.Is(err, pgx.ErrNoRows) {
				return "", store.ErrNotFound
			}
		}
		// Handle other types of errors or re-wrap the original error
		return "", fmt.Errorf("database error: %w", err)
	}
	return url, nil
}

func (p *Persistance) SaveClick(ipAddress, code string) error {

	query := `
		INSERT INTO url_clicks (code, ip_address) VALUES ($1, $2)
	`
	_, err := p.db.Pool.Exec(context.Background(), query, code, ipAddress)
	if err != nil {
		return err
	}

	return nil 
}
