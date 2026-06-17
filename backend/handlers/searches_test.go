package handlers

import (
	"encoding/json"
	"testing"
)

func TestCreateSavedSearch(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"query":      "AI video",
		"category":   "ai-video",
		"minPrice":   50,
		"maxPrice":   500,
		"alertEmail": true,
	})
	req := authRequest("POST", "/api/v1/searches", body, token)
	rr := newRecorder()

	CreateSavedSearchHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateSavedSearch_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"query": "test"})
	req := authRequest("POST", "/api/v1/searches", body, "")
	rr := newRecorder()

	CreateSavedSearchHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestListSavedSearches(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"query": "AI video"})
	req := authRequest("POST", "/api/v1/searches", body, token)
	rr := newRecorder()
	CreateSavedSearchHandler(rr, req)

	listReq := authRequest("GET", "/api/v1/searches", nil, token)
	listRR := newRecorder()
	ListSavedSearchesHandler(listRR, listReq)

	if listRR.Code != 200 {
		t.Fatalf("expected 200, got %d", listRR.Code)
	}

	var searches []map[string]interface{}
	json.NewDecoder(listRR.Body).Decode(&searches)
	if len(searches) != 1 {
		t.Errorf("expected 1 search, got %d", len(searches))
	}
}

func TestDeleteSavedSearch(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"query": "AI video"})
	req := authRequest("POST", "/api/v1/searches", body, token)
	rr := newRecorder()
	CreateSavedSearchHandler(rr, req)

	searchID := decodeJSON(rr)["id"].(string)

	deleteReq := authRequest("DELETE", "/api/v1/searches/"+searchID, nil, token)
	deleteRR := newRecorder()
	DeleteSavedSearchHandler(deleteRR, deleteReq)

	if deleteRR.Code != 200 {
		t.Fatalf("expected 200, got %d", deleteRR.Code)
	}
}

func TestTrendingSearches(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/searches/trending", nil, "")
	rr := newRecorder()

	TrendingSearchesHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestScoreGig(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"title":       "I will create professional AI videos with Sora 2",
		"description": "Professional AI video generation using Sora 2 and Kling 3.0. I offer high quality cinematic videos with fast delivery and unlimited revisions. Satisfaction guaranteed!",
		"category":    "ai-video",
		"tags":        []string{"Sora 2", "AI video", "cinematic", "4K", "professional"},
	})
	req := authRequest("POST", "/api/v1/ai/gig/score", body, token)
	rr := newRecorder()

	ScoreGigHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["score"].(float64) < 5 {
		t.Errorf("expected score >= 5, got %v", result["score"])
	}
}

func TestScoreGig_PoorGig(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"title":       "video",
		"description": "I do video",
		"tags":        []string{},
	})
	req := authRequest("POST", "/api/v1/ai/gig/score", body, token)
	rr := newRecorder()

	ScoreGigHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["score"].(float64) > 6 {
		t.Errorf("expected low score, got %v", result["score"])
	}
	if len(result["suggestions"].([]interface{})) == 0 {
		t.Error("expected suggestions for poor gig")
	}
}

func TestScoreGig_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"title": "test"})
	req := authRequest("POST", "/api/v1/ai/gig/score", body, "")
	rr := newRecorder()

	ScoreGigHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
