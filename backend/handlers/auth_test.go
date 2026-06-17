package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"novafield-api/database"
	"novafield-api/models"
	"testing"
)

func TestRegister_Success(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]string{
		"email":    "new@test.com",
		"password": "password123",
		"name":     "New User",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	RegisterHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["token"] == nil {
		t.Fatal("expected token in response")
	}
	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatal("expected user in response")
	}
	if user["name"] != "New User" {
		t.Errorf("expected name New User, got %v", user["name"])
	}
	if user["role"] != "client" {
		t.Errorf("expected role client, got %v", user["role"])
	}
}

func TestRegister_Freelancer(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]string{
		"email":    "free@test.com",
		"password": "password123",
		"name":     "Freelancer",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/register?role=freelancer", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	RegisterHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	user := result["user"].(map[string]interface{})
	if user["role"] != "freelancer" {
		t.Errorf("expected role freelancer, got %v", user["role"])
	}
}

func TestRegister_MissingFields(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]string{"email": "x@x.com"})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	RegisterHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]string{
		"email":    "notanemail",
		"password": "password123",
		"name":     "Test",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	RegisterHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["error"] != "Invalid email format" {
		t.Errorf("expected invalid email error, got %v", result["error"])
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]string{
		"email":    "test@test.com",
		"password": "123",
		"name":     "Test",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	RegisterHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["error"] != "Password must be at least 6 characters" {
		t.Errorf("expected short password error, got %v", result["error"])
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]string{
		"email":    "dup@test.com",
		"password": "password123",
		"name":     "First",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	RegisterHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("first register: expected 201, got %d", rr.Code)
	}

	body2 := jsonBody(map[string]string{
		"email":    "dup@test.com",
		"password": "password123",
		"name":     "Second",
	})
	req2 := httptest.NewRequest("POST", "/api/v1/auth/register", body2)
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	RegisterHandler(rr2, req2)

	if rr2.Code != 409 {
		t.Fatalf("duplicate: expected 409, got %d", rr2.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	resetDB()

	regBody := jsonBody(map[string]string{
		"email":    "login@test.com",
		"password": "password123",
		"name":     "Login User",
	})
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", regBody)
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	RegisterHandler(regRR, regReq)

	loginBody := jsonBody(map[string]string{
		"email":    "login@test.com",
		"password": "password123",
	})
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	LoginHandler(loginRR, loginReq)

	if loginRR.Code != 200 {
		t.Fatalf("expected 200, got %d", loginRR.Code)
	}

	result := decodeJSON(loginRR)
	if result["token"] == nil {
		t.Fatal("expected token in response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	resetDB()

	regBody := jsonBody(map[string]string{
		"email":    "wrong@test.com",
		"password": "password123",
		"name":     "Wrong Pass",
	})
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", regBody)
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	RegisterHandler(regRR, regReq)

	loginBody := jsonBody(map[string]string{
		"email":    "wrong@test.com",
		"password": "wrongpassword",
	})
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	LoginHandler(loginRR, loginReq)

	if loginRR.Code != 401 {
		t.Fatalf("expected 401, got %d", loginRR.Code)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	resetDB()

	loginBody := jsonBody(map[string]string{
		"email":    "nobody@test.com",
		"password": "password123",
	})
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	LoginHandler(loginRR, loginReq)

	if loginRR.Code != 401 {
		t.Fatalf("expected 401, got %d", loginRR.Code)
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %v", result["status"])
	}
	if result["service"] != "novafield-api" {
		t.Errorf("expected service novafield-api, got %v", result["service"])
	}
}

func TestRegister_Login_RoundTrip(t *testing.T) {
	resetDB()

	regBody := jsonBody(map[string]string{
		"email":    "roundtrip@test.com",
		"password": "mypassword",
		"name":     "Round Trip",
	})
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", regBody)
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	RegisterHandler(regRR, regReq)

	if regRR.Code != 201 {
		t.Fatalf("register: expected 201, got %d", regRR.Code)
	}

	loginBody := jsonBody(map[string]string{
		"email":    "roundtrip@test.com",
		"password": "mypassword",
	})
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	LoginHandler(loginRR, loginReq)

	if loginRR.Code != 200 {
		t.Fatalf("login: expected 200, got %d", loginRR.Code)
	}

	result := decodeJSON(loginRR)
	token, ok := result["token"].(string)
	if !ok || token == "" {
		t.Fatal("expected non-empty token string")
	}

	meReq := authRequest("GET", "/api/v1/me", nil, token)
	meRR := httptest.NewRecorder()
	GetMeHandler(meRR, meReq)

	if meRR.Code != 200 {
		t.Fatalf("me: expected 200, got %d", meRR.Code)
	}

	meResult := decodeJSON(meRR)
	if meResult["email"] != "roundtrip@test.com" {
		t.Errorf("expected email roundtrip@test.com, got %v", meResult["email"])
	}
}

func TestGetMe_Unauthorized(t *testing.T) {
	resetDB()

	req := httptest.NewRequest("GET", "/api/v1/me", nil)
	rr := httptest.NewRecorder()

	GetMeHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestGetMe_InvalidToken(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/me", nil, "invalid-token")
	rr := httptest.NewRecorder()

	GetMeHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestUpdateProfile(t *testing.T) {
	resetDB()

	user, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"name":       "Updated Name",
		"bio":        "New bio",
		"skills":     []string{"Go", "Rust"},
		"hourlyRate": 100,
	})
	req := authRequest("PUT", "/api/v1/me", body, token)
	rr := httptest.NewRecorder()

	UpdateProfileHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	meReq := authRequest("GET", "/api/v1/me", nil, token)
	meRR := httptest.NewRecorder()
	GetMeHandler(meRR, meReq)

	if meRR.Code != 200 {
		t.Fatalf("get me: expected 200, got %d", meRR.Code)
	}

	meResult := decodeJSON(meRR)
	if meResult["name"] != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %v", meResult["name"])
	}
	if meResult["bio"] != "New bio" {
		t.Errorf("expected bio 'New bio', got %v", meResult["bio"])
	}

	_ = user
}

func TestGetProfile_Public(t *testing.T) {
	resetDB()

	user, _ := createTestUser("freelancer")

	req := httptest.NewRequest("GET", "/api/v1/users/"+user.ID, nil)
	rr := httptest.NewRecorder()

	GetProfileHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["id"] != user.ID {
		t.Errorf("expected id %s, got %v", user.ID, result["id"])
	}
	if result["passwordHash"] != nil {
		t.Error("passwordHash should not be exposed in public profile")
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	resetDB()

	req := httptest.NewRequest("GET", "/api/v1/users/nonexistent", nil)
	rr := httptest.NewRecorder()

	GetProfileHandler(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetCategories(t *testing.T) {
	resetDB()

	d := database.GetDB()
	d.Mu.Lock()
	d.Categories = []models.Category{
		{ID: "cat-1", Name: "AI Video", Slug: "ai-video"},
		{ID: "cat-2", Name: "AI Image", Slug: "ai-image"},
	}
	d.Mu.Unlock()

	req := httptest.NewRequest("GET", "/api/v1/categories", nil)
	rr := httptest.NewRecorder()

	GetCategoriesHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var categories []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&categories)
	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}
}
