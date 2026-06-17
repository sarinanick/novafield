package handlers

import (
	"encoding/json"
	"novafield-api/database"
	"testing"
)

func TestCreateBrief(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"title":       "Need AI chatbot",
		"description": "Build a customer support chatbot",
		"category":    "ai-chatbots",
		"skills":      []string{"GPT-4", "LangChain"},
		"budgetMin":   500,
		"budgetMax":   2000,
		"timeline":    "1_month",
	})
	req := authRequest("POST", "/api/v1/briefs", body, token)
	rr := newRecorder()

	CreateBriefHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateBrief_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"title": "test"})
	req := authRequest("POST", "/api/v1/briefs", body, "")
	rr := newRecorder()

	CreateBriefHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestListBriefs(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"title": "Brief 1"})
	req := authRequest("POST", "/api/v1/briefs", body, token)
	rr := newRecorder()
	CreateBriefHandler(rr, req)

	listReq := authRequest("GET", "/api/v1/briefs", nil, token)
	listRR := newRecorder()
	ListBriefsHandler(listRR, listReq)

	if listRR.Code != 200 {
		t.Fatalf("expected 200, got %d", listRR.Code)
	}

	var briefs []map[string]interface{}
	json.NewDecoder(listRR.Body).Decode(&briefs)
	if len(briefs) != 1 {
		t.Errorf("expected 1 brief, got %d", len(briefs))
	}
}

func TestGetBrief(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"title": "Test Brief"})
	req := authRequest("POST", "/api/v1/briefs", body, token)
	rr := newRecorder()
	CreateBriefHandler(rr, req)

	briefID := decodeJSON(rr)["id"].(string)

	getReq := authRequest("GET", "/api/v1/briefs/"+briefID, nil, token)
	getRR := newRecorder()
	GetBriefHandler(getRR, getReq)

	if getRR.Code != 200 {
		t.Fatalf("expected 200, got %d", getRR.Code)
	}
}

func TestSubmitProposal(t *testing.T) {
	resetDB()
	client, clientToken := createTestUser("client")
	_, freelancerToken := createTestUser("freelancer")

	briefBody := jsonBody(map[string]interface{}{"title": "Need help"})
	briefReq := authRequest("POST", "/api/v1/briefs", briefBody, clientToken)
	briefRR := newRecorder()
	CreateBriefHandler(briefRR, briefReq)

	briefID := decodeJSON(briefRR)["id"].(string)

	body := jsonBody(map[string]interface{}{
		"coverLetter": "I can do this!",
		"price":       1000,
		"timeline":    "2 weeks",
	})
	req := authRequest("POST", "/api/v1/briefs/"+briefID+"/proposals", body, freelancerToken)
	rr := newRecorder()
	SubmitProposalHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	_ = client
}

