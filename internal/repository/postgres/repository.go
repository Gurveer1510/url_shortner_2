package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Gurveer1510/urlshortner/internal/db"
	"github.com/Gurveer1510/urlshortner/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type URLRepository struct {
    db *db.DB
}

func NewURLRepository(db *db.DB) *URLRepository {
    return &URLRepository{db: db}
}

func (p *URLRepository) Save(ctx context.Context, userId string, urlReq domain.UrlReq) error {
	query := `
		INSERT INTO links (code, url, expires_at, user_id) VALUES ($1, $2, $3, $4) 
	`
	_, err := p.db.Pool.Exec(ctx, query, urlReq.Code, urlReq.Url, urlReq.ExpiresAt, userId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return domain.ErrConflict
			}
		}
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (p *URLRepository) Get(ctx context.Context, code string) (*domain.Link, error) {
	log.Println("GOT HERE INSTEAD OF REDIS")
	var shortUrl domain.Link
	query := `
		SELECT url, code, expires_at FROM links WHERE code = $1
	`
	err := p.db.Pool.QueryRow(ctx, query, code).Scan(&shortUrl.Url, &shortUrl.Code, &shortUrl.ExpiresAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, fmt.Errorf("database error: %w", err)
		}

		return nil, fmt.Errorf("database error: %w", err)
	}
	return &shortUrl, nil
}

func (p *URLRepository) SaveClick(ctx context.Context, ipAddress, code string) error {
	query := `
		INSERT INTO url_clicks (code, ip_address) VALUES ($1, $2)
	`
	_, err := p.db.Pool.Exec(ctx, query, code, ipAddress)
	if err != nil {
		return err
	}
	bumpCount := `
		UPDATE links SET clicks=clicks+1 WHERE code=$1
	`
	_, err = p.db.Pool.Exec(ctx, bumpCount, code)
	if err != nil {
		return err
	}
	return nil
}

func (p *URLRepository) GetStats(ctx context.Context, code string) (*domain.StatsResp, error) {
	linkQuery := `SELECT code, url, clicks, created_at, expires_at FROM links WHERE code = $1`
	var shortUrl domain.StatsResp
	err := p.db.Pool.QueryRow(ctx, linkQuery, code).Scan(&shortUrl.Code, &shortUrl.Url, &shortUrl.Clicks, &shortUrl.CreatedAt, &shortUrl.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	var isExpired bool
	if shortUrl.ExpiresAt != nil {
		isExpired = shortUrl.ExpiresAt.Before(time.Now())
	}
	shortUrl.IsExpired = isExpired

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
		click := domain.Click{IpAddress: ipAddr, CreatedAt: createdAt}
		shortUrl.Data = append(shortUrl.Data, click)
	}

	return &shortUrl, nil
}

// User persistence moved to user_repository.go
