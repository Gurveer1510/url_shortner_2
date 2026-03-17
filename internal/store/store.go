package store

import (
	"context"

	urltype "github.com/Gurveer1510/urlshortner/internal/apptypes"
)

type Store interface {
	Save(ctx context.Context, urlReq urltype.UrlReq) error
	Get(ctx context.Context, code string) (*urltype.UrlReq, error)
	SaveClick(ctx context.Context, ipAddress, code string) error
	GetStats(ctx context.Context, code string) (*urltype.StatsResp, error)
}