func TestSubmitProposal_Client_Forbidden(t *testing.T) {
	resetDB()
	_, clientToken := createTestUser("client")

	briefBody := jsonBody(map[string]interface{}{"title": "Need help"})
	briefReq := authRequest("POST", "/api/v1/briefs", briefBody, clientToken)
	briefRR := newRecorder()
	CreateBriefHandler(briefRR, briefReq)

	briefID := decodeJSON(briefRR)["id"].(string)

	body := jsonBody(map[string]interface{}{"coverLetter": "test", "price": 100})
	req := authRequest("POST", "/api/v1/briefs/"+briefID+"/proposals", body, clientToken)
	rr := newRecorder()
	SubmitProposalHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestListProposals_AsClient(t *testing.T) {
	resetDB()
	_, clientToken := createTestUser("client")
	_, freelancerToken := createTestUser("freelancer")

	briefBody := jsonBody(map[string]interface{}{"title": "Need help"})
	briefReq := authRequest("POST", "/api/v1/briefs", briefBody, clientToken)
	briefRR := newRecorder()
	CreateBriefHandler(briefRR, briefReq)

	briefID := decodeJSON(briefRR)["id"].(string)

	proposalBody := jsonBody(map[string]interface{}{"coverLetter": "I can help", "price": 500})
	proposalReq := authRequest("POST", "/api/v1/briefs/"+briefID+"/proposals", proposalBody, freelancerToken)
	proposalRR := newRecorder()
	SubmitProposalHandler(proposalRR, proposalReq)

	listReq := authRequest("GET", "/api/v1/briefs/"+briefID+"/proposals", nil, clientToken)
	listRR := newRecorder()
	ListProposalsHandler(listRR, listReq)

	if listRR.Code != 200 {
		t.Fatalf("expected 200, got %d", listRR.Code)
	}

	var proposals []map[string]interface{}
	json.NewDecoder(listRR.Body).Decode(&proposals)
	if len(proposals) != 1 {
		t.Errorf("expected 1 proposal, got %d", len(proposals))
	}
}

func TestAcceptProposal(t *testing.T) {
	resetDB()
	_, clientToken := createTestUser("client")
	freelancer, freelancerToken := createTestUser("freelancer")

	briefBody := jsonBody(map[string]interface{}{"title": "Need help"})
	briefReq := authRequest("POST", "/api/v1/briefs", briefBody, clientToken)
	briefRR := newRecorder()
	CreateBriefHandler(briefRR, briefReq)

	briefID := decodeJSON(briefRR)["id"].(string)

	proposalBody := jsonBody(map[string]interface{}{"coverLetter": "I can help", "price": 500})
	proposalReq := authRequest("POST", "/api/v1/briefs/"+briefID+"/proposals", proposalBody, freelancerToken)
	proposalRR := newRecorder()
	SubmitProposalHandler(proposalRR, proposalReq)

	proposalID := decodeJSON(proposalRR)["id"].(string)

	acceptReq := authRequest("POST", "/api/v1/briefs/"+briefID+"/proposals/"+proposalID+"/accept", nil, clientToken)
	acceptRR := newRecorder()
	AcceptProposalHandler(acceptRR, acceptReq)

	if acceptRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", acceptRR.Code, acceptRR.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, b := range d.ClientBriefs {
		if b.ID == briefID {
			if b.Status != "in_progress" {
				t.Errorf("expected in_progress, got %s", b.Status)
			}
			break
		}
	}
	d.Mu.RUnlock()

	_ = freelancer
}

func TestListMyBriefs(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"title": "My Brief"})
	req := authRequest("POST", "/api/v1/briefs", body, token)
	rr := newRecorder()
	CreateBriefHandler(rr, req)

	myReq := authRequest("GET", "/api/v1/my-briefs", nil, token)
	myRR := newRecorder()
	ListMyBriefsHandler(myRR, myReq)

	if myRR.Code != 200 {
		t.Fatalf("expected 200, got %d", myRR.Code)
	}

	var briefs []map[string]interface{}
	json.NewDecoder(myRR.Body).Decode(&briefs)
	if len(briefs) != 1 {
		t.Errorf("expected 1 brief, got %d", len(briefs))
	}
}

func TestBrief_FullLifecycle(t *testing.T) {
	resetDB()
	client, clientToken := createTestUser("client")
	freelancer, freelancerToken := createTestUser("freelancer")

	briefBody := jsonBody(map[string]interface{}{
		"title":       "AI Chatbot Project",
		"description": "Build a chatbot",
		"skills":      []string{"GPT-4"},
		"budgetMin":   500,
		"budgetMax":   1500,
	})
	briefReq := authRequest("POST", "/api/v1/briefs", briefBody, clientToken)
	briefRR := newRecorder()
	CreateBriefHandler(briefRR, briefReq)

	if briefRR.Code != 201 {
		t.Fatalf("create brief: expected 201, got %d", briefRR.Code)
	}
	briefID := decodeJSON(briefRR)["id"].(string)

	proposalBody := jsonBody(map[string]interface{}{
		"coverLetter": "I'm expert in chatbots",
		"price":       1000,
		"timeline":    "3 weeks",
	})
	proposalReq := authRequest("POST", "/api/v1/briefs/"+briefID+"/proposals", proposalBody, freelancerToken)
	proposalRR := newRecorder()
	SubmitProposalHandler(proposalRR, proposalReq)

	if proposalRR.Code != 201 {
		t.Fatalf("submit proposal: expected 201, got %d", proposalRR.Code)
	}
	proposalID := decodeJSON(proposalRR)["id"].(string)

	acceptReq := authRequest("POST", "/api/v1/briefs/"+briefID+"/proposals/"+proposalID+"/accept", nil, clientToken)
	acceptRR := newRecorder()
	AcceptProposalHandler(acceptRR, acceptReq)

	if acceptRR.Code != 200 {
		t.Fatalf("accept: expected 200, got %d", acceptRR.Code)
	}

	_ = client
	_ = freelancer
}
