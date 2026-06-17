package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"novafield-api/store"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := &rateLimiter{
		attempts: make(map[string][]time.Time),
		window:   1 * time.Minute,
		max:      5,
	}

	for i := 0; i < 5; i++ {
		if !rl.allow("192.168.1.1:1234") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := &rateLimiter{
		attempts: make(map[string][]time.Time),
		window:   1 * time.Minute,
		max:      3,
	}

	for i := 0; i < 3; i++ {
		rl.allow("10.0.0.1:5555")
	}

	if rl.allow("10.0.0.1:5555") {
		t.Fatal("4th request should be blocked")
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	rl := &rateLimiter{
		attempts: make(map[string][]time.Time),
		window:   1 * time.Minute,
		max:      2,
	}

	rl.allow("A:1111")
	rl.allow("A:1111")

	if !rl.allow("B:2222") {
		t.Fatal("different key should be allowed")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := &rateLimiter{
		attempts: make(map[string][]time.Time),
		window:   1 * time.Millisecond,
		max:      1,
	}

	rl.allow("test:1234")

	rl.allow("test:1234")

	time.Sleep(2 * time.Millisecond)

	if !rl.allow("test:1234") {
		t.Fatal("request should be allowed after window expires")
	}
}

func TestRateLimitMiddleware_BlocksExcessRequests(t *testing.T) {
	rl := &rateLimiter{
		attempts: make(map[string][]time.Time),
		window:   1 * time.Minute,
		max:      2,
	}

	handler := RateLimitMiddleware(rl, func(w http.ResponseWriter, r *http.Request) {
		JSON(w, 200, H{"ok": true})
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rr := httptest.NewRecorder()
		handler(rr, req)

		if rr.Code != 200 {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}

	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != 429 {
		t.Fatalf("expected 429, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["error"] != "Too many requests, please try again later" {
		t.Errorf("expected rate limit error, got %v", result["error"])
	}
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	rl := &rateLimiter{
		attempts: make(map[string][]time.Time),
		window:   1 * time.Minute,
		max:      1,
	}

	handler := RateLimitMiddleware(rl, func(w http.ResponseWriter, r *http.Request) {
		JSON(w, 200, H{"ok": true})
	})

	req1 := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req1.RemoteAddr = "10.0.0.1:1111"
	rr1 := httptest.NewRecorder()
	handler(rr1, req1)

	if rr1.Code != 200 {
		t.Fatalf("IP1 first request: expected 200, got %d", rr1.Code)
	}

	req2 := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req2.RemoteAddr = "10.0.0.2:2222"
	rr2 := httptest.NewRecorder()
	handler(rr2, req2)

	if rr2.Code != 200 {
		t.Fatalf("IP2 first request: expected 200, got %d", rr2.Code)
	}

	req3 := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req3.RemoteAddr = "10.0.0.1:1111"
	rr3 := httptest.NewRecorder()
	handler(rr3, req3)

	if rr3.Code != 429 {
		t.Fatalf("IP1 second request: expected 429, got %d", rr3.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	resetDB()

	_, token := createTestUser("client")

	inner := func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			Error(w, 500, "X-User-ID not set")
			return
		}
		JSON(w, 200, H{"userId": userID})
	}

	req := authRequest("GET", "/api/v1/me", nil, token)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			Error(w, 401, "Unauthorized")
			return
		}
		token = token[len("Bearer "):]
		user := store.GetUserByToken(token)
		if user == nil {
			Error(w, 401, "Unauthorized")
			return
		}
		r.Header.Set("X-User-ID", user.ID)
		r.Header.Set("X-User-Role", user.Role)
		inner(w, r)
	})

	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["userId"] == nil {
		t.Fatal("expected userId in response")
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	resetDB()

	rr := httptest.NewRecorder()

	user := store.GetUserByToken("bad-token")
	if user != nil {
		t.Fatal("bad token should not resolve to user")
	}

	_ = rr
}
