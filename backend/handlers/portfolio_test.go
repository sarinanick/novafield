package handlers

import (
	"encoding/json"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"testing"
)

func TestCreateCaseStudy(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"title":       "E-commerce AI Chatbot",
		"clientName":  "TechCorp",
		"category":    "ai-chatbots",
		"challenge":   "Customer support overwhelmed with 1000+ daily queries",
		"approach":    "Built GPT-4 powered chatbot with custom knowledge base",
		"results":     "Reduced support tickets by 70%, improved response time by 5x",
		"skills":      []string{"GPT-4", "LangChain", "Python"},
		"duration":    "3 weeks",
		"budgetRange": "$2000-$5000",
		"isPublic":    true,
	})
	req := authRequest("POST", "/api/v1/portfolio/cases", body, token)
	rr := newRecorder()

	CreateCaseStudyHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["id"] == nil {
		t.Fatal("expected case study id")
	}
}

func TestCreateCaseStudy_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"title": "test"})
	req := authRequest("POST", "/api/v1/portfolio/cases", body, "")
	rr := newRecorder()

	CreateCaseStudyHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestCreateCaseStudy_MissingTitle(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{"challenge": "no title"})
	req := authRequest("POST", "/api/v1/portfolio/cases", body, token)
	rr := newRecorder()

	CreateCaseStudyHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetCaseStudy_Public(t *testing.T) {
	resetDB()
	user, _ := createTestUser("freelancer")
	csID := createTestCaseStudy(user.ID, true)

	req := authRequest("GET", "/api/v1/portfolio/cases/"+csID, nil, "")
	rr := newRecorder()

	GetCaseStudyHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["id"] != csID {
		t.Errorf("expected id %s, got %v", csID, result["id"])
	}
}

