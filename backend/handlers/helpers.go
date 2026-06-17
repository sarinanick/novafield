package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
	"sync"
	"time"
)

type H = map[string]interface{}

func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, H{"error": msg})
}

func Decode(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func GetUser(r *http.Request) *models.User {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return store.GetUserByToken(token)
}

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	window   time.Duration
	max      int
}

var LoginLimiter = &rateLimiter{
	attempts: make(map[string][]time.Time),
	window:   1 * time.Minute,
	max:      10,
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	var valid []time.Time
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.attempts[key] = valid

	if len(valid) >= rl.max {
		return false
	}
	rl.attempts[key] = append(rl.attempts[key], now)
	return true
}

func RateLimitMiddleware(rl *rateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Forwarded-For")
		if key == "" {
			key = r.Header.Get("X-Real-IP")
		}
		if key == "" {
			key = r.RemoteAddr
		}
		if !rl.allow(key) {
			Error(w, 429, "Too many requests, please try again later")
			return
		}
		next(w, r)
	}
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func normalizeSlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
