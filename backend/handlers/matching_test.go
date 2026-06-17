package handlers

import (
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"testing"
)

func seedFreelancers() {
	d := database.GetDB()
	d.Mu.Lock()
	d.Users = append(d.Users, models.User{
		ID: store.NewID(), Email: "fl1@test.com", PasswordHash: store.HashPassword("pass"),
		Name: "Video Expert", Role: "freelancer", Skills: []string{"Sora 2", "Kling 3.0", "Cinema"},
		HourlyRate: 80, Rating: 4.8, ReviewsCount: 25, JoinedAt: store.Now(),
	})
	d.Users = append(d.Users, models.User{
		ID: store.NewID(), Email: "fl2@test.com", PasswordHash: store.HashPassword("pass"),
		Name: "Image Pro", Role: "freelancer", Skills: []string{"DALL-E 3", "Midjourney", "Photoshop"},
		HourlyRate: 50, Rating: 4.2, ReviewsCount: 10, JoinedAt: store.Now(),
	})
	d.Users = append(d.Users, models.User{
		ID: store.NewID(), Email: "fl3@test.com", PasswordHash: store.HashPassword("pass"),
		Name: "Full Stack AI", Role: "freelancer", Skills: []string{"Sora 2", "DALL-E 3", "GPT-4", "Python"},
		HourlyRate: 120, Rating: 4.9, ReviewsCount: 50, JoinedAt: store.Now(),
	})
	d.Mu.Unlock()
}

func TestMatchFreelancers_Success(t *testing.T) {
	resetDB()
	seedFreelancers()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"title":       "AI Video Production",
		"description": "Need a cinematic AI video for product launch",
		"category":    "ai-video",
		"skills":      []string{"Sora 2", "Cinema"},
		"budgetMin":   100,
		"budgetMax":   500,
		"timeline":    "1_week",
	})
	req := authRequest("POST", "/api/v1/match/freelancers", body, token)
	rr := newRecorder()

	MatchFreelancersHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["briefId"] == nil {
		t.Fatal("expected briefId")
	}

	matches, ok := result["matches"].([]interface{})
	if !ok {
		t.Fatal("expected matches array")
	}
	if len(matches) == 0 {
		t.Fatal("expected at least 1 match")
	}

	first := matches[0].(map[string]interface{})
	if first["score"].(float64) <= 0 {
		t.Error("expected positive score")
	}
	if first["freelancerId"] == nil {
		t.Error("expected freelancerId")
	}
}

func TestMatchFreelancers_RankedByScore(t *testing.T) {
	resetDB()
	seedFreelancers()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"title":    "Need Sora expert",
		"skills":   []string{"Sora 2"},
		"budgetMin": 50,
		"budgetMax": 200,
	})
	req := authRequest("POST", "/api/v1/match/freelancers", body, token)
	rr := newRecorder()

	MatchFreelancersHandler(rr, req)

	result := decodeJSON(rr)
	matches := result["matches"].([]interface{})

	for i := 1; i < len(matches); i++ {
		prev := matches[i-1].(map[string]interface{})["score"].(float64)
		curr := matches[i].(map[string]interface{})["score"].(float64)
		if curr > prev {
			t.Errorf("match %d score %f > match %d score %f", i, curr, i-1, prev)
		}
	}
}

func TestMatchFreelancers_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"title": "test"})
	req := authRequest("POST", "/api/v1/match/freelancers", body, "")
	rr := newRecorder()

	MatchFreelancersHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestMatchFreelancers_MissingTitle(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"description": "no title"})
	req := authRequest("POST", "/api/v1/match/freelancers", body, token)
	rr := newRecorder()

	MatchFreelancersHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestMatchFreelancers_SkillMatch(t *testing.T) {
	resetDB()
	seedFreelancers()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"title":  "Need DALL-E expert",
		"skills": []string{"DALL-E 3"},
	})
	req := authRequest("POST", "/api/v1/match/freelancers", body, token)
	rr := newRecorder()

	MatchFreelancersHandler(rr, req)

	result := decodeJSON(rr)
	matches := result["matches"].([]interface{})

	foundDALLE := false
	for _, m := range matches {
		match := m.(map[string]interface{})
		flID := match["freelancerId"].(string)
		fl := store.FindUserByID(flID)
		if fl != nil {
			for _, s := range fl.Skills {
				if s == "DALL-E 3" {
					foundDALLE = true
					break
				}
			}
		}
	}
	if !foundDALLE {
		t.Error("expected DALL-E freelancer in results")
	}
}

