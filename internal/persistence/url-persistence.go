package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Gurveer1510/urlshortner/internal/apperrors"
	"github.com/Gurveer1510/urlshortner/internal/db"
	urltype "github.com/Gurveer1510/urlshortner/internal/apptypes"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Persistence struct {
	db *db.DB
}

func NewPersistence(db *db.DB) *Persistence {
	return &Persistence{db: db}
}

func (p *Persistence) Save(ctx context.Context, urlReq urltype.UrlReq) error {
	// log.Println("IN PERSISTENC:", urlReq)
	query := `
		INSERT INTO links (code, url, expires_at) VALUES ($1, $2, $3) 
	`
	_, err := p.db.Pool.Exec(ctx, query, urlReq.Code, urlReq.Url, urlReq.ExpiresAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return apperrors.ErrConflict
			}
		}
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (p *Persistence) Get(ctx context.Context, code string) (*urltype.UrlReq, error) {
	var shortUrl urltype.UrlReq
	query := `
		UPDATE links SET clicks=clicks+1 WHERE code=$1 RETURNING url, code, expires_at
	`
	err := p.db.Pool.QueryRow(ctx, query, code).Scan(&shortUrl.Url, &shortUrl.Code, &shortUrl.ExpiresAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, fmt.Errorf("database error: %w", err)
		}

		return nil, fmt.Errorf("database error: %w", err)
	}
	return &shortUrl, nil
}

func (p *Persistence) SaveClick(ctx context.Context, ipAddress, code string) error {

	query := `
		INSERT INTO url_clicks (code, ip_address) VALUES ($1, $2)
	`
	_, err := p.db.Pool.Exec(ctx, query, code, ipAddress)
	if err != nil {
		return err
	}

	return nil
}

func (p *Persistence) GetStats(ctx context.Context, code string) (*urltype.StatsResp, error) {
	// First, get the link
	linkQuery := `SELECT code, url, clicks, created_at, expires_at FROM links WHERE code = $1`
	var shortUrl urltype.StatsResp
	err := p.db.Pool.QueryRow(ctx, linkQuery, code).Scan(&shortUrl.Code, &shortUrl.Url, &shortUrl.Clicks, &shortUrl.CreatedAt, &shortUrl.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	var isExpired bool
	if shortUrl.ExpiresAt != nil {
		isExpired = shortUrl.ExpiresAt.Before(time.Now())
	}
	shortUrl.IsExpired = isExpired

	// Then, get the clicks
	clicksQuery := `SELECT ip_address, created_at FROM url_clicks WHERE code = $1 ORDER BY created_at`
	rows, err := p.db.Pool.Query(ctx, clicksQuery, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ip sql.NullString
		var createdAt time.Time
		err := rows.Scan(&ip, &createdAt)
		if err != nil {
			return nil, err
		}
		ipAddr := ""
		if ip.Valid {
			ipAddr = ip.String
		}
		click := urltype.Click{IpAddress: ipAddr, CreatedAt: createdAt}
		shortUrl.Data = append(shortUrl.Data, click)
	}

	return &shortUrl, nil
}