func TestGetCaseStudy_Private_Owner(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	csID := createTestCaseStudy(user.ID, false)

	req := authRequest("GET", "/api/v1/portfolio/cases/"+csID, nil, token)
	rr := newRecorder()

	GetCaseStudyHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetCaseStudy_Private_Other(t *testing.T) {
	resetDB()
	user, _ := createTestUser("freelancer")
	_, token := createTestUser("client")
	csID := createTestCaseStudy(user.ID, false)

	req := authRequest("GET", "/api/v1/portfolio/cases/"+csID, nil, token)
	rr := newRecorder()

	GetCaseStudyHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGetCaseStudy_NotFound(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/portfolio/cases/nonexistent", nil, "")
	rr := newRecorder()

	GetCaseStudyHandler(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateCaseStudy(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	csID := createTestCaseStudy(user.ID, true)

	body := jsonBody(map[string]interface{}{
		"title":    "Updated Title",
		"results":  "Even better results",
		"isPublic": false,
	})
	req := authRequest("PUT", "/api/v1/portfolio/cases/"+csID, body, token)
	rr := newRecorder()

	UpdateCaseStudyHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, cs := range d.CaseStudies {
		if cs.ID == csID {
			if cs.Title != "Updated Title" {
				t.Errorf("expected updated title, got %s", cs.Title)
			}
			if cs.IsPublic != false {
				t.Error("expected isPublic to be false")
			}
			break
		}
	}
	d.Mu.RUnlock()
}

func TestUpdateCaseStudy_NonOwner(t *testing.T) {
	resetDB()
	owner, _ := createTestUser("freelancer")
	_, token := createTestUser("freelancer")
	csID := createTestCaseStudy(owner.ID, true)

	body := jsonBody(map[string]interface{}{"title": "Hacked"})
	req := authRequest("PUT", "/api/v1/portfolio/cases/"+csID, body, token)
	rr := newRecorder()

	UpdateCaseStudyHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestDeleteCaseStudy(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	csID := createTestCaseStudy(user.ID, true)

	req := authRequest("DELETE", "/api/v1/portfolio/cases/"+csID, nil, token)
	rr := newRecorder()

	DeleteCaseStudyHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, cs := range d.CaseStudies {
		if cs.ID == csID {
			t.Error("case study should be deleted")
		}
	}
	d.Mu.RUnlock()
}

func TestGetUserPortfolio(t *testing.T) {
	resetDB()
	user, _ := createTestUser("freelancer")
	createTestCaseStudy(user.ID, true)
	createTestCaseStudy(user.ID, false)

	req := authRequest("GET", "/api/v1/users/"+user.ID+"/portfolio", nil, "")
	rr := newRecorder()

	GetUserPortfolioHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var cases []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&cases)
	if len(cases) != 1 {
		t.Errorf("expected 1 public case study, got %d", len(cases))
	}
}

func TestAttachCaseToGig(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	gigID, _ := seedTestGig(user.ID)
	csID := createTestCaseStudy(user.ID, true)

	req := authRequest("POST", "/api/v1/gigs/"+gigID+"/portfolio/"+csID, nil, token)
	rr := newRecorder()

	AttachCaseToGigHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, cs := range d.CaseStudies {
		if cs.ID == csID {
			found := false
			for _, id := range cs.LinkedGigIDs {
				if id == gigID {
					found = true
					break
				}
			}
			if !found {
				t.Error("expected gig to be linked")
			}
			break
		}
	}
	d.Mu.RUnlock()
}

func TestAttachCaseToGig_NonOwner(t *testing.T) {
	resetDB()
	owner, _ := createTestUser("freelancer")
	_, token := createTestUser("freelancer")
	gigID, _ := seedTestGig(owner.ID)
	csID := createTestCaseStudy(owner.ID, true)

	req := authRequest("POST", "/api/v1/gigs/"+gigID+"/portfolio/"+csID, nil, token)
	rr := newRecorder()

	AttachCaseToGigHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestDetachCaseFromGig(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	gigID, _ := seedTestGig(user.ID)
	csID := createTestCaseStudy(user.ID, true)

	attachReq := authRequest("POST", "/api/v1/gigs/"+gigID+"/portfolio/"+csID, nil, token)
	attachRR := newRecorder()
	AttachCaseToGigHandler(attachRR, attachReq)

	detachReq := authRequest("DELETE", "/api/v1/gigs/"+gigID+"/portfolio/"+csID, nil, token)
	detachRR := newRecorder()
	DetachCaseFromGigHandler(detachRR, detachReq)

	if detachRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", detachRR.Code, detachRR.Body.String())
	}
}

func TestSubmitTestimonial(t *testing.T) {
	resetDB()
	user, _ := createTestUser("freelancer")
	_, token := createTestUser("client")
	csID := createTestCaseStudy(user.ID, true)

	body := jsonBody(map[string]interface{}{
		"rating": 5,
		"text":   "Amazing work! Exceeded all expectations.",
	})
	req := authRequest("POST", "/api/v1/portfolio/cases/"+csID+"/testimonial", body, token)
	rr := newRecorder()

	SubmitTestimonialHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	if len(d.Testimonials) != 1 {
		t.Errorf("expected 1 testimonial, got %d", len(d.Testimonials))
	}
	d.Mu.RUnlock()
}

func TestSubmitTestimonial_InvalidRating(t *testing.T) {
	resetDB()
	user, _ := createTestUser("freelancer")
	_, token := createTestUser("client")
	csID := createTestCaseStudy(user.ID, true)

	body := jsonBody(map[string]interface{}{
		"rating": 6,
		"text":   "Invalid rating",
	})
	req := authRequest("POST", "/api/v1/portfolio/cases/"+csID+"/testimonial", body, token)
	rr := newRecorder()

	SubmitTestimonialHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func createTestCaseStudy(userID string, isPublic bool) string {
	csID := store.NewID()
	d := database.GetDB()
	d.Mu.Lock()
	d.CaseStudies = append(d.CaseStudies, models.CaseStudy{
		ID: csID, UserID: userID, Title: "Test Case Study",
		ClientName: "Test Client", Category: "ai-chatbots",
		Challenge: "Test challenge", Approach: "Test approach",
		Results: "Test results", Images: []string{}, Skills: []string{"Go"},
		Duration: "2 weeks", LinkedGigIDs: []string{},
		IsPublic: isPublic, CreatedAt: store.Now(),
	})
	d.Mu.Unlock()
	return csID
}
