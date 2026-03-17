package apptypes

import "time"

type UrlReq struct {
	Url       string     `json:"url"`
	Code      string     `json:"code"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type StatsResp struct {
	Code      string     `json:"code"`
	Url       string     `json:"url"`
	Clicks    int        `json:"clicks"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	IsExpired bool       `json:"is_expired"`
	Data      []Click    `json:"data"`
}



type Click struct {
	IpAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}
