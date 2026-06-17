package handlers

import (
	"net/http/httptest"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"testing"
)

func seedRecommendationData() (*models.User, string) {
	d := database.GetDB()
	d.Mu.Lock()

	d.Categories = []models.Category{
		{ID: "cat-1", Name: "AI Video", Slug: "ai-video"},
		{ID: "cat-2", Name: "AI Image", Slug: "ai-image"},
	}

	seller1 := models.User{
		ID: store.NewID(), Email: "rec_seller1@test.com", PasswordHash: store.HashPassword("pass"),
		Name: "Video Pro", Role: "freelancer", Skills: []string{"Sora"}, Rating: 4.8, JoinedAt: store.Now(),
	}
	seller2 := models.User{
		ID: store.NewID(), Email: "rec_seller2@test.com", PasswordHash: store.HashPassword("pass"),
		Name: "Image Pro", Role: "freelancer", Skills: []string{"DALL-E"}, Rating: 4.2, JoinedAt: store.Now(),
	}
	d.Users = append(d.Users, seller1, seller2)

	gig1 := models.Gig{
		ID: store.NewID(), FreelancerID: seller1.ID, Title: "AI Video Production",
		Category: "ai-video", Price: 200, Status: "active", Rating: 4.8, OrdersCount: 10,
		Tags: []string{"video", "AI"}, CreatedAt: store.Now(), UpdatedAt: store.Now(),
	}
	gig2 := models.Gig{
		ID: store.NewID(), FreelancerID: seller2.ID, Title: "AI Image Design",
		Category: "ai-image", Price: 100, Status: "active", Rating: 4.2, OrdersCount: 3,
		Tags: []string{"image", "AI"}, CreatedAt: store.Now(), UpdatedAt: store.Now(),
	}
	gig3 := models.Gig{
		ID: store.NewID(), FreelancerID: seller1.ID, Title: "Cinematic AI Clips",
		Category: "ai-video", Price: 300, Status: "active", Rating: 4.9, OrdersCount: 15,
		Tags: []string{"cinema", "AI"}, CreatedAt: store.Now(), UpdatedAt: store.Now(),
	}
	d.Gigs = append(d.Gigs, gig1, gig2, gig3)
	d.Mu.Unlock()

	client, clientToken := createTestUser("client")
	return client, clientToken
}

func TestGetRecommendations_Success(t *testing.T) {
	resetDB()
	_, token := seedRecommendationData()

	req := authRequest("GET", "/api/v1/recommendations/gigs", nil, token)
	rr := newRecorder()

	GetRecommendationsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	recs := result["recommendations"].([]interface{})
	if len(recs) == 0 {
		t.Fatal("expected at least 1 recommendation")
	}

	first := recs[0].(map[string]interface{})
	if first["score"].(float64) <= 0 {
		t.Error("expected positive score")
	}
	if first["gig"] == nil {
		t.Error("expected gig in recommendation")
	}
}

func TestGetRecommendations_Unauthorized(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/recommendations/gigs", nil, "")
	rr := newRecorder()

	GetRecommendationsHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestGetRecommendations_ExcludesOwnGigs(t *testing.T) {
	resetDB()
	client, token := seedRecommendationData()
	_ = client

	req := authRequest("GET", "/api/v1/recommendations/gigs", nil, token)
	rr := newRecorder()

	GetRecommendationsHandler(rr, req)

	result := decodeJSON(rr)
	recs := result["recommendations"].([]interface{})

	for _, rec := range recs {
		gig := rec.(map[string]interface{})["gig"].(map[string]interface{})
		if gig["freelancerId"] == client.ID {
			t.Error("should not recommend own gigs")
		}
	}
}

func TestGetSimilarGigs_Success(t *testing.T) {
	resetDB()
	seedRecommendationData()

	d := database.GetDB()
	d.Mu.RLock()
	var gigID string
	for _, g := range d.Gigs {
		if g.Category == "ai-video" {
			gigID = g.ID
			break
		}
	}
	d.Mu.RUnlock()

	req := httptest.NewRequest("GET", "/api/v1/recommendations/similar/"+gigID, nil)
	rr := newRecorder()

	GetSimilarGigsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	similar := result["similar"].([]interface{})
	if len(similar) == 0 {
		t.Fatal("expected at least 1 similar gig")
	}
}

func TestGetSimilarGigs_NotFound(t *testing.T) {
	resetDB()

	req := httptest.NewRequest("GET", "/api/v1/recommendations/similar/nonexistent", nil)
	rr := newRecorder()

	GetSimilarGigsHandler(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestRecommendationFeedback_Success(t *testing.T) {
	resetDB()
	_, token := seedRecommendationData()

	d := database.GetDB()
	d.Mu.RLock()
	gigID := d.Gigs[0].ID
	d.Mu.RUnlock()

	body := jsonBody(map[string]interface{}{
		"gigId":     gigID,
		"eventType": "click",
	})
	req := authRequest("POST", "/api/v1/recommendations/feedback", body, token)
	rr := newRecorder()

	RecommendationFeedbackHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d.Mu.RLock()
	if len(d.RecFeedback) != 1 {
		t.Errorf("expected 1 feedback, got %d", len(d.RecFeedback))
	}
	d.Mu.RUnlock()
}

func TestRecommendationFeedback_InvalidEvent(t *testing.T) {
	resetDB()
	_, token := seedRecommendationData()

	body := jsonBody(map[string]interface{}{
		"gigId":     "fake",
		"eventType": "invalid",
	})
	req := authRequest("POST", "/api/v1/recommendations/feedback", body, token)
	rr := newRecorder()

	RecommendationFeedbackHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRecommendationFeedback_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{
		"gigId":     "fake",
		"eventType": "click",
	})
	req := authRequest("POST", "/api/v1/recommendations/feedback", body, "")
	rr := newRecorder()

	RecommendationFeedbackHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
