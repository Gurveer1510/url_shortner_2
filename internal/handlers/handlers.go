package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

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
	var body struct {
		Url  string `json:"url"`
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Url == "" {
		http.Error(rw, "invalid request", http.StatusBadRequest)
		return
	}

	if body.Code == "" {
		code, err := h.usecase.Shorten(body.Url, "")
		if err != nil {
			fmt.Println(err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(rw).Encode(map[string]string{
			"short_url": h.baseUrl + "/" + code,
		})
	} else {
		code, err := h.usecase.Shorten(body.Url, body.Code)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		json.NewEncoder(rw).Encode(map[string]string{
			"short_url": h.baseUrl + "/" + code,
		})
	}

}

func (h *Handlers) Redirect(rw http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	ip := GetIP(r)
	url, err := h.usecase.Get(ip, code)
	if err != nil {
		http.Error(rw, "not found", http.StatusNotFound)
		return
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
