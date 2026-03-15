package persistance

import (
	"context"
	"errors"
	"fmt"

	"github.com/Gurveer1510/urlshortner/internal/db"
	"github.com/Gurveer1510/urlshortner/internal/store"
	urltype "github.com/Gurveer1510/urlshortner/internal/urlType"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Persistance struct {
	db *db.DB
}

func NewPersistance(db *db.DB) *Persistance {
	return &Persistance{db: db}
}

func (p *Persistance) Save(urlReq urltype.UrlReq) error {
	// log.Println("IN PERSISTENC:", urlReq)
	query := `
		INSERT INTO links (code, url, expires_at) VALUES ($1, $2, $3) 
	`
	_, err := p.db.Pool.Exec(context.Background(), query, urlReq.Code, urlReq.Url, urlReq.ExpiresAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return store.ErrConflict
			}
		}
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (p *Persistance) Get(code string) (*urltype.UrlReq, error) {
	var shortUrl urltype.UrlReq
	query := `
		UPDATE links SET clicks=clicks+1 WHERE code=$1 RETURNING url, code, expires_at
	`
	err := p.db.Pool.QueryRow(context.Background(), query, code).Scan(&shortUrl.Url, &shortUrl.Code, &shortUrl.ExpiresAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, fmt.Errorf("database error: %w", err)
		}

		return nil, fmt.Errorf("database error: %w", err)
	}
	return &shortUrl, nil
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
