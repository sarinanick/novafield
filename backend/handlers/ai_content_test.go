package handlers

import (
	"testing"
)

func TestGenerateGig_Success(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"serviceType":    "AI Video Generation",
		"skills":         []string{"Sora 2", "Kling 3.0"},
		"experience":     "expert",
		"targetAudience": "startups",
		"tone":           "professional",
		"priceMin":       100,
		"priceMax":       500,
		"uniqueSelling":  "10+ years in video production",
	})
	req := authRequest("POST", "/api/v1/ai/gig/generate", body, token)
	rr := newRecorder()

	GenerateGigHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["title"] == nil || result["title"] == "" {
		t.Fatal("expected title")
	}
	if result["description"] == nil || result["description"] == "" {
		t.Fatal("expected description")
	}

	packages, ok := result["packages"].([]interface{})
	if !ok || len(packages) != 3 {
		t.Fatalf("expected 3 packages, got %v", result["packages"])
	}

	faq, ok := result["faq"].([]interface{})
	if !ok || len(faq) < 3 {
		t.Fatalf("expected at least 3 FAQ entries, got %v", result["faq"])
	}
}

func TestGenerateGig_MinimalInput(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"serviceType": "Logo Design",
	})
	req := authRequest("POST", "/api/v1/ai/gig/generate", body, token)
	rr := newRecorder()

	GenerateGigHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["title"] == nil {
		t.Fatal("expected title")
	}
}

func TestGenerateGig_Client_Forbidden(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"serviceType": "test"})
	req := authRequest("POST", "/api/v1/ai/gig/generate", body, token)
	rr := newRecorder()

	GenerateGigHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGenerateGig_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"serviceType": "test"})
	req := authRequest("POST", "/api/v1/ai/gig/generate", body, "")
	rr := newRecorder()

	GenerateGigHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGenerateGig_MissingServiceType(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{"skills": []string{"Go"}})
	req := authRequest("POST", "/api/v1/ai/gig/generate", body, token)
	rr := newRecorder()

	GenerateGigHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestImproveGig_Success(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"title":       "Video editing services",
		"description": "I edit videos for you.",
	})
	req := authRequest("POST", "/api/v1/ai/gig/improve", body, token)
	rr := newRecorder()

	ImproveGigHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["improvedTitle"] == nil {
		t.Fatal("expected improvedTitle")
	}
	if result["improvedDescription"] == nil {
		t.Fatal("expected improvedDescription")
	}
}

func TestImproveGig_MissingDescription(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{"title": "test"})
	req := authRequest("POST", "/api/v1/ai/gig/improve", body, token)
	rr := newRecorder()

	ImproveGigHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGenerateFAQ_Success(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"title":       "AI Video Generation",
		"description": "Professional videos with Sora 2",
		"category":    "ai-video",
	})
	req := authRequest("POST", "/api/v1/ai/gig/faq", body, token)
	rr := newRecorder()

	GenerateFAQHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	faq := result["faq"].([]interface{})
	if len(faq) < 3 {
		t.Fatalf("expected at least 3 FAQ entries, got %d", len(faq))
	}

	first := faq[0].(map[string]interface{})
	if first["question"] == nil || first["answer"] == nil {
		t.Error("expected question and answer in FAQ entry")
	}
}

func TestGenerateFAQ_MissingTitle(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{"description": "test"})
	req := authRequest("POST", "/api/v1/ai/gig/faq", body, token)
	rr := newRecorder()

	GenerateFAQHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGeneratePackages_Success(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"serviceType": "AI Chatbot Development",
		"priceMin":    200,
		"priceMax":    1000,
	})
	req := authRequest("POST", "/api/v1/ai/gig/packages", body, token)
	rr := newRecorder()

	GeneratePackagesHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	packages := result["packages"].([]interface{})
	if len(packages) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(packages))
	}

	basic := packages[0].(map[string]interface{})
	if basic["tier"] != "basic" {
		t.Errorf("expected basic tier, got %v", basic["tier"])
	}
	if basic["price"].(float64) != 200 {
		t.Errorf("expected basic price 200, got %v", basic["price"])
	}
}

func TestGeneratePackages_MissingServiceType(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{"priceMin": 100})
	req := authRequest("POST", "/api/v1/ai/gig/packages", body, token)
	rr := newRecorder()

	GeneratePackagesHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSEOSuggestions_Success(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"title":    "AI Video Production",
		"category": "ai-video",
	})
	req := authRequest("POST", "/api/v1/ai/gig/seo", body, token)
	rr := newRecorder()

	SEOSuggestionsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	tips := result["tips"].([]interface{})
	if len(tips) < 5 {
		t.Fatalf("expected at least 5 SEO tips, got %d", len(tips))
	}
}

func TestSEOSuggestions_MissingTitle(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{"category": "test"})
	req := authRequest("POST", "/api/v1/ai/gig/seo", body, token)
	rr := newRecorder()

	SEOSuggestionsHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
