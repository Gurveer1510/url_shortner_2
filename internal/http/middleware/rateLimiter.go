package middleware

import (
	"net/http"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens > 1 {
		tb.tokens--
		return true
	}
	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapased := now.Sub(tb.lastRefill).Seconds()

	tb.tokens = min(tb.capacity, tb.tokens+elapased*tb.refillRate)
	tb.lastRefill = now
}

func min(x, y float64) float64 {
	if x < y {
		return x
	}
	return y
}

type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.Mutex
	rate    float64
	cap     float64
}

func NewRateLimiter(rate, capacity float64) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
		rate: rate,
		cap: capacity,
	}
}

func (rl *RateLimiter) getBucket(ip string) *TokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if _, ok := rl.buckets[ip]; !ok {
		rl.buckets[ip] = NewTokenBucket(rl.cap, rl.rate)
	}

	return rl.buckets[ip]
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) { 
		ip := r.RemoteAddr
		bucket := rl.getBucket(ip)

		if !bucket.Allow() {
			rw.Header().Set("Retry-After", "1")
			http.Error(rw, "Rate limit exeeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(rw, r)
	})
	
}
