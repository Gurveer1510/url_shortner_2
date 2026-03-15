package urltype

import "time"

type UrlReq struct {
	Url       string    `json:"url"`
	Code      string    `json:"code"`
	ExpiresAt *time.Time `json:"expires_at"`
}
