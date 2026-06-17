package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"testing"
)

func seedTestGig(freelancerID string) (string, string) {
	gigID := store.NewID()
	pkgID := store.NewID()

	d := database.GetDB()
	d.Mu.Lock()
	d.Gigs = append(d.Gigs, models.Gig{
		ID:           gigID,
		FreelancerID: freelancerID,
		Title:        "Test Gig",
		Description:  "A test gig",
		Category:     "ai-video",
		Price:        50,
		Status:       "active",
		CreatedAt:    store.Now(),
		UpdatedAt:    store.Now(),
	})
	d.Packages = append(d.Packages, models.Package{
		ID:           pkgID,
		GigID:        gigID,
		Name:         "Basic",
		Description:  "Basic package",
		Price:        50,
		DeliveryDays: 3,
		Revisions:    1,
	})
	d.Mu.Unlock()

	return gigID, pkgID
}

func TestCreateGig_AsFreelancer(t *testing.T) {
	resetDB()

	user, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"title":       "My AI Gig",
		"description": "I will do AI stuff",
		"category":    "ai-video",
		"tags":        []string{"AI", "video"},
		"priceType":   "fixed",
		"packages": []map[string]interface{}{
			{"name": "Basic", "price": 25, "deliveryDays": 2, "revisions": 1},
		},
	})
	req := authRequest("POST", "/api/v1/gigs", body, token)
	rr := newRecorder()

	CreateGigHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["id"] == nil {
		t.Fatal("expected gig id in response")
	}
	if result["message"] != "Gig created" {
		t.Errorf("expected 'Gig created', got %v", result["message"])
	}

	_ = user
}

func TestCreateGig_AsClient_Forbidden(t *testing.T) {
	resetDB()

	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"title":    "Should Fail",
		"category": "ai-video",
	})
	req := authRequest("POST", "/api/v1/gigs", body, token)
	rr := newRecorder()

	CreateGigHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestCreateGig_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"title": "No Auth"})
	req := authRequest("POST", "/api/v1/gigs", body, "")
	rr := newRecorder()

	CreateGigHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGetGigs_Empty(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/gigs", nil, "")
	rr := newRecorder()

	GetGigsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	gigs, ok := result["gigs"].([]interface{})
	if !ok {
		t.Fatal("expected gigs array")
	}
	if len(gigs) != 0 {
		t.Errorf("expected 0 gigs, got %d", len(gigs))
	}
}

func TestGetGigs_WithResults(t *testing.T) {
	resetDB()

	user, _ := createTestUser("freelancer")
	seedTestGig(user.ID)

	req := authRequest("GET", "/api/v1/gigs", nil, "")
	rr := newRecorder()

	GetGigsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	gigs := result["gigs"].([]interface{})
	if len(gigs) != 1 {
		t.Fatalf("expected 1 gig, got %d", len(gigs))
	}

	gig := gigs[0].(map[string]interface{})
	if gig["title"] != "Test Gig" {
		t.Errorf("expected title 'Test Gig', got %v", gig["title"])
	}
}

func TestGetGig_ByID(t *testing.T) {
	resetDB()

	user, _ := createTestUser("freelancer")
	gigID, _ := seedTestGig(user.ID)

	req := authRequest("GET", "/api/v1/gigs/"+gigID, nil, "")
	rr := newRecorder()

	GetGigHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["id"] != gigID {
		t.Errorf("expected id %s, got %v", gigID, result["id"])
	}
}

func TestGetGig_NotFound(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/gigs/nonexistent", nil, "")
	rr := newRecorder()

	GetGigHandler(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteGig_AsOwner(t *testing.T) {
	resetDB()

	user, token := createTestUser("freelancer")
	gigID, _ := seedTestGig(user.ID)

	req := authRequest("DELETE", "/api/v1/gigs/"+gigID, nil, token)
	rr := newRecorder()

	DeleteGigHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteGig_AsNonOwner(t *testing.T) {
	resetDB()

	owner, _ := createTestUser("freelancer")
	_, token := createTestUser("freelancer")
	gigID, _ := seedTestGig(owner.ID)

	req := authRequest("DELETE", "/api/v1/gigs/"+gigID, nil, token)
	rr := newRecorder()

	DeleteGigHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGetMyGigs(t *testing.T) {
	resetDB()

	user, token := createTestUser("freelancer")
	seedTestGig(user.ID)

	req := authRequest("GET", "/api/v1/gigs/my-gigs", nil, token)
	rr := newRecorder()

	GetMyGigsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var gigs []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&gigs)
	if len(gigs) != 1 {
		t.Fatalf("expected 1 gig, got %d", len(gigs))
	}
}

func TestGetMyGigs_Unauthorized(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/gigs/my-gigs", nil, "")
	rr := newRecorder()

	GetMyGigsHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func newRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}