func TestGetMatchProjects_AsFreelancer(t *testing.T) {
	resetDB()
	seedFreelancers()

	d := database.GetDB()
	d.Mu.RLock()
	var freelancer models.User
	for _, u := range d.Users {
		if u.Role == "freelancer" && u.Name == "Video Expert" {
			freelancer = u
			break
		}
	}
	d.Mu.RUnlock()

	client, clientToken := createTestUser("client")
	_ = client

	body := jsonBody(map[string]interface{}{
		"title":  "Video project",
		"skills": []string{"Sora 2"},
	})
	req := authRequest("POST", "/api/v1/match/freelancers", body, clientToken)
	rr := newRecorder()
	MatchFreelancersHandler(rr, req)

	token := store.GenerateToken(freelancer.ID, freelancer.Email, freelancer.Role)

	projReq := authRequest("GET", "/api/v1/match/projects", nil, token)
	projRR := newRecorder()
	GetMatchProjectsHandler(projRR, projReq)

	if projRR.Code != 200 {
		t.Fatalf("expected 200, got %d", projRR.Code)
	}
}

func TestGetMatchProjects_AsClient_Forbidden(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("GET", "/api/v1/match/projects", nil, token)
	rr := newRecorder()

	GetMatchProjectsHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestMatchFeedback(t *testing.T) {
	resetDB()
	seedFreelancers()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"title":  "Test project",
		"skills": []string{"Sora 2"},
	})
	req := authRequest("POST", "/api/v1/match/freelancers", body, token)
	rr := newRecorder()
	MatchFreelancersHandler(rr, req)

	result := decodeJSON(rr)
	matches := result["matches"].([]interface{})
	matchID := matches[0].(map[string]interface{})["id"].(string)

	feedbackBody := jsonBody(map[string]interface{}{
		"matchId": matchID,
		"action":  "viewed",
	})
	fbReq := authRequest("POST", "/api/v1/match/feedback", feedbackBody, token)
	fbRR := newRecorder()
	MatchFeedbackHandler(fbRR, fbReq)

	if fbRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", fbRR.Code, fbRR.Body.String())
	}
}

func TestMatchFeedback_InvalidAction(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"matchId": "fake",
		"action":  "invalid",
	})
	req := authRequest("POST", "/api/v1/match/feedback", body, token)
	rr := newRecorder()

	MatchFeedbackHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestMatchHistory(t *testing.T) {
	resetDB()
	seedFreelancers()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"title":  "History test",
		"skills": []string{"Sora 2"},
	})
	req := authRequest("POST", "/api/v1/match/freelancers", body, token)
	rr := newRecorder()
	MatchFreelancersHandler(rr, req)

	historyReq := authRequest("GET", "/api/v1/match/history", nil, token)
	historyRR := newRecorder()
	MatchHistoryHandler(historyRR, historyReq)

	if historyRR.Code != 200 {
		t.Fatalf("expected 200, got %d", historyRR.Code)
	}
}

func TestMatchFreelancers_SavedToDB(t *testing.T) {
	resetDB()
	seedFreelancers()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"title":  "DB persistence test",
		"skills": []string{"Sora 2"},
	})
	req := authRequest("POST", "/api/v1/match/freelancers", body, token)
	rr := newRecorder()
	MatchFreelancersHandler(rr, req)

	d := database.GetDB()
	d.Mu.RLock()
	if len(d.ProjectBriefs) != 1 {
		t.Errorf("expected 1 brief, got %d", len(d.ProjectBriefs))
	}
	if len(d.MatchResults) == 0 {
		t.Error("expected match results saved")
	}
	d.Mu.RUnlock()
}
