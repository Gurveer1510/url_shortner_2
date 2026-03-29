package domain

import "context"

type UrlStore interface {
    Save(ctx context.Context, userId string, urlReq UrlReq) error
    Get(ctx context.Context, code string) (*Link, error)
    SaveClick(ctx context.Context, ipAddress, code string) error
    GetStats(ctx context.Context, code string) (*StatsResp, error)
}

