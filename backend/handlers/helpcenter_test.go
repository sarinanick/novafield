package handlers

import (
	"encoding/json"
	"testing"
)

func TestListHelpArticles(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/help/articles", nil, "")
	rr := newRecorder()

	ListHelpArticlesHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var articles []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&articles)
	if len(articles) < 5 {
		t.Errorf("expected at least 5 articles, got %d", len(articles))
	}
}

func TestGetHelpArticle(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/help/articles/help-1", nil, "")
	rr := newRecorder()

	GetHelpArticleHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["id"] != "help-1" {
		t.Errorf("expected help-1, got %v", result["id"])
	}
}

func TestGetHelpArticle_NotFound(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/help/articles/nonexistent", nil, "")
	rr := newRecorder()

	GetHelpArticleHandler(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestSearchHelp(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/help/search?q=payment", nil, "")
	rr := newRecorder()

	SearchHelpHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var results []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&results)
	if len(results) == 0 {
		t.Error("expected search results for 'payment'")
	}
}

func TestSearchHelp_NoResults(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/help/search?q=xyznonexistent", nil, "")
	rr := newRecorder()

	SearchHelpHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var results []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&results)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestCreateTicket(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"subject":  "Can't place order",
		"category": "orders",
		"priority": "high",
		"content":  "I keep getting an error when trying to place an order",
	})
	req := authRequest("POST", "/api/v1/support/tickets", body, token)
	rr := newRecorder()

	CreateTicketHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateTicket_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"subject": "test", "content": "test"})
	req := authRequest("POST", "/api/v1/support/tickets", body, "")
	rr := newRecorder()

	CreateTicketHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestListMyTickets(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"subject": "Test", "content": "Help me"})
	req := authRequest("POST", "/api/v1/support/tickets", body, token)
	rr := newRecorder()
	CreateTicketHandler(rr, req)

	listReq := authRequest("GET", "/api/v1/support/tickets", nil, token)
	listRR := newRecorder()
	ListMyTicketsHandler(listRR, listReq)

	if listRR.Code != 200 {
		t.Fatalf("expected 200, got %d", listRR.Code)
	}

	var tickets []map[string]interface{}
	json.NewDecoder(listRR.Body).Decode(&tickets)
	if len(tickets) != 1 {
		t.Errorf("expected 1 ticket, got %d", len(tickets))
	}
}

func TestAdminListTickets(t *testing.T) {
	resetDB()
	_, clientToken := createTestUser("client")
	_, adminToken := createTestUser("admin")

	body := jsonBody(map[string]interface{}{"subject": "Test", "content": "Help me"})
	req := authRequest("POST", "/api/v1/support/tickets", body, clientToken)
	rr := newRecorder()
	CreateTicketHandler(rr, req)

	adminReq := authRequest("GET", "/api/v1/admin/support/tickets", nil, adminToken)
	adminRR := newRecorder()
	AdminListTicketsHandler(adminRR, adminReq)

	if adminRR.Code != 200 {
		t.Fatalf("expected 200, got %d", adminRR.Code)
	}
}

func TestAdminUpdateTicket(t *testing.T) {
	resetDB()
	_, clientToken := createTestUser("client")
	_, adminToken := createTestUser("admin")

	body := jsonBody(map[string]interface{}{"subject": "Test", "content": "Help me"})
	req := authRequest("POST", "/api/v1/support/tickets", body, clientToken)
	rr := newRecorder()
	CreateTicketHandler(rr, req)

	ticketID := decodeJSON(rr)["id"].(string)

	updateBody := jsonBody(map[string]interface{}{
		"status": "in_progress",
		"reply":  "We're looking into this issue",
	})
	updateReq := authRequest("PUT", "/api/v1/admin/support/tickets/"+ticketID, updateBody, adminToken)
	updateRR := newRecorder()
	AdminUpdateTicketHandler(updateRR, updateReq)

	if updateRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", updateRR.Code, updateRR.Body.String())
	}
}

func TestAddWorkspaceComment(t *testing.T) {
	resetDB()
	buyer, buyerToken := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	body := jsonBody(map[string]interface{}{"content": "Great progress on the design!"})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/workspace/comments", body, buyerToken)
	rr := newRecorder()

	AddWorkspaceCommentHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAddWorkspaceTask(t *testing.T) {
	resetDB()
	buyer, buyerToken := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	body := jsonBody(map[string]interface{}{
		"title":       "Create mockups",
		"description": "Design the landing page mockups",
		"assignedTo":  seller.ID,
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/workspace/tasks", body, buyerToken)
	rr := newRecorder()

	AddWorkspaceTaskHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateWorkspaceTask(t *testing.T) {
	resetDB()
	buyer, buyerToken := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	taskBody := jsonBody(map[string]interface{}{"title": "Task 1", "assignedTo": seller.ID})
	taskReq := authRequest("POST", "/api/v1/orders/"+orderID+"/workspace/tasks", taskBody, buyerToken)
	taskRR := newRecorder()
	AddWorkspaceTaskHandler(taskRR, taskReq)

	taskID := decodeJSON(taskRR)["id"].(string)

	updateBody := jsonBody(map[string]interface{}{"status": "done"})
	updateReq := authRequest("PUT", "/api/v1/orders/"+orderID+"/workspace/tasks/"+taskID, updateBody, buyerToken)
	updateRR := newRecorder()
	UpdateWorkspaceTaskHandler(updateRR, updateReq)

	if updateRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", updateRR.Code, updateRR.Body.String())
	}
}
