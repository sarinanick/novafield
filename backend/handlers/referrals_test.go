package handlers

import (
	"encoding/json"
	"novafield-api/database"
	"testing"
)

func TestGetReferralLink(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("GET", "/api/v1/referrals/link", nil, token)
	rr := newRecorder()

	GetReferralLinkHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["code"] == nil {
		t.Fatal("expected referral code")
	}
	if result["link"] == nil {
		t.Fatal("expected referral link")
	}
}

func TestGetReferralLink_Unauthorized(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/referrals/link", nil, "")
	rr := newRecorder()

	GetReferralLinkHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestGetReferralStats(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("GET", "/api/v1/referrals/stats", nil, token)
	rr := newRecorder()

	GetReferralStatsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["totalReferrals"].(float64) != 0 {
		t.Error("expected 0 referrals")
	}
}

func TestGetReferralEarnings(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("GET", "/api/v1/referrals/earnings", nil, token)
	rr := newRecorder()

	GetReferralEarningsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestListAssessments(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/assessments", nil, "")
	rr := newRecorder()

	ListAssessmentsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var assessments []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&assessments)
	if len(assessments) < 3 {
		t.Errorf("expected at least 3 assessments, got %d", len(assessments))
	}
}

func TestGetAssessment(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/assessments/assess-gpt4", nil, "")
	rr := newRecorder()

	GetAssessmentHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["id"] != "assess-gpt4" {
		t.Errorf("expected assess-gpt4, got %v", result["id"])
	}
}

func TestStartAssessment(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	req := authRequest("POST", "/api/v1/assessments/assess-gpt4/start", nil, token)
	rr := newRecorder()

	StartAssessmentHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["questions"] == nil {
		t.Fatal("expected questions")
	}
}

func TestSubmitAssessment_Pass(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"answers": []int{2, 0, 1, 3, 1},
	})
	req := authRequest("POST", "/api/v1/assessments/assess-gpt4/submit", body, token)
	rr := newRecorder()

	SubmitAssessmentHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["passed"] != true {
		t.Error("expected to pass")
	}
	if result["score"].(float64) < 70 {
		t.Error("expected score >= 70")
	}

	d := database.GetDB()
	d.Mu.RLock()
	if len(d.Badges) != 1 {
		t.Errorf("expected 1 badge, got %d", len(d.Badges))
	}
	d.Mu.RUnlock()
}

func TestSubmitAssessment_Fail(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"answers": []int{0, 1, 2, 0, 0},
	})
	req := authRequest("POST", "/api/v1/assessments/assess-gpt4/submit", body, token)
	rr := newRecorder()

	SubmitAssessmentHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["passed"] != false {
		t.Error("expected to fail")
	}

	d := database.GetDB()
	d.Mu.RLock()
	if len(d.Badges) != 0 {
		t.Errorf("expected 0 badges, got %d", len(d.Badges))
	}
	d.Mu.RUnlock()
}

func TestGetAssessmentResults(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{"answers": []int{2, 0, 1, 3, 1}})
	req := authRequest("POST", "/api/v1/assessments/assess-gpt4/submit", body, token)
	rr := newRecorder()
	SubmitAssessmentHandler(rr, req)

	resultsReq := authRequest("GET", "/api/v1/assessments/results", nil, token)
	resultsRR := newRecorder()
	GetAssessmentResultsHandler(resultsRR, resultsReq)

	if resultsRR.Code != 200 {
		t.Fatalf("expected 200, got %d", resultsRR.Code)
	}

	var results []map[string]interface{}
	json.NewDecoder(resultsRR.Body).Decode(&results)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestGetUserBadges(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{"answers": []int{2, 0, 1, 3, 1}})
	req := authRequest("POST", "/api/v1/assessments/assess-gpt4/submit", body, token)
	rr := newRecorder()
	SubmitAssessmentHandler(rr, req)

	badgesReq := authRequest("GET", "/api/v1/users/"+user.ID+"/badges", nil, "")
	badgesRR := newRecorder()
	GetUserBadgesHandler(badgesRR, badgesReq)

	if badgesRR.Code != 200 {
		t.Fatalf("expected 200, got %d", badgesRR.Code)
	}

	var badges []map[string]interface{}
	json.NewDecoder(badgesRR.Body).Decode(&badges)
	if len(badges) != 1 {
		t.Errorf("expected 1 badge, got %d", len(badges))
	}
}

func TestGetSkillLeaderboard(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{"answers": []int{2, 0, 1, 3, 1}})
	req := authRequest("POST", "/api/v1/assessments/assess-gpt4/submit", body, token)
	rr := newRecorder()
	SubmitAssessmentHandler(rr, req)

	leaderReq := authRequest("GET", "/api/v1/assessments/leaderboard/ai-chatbots", nil, "")
	leaderRR := newRecorder()
	GetSkillLeaderboardHandler(leaderRR, leaderReq)

	if leaderRR.Code != 200 {
		t.Fatalf("expected 200, got %d", leaderRR.Code)
	}

	var entries []map[string]interface{}
	json.NewDecoder(leaderRR.Body).Decode(&entries)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	_ = user
}
