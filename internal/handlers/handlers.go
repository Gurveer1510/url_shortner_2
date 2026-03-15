package handlers

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Gurveer1510/urlshortner/internal/store"
	urltype "github.com/Gurveer1510/urlshortner/internal/urlType"
	"github.com/Gurveer1510/urlshortner/internal/usecase"
)

type Handlers struct {
	usecase *usecase.Usecase
	baseUrl string
}

func NewHandler(uc *usecase.Usecase, baseUrl string) *Handlers {
	return &Handlers{
		usecase: uc,
		baseUrl: baseUrl,
	}
}

func (h *Handlers) Shorten(rw http.ResponseWriter, r *http.Request) {
	var body urltype.UrlReq

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Url == "" {
		http.Error(rw, "invalid request", http.StatusBadRequest)
		return
	}
	// log.Println(body)
	_, err := url.ParseRequestURI(body.Url)
	if err != nil {
		json.NewEncoder(rw).Encode(map[string]string{
			"error": "invalid URL string",
		})
	}

	code, err := h.usecase.Shorten(body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(rw).Encode(map[string]string{
		"short_url": h.baseUrl + "/" + code,
	})
}

func (h *Handlers) Redirect(rw http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	ip := GetIP(r)
	url, err := h.usecase.Get(ip, code)
	if err != nil  {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, store.ErrExpiredCode){
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
	}
	http.Redirect(rw, r, url, http.StatusFound)
}

func GetIP(r *http.Request) string {

	ip := r.Header.Get("X-Forwarded-For")

	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}

	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return ip
}
