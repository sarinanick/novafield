package handlers

import (
	"encoding/json"
	"novafield-api/database"
	"novafield-api/models"
	"testing"
)

func seedMilestoneOrder() (*models.User, string, *models.User, string, string) {
	buyer, buyerToken := createTestUser("client")
	seller, sellerToken := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)
	return buyer, buyerToken, seller, sellerToken, orderID
}

func TestAddMilestone(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedMilestoneOrder()

	body := jsonBody(map[string]interface{}{
		"title":       "Design Phase",
		"description": "Create UI mockups",
		"price":       100,
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones", body, buyerToken)
	rr := newRecorder()

	AddMilestoneHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAddMilestone_Unauthorized(t *testing.T) {
	resetDB()
	_, _, _, _, orderID := seedMilestoneOrder()

	body := jsonBody(map[string]interface{}{"title": "test", "price": 100})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones", body, "")
	rr := newRecorder()

	AddMilestoneHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestListMilestones(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedMilestoneOrder()

	body := jsonBody(map[string]interface{}{"title": "M1", "price": 50})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones", body, buyerToken)
	rr := newRecorder()
	AddMilestoneHandler(rr, req)

	listReq := authRequest("GET", "/api/v1/orders/"+orderID+"/milestones", nil, buyerToken)
	listRR := newRecorder()
	ListMilestonesHandler(listRR, listReq)

	if listRR.Code != 200 {
		t.Fatalf("expected 200, got %d", listRR.Code)
	}

	var milestones []map[string]interface{}
	json.NewDecoder(listRR.Body).Decode(&milestones)
	if len(milestones) != 1 {
		t.Errorf("expected 1 milestone, got %d", len(milestones))
	}
}

func TestCompleteMilestone_AsSeller(t *testing.T) {
	resetDB()
	_, buyerToken, _, sellerToken, orderID := seedMilestoneOrder()

	body := jsonBody(map[string]interface{}{"title": "M1", "price": 100})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones", body, buyerToken)
	rr := newRecorder()
	AddMilestoneHandler(rr, req)

	d := database.GetDB()
	d.Mu.RLock()
	var milestoneID string
	for _, m := range d.Milestones {
		if m.OrderID == orderID {
			milestoneID = m.ID
			break
		}
	}
	d.Mu.RUnlock()

	completeReq := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones/"+milestoneID+"/complete", nil, sellerToken)
	completeRR := newRecorder()
	CompleteMilestoneHandler(completeRR, completeReq)

	if completeRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", completeRR.Code, completeRR.Body.String())
	}
}

func TestCompleteMilestone_AsBuyer_Forbidden(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedMilestoneOrder()

	body := jsonBody(map[string]interface{}{"title": "M1", "price": 100})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones", body, buyerToken)
	rr := newRecorder()
	AddMilestoneHandler(rr, req)

	d := database.GetDB()
	d.Mu.RLock()
	var milestoneID string
	for _, m := range d.Milestones {
		if m.OrderID == orderID {
			milestoneID = m.ID
			break
		}
	}
	d.Mu.RUnlock()

	completeReq := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones/"+milestoneID+"/complete", nil, buyerToken)
	completeRR := newRecorder()
	CompleteMilestoneHandler(completeRR, completeReq)

	if completeRR.Code != 403 {
		t.Fatalf("expected 403, got %d", completeRR.Code)
	}
}

func TestApproveMilestone_AsBuyer(t *testing.T) {
	resetDB()
	_, buyerToken, _, sellerToken, orderID := seedMilestoneOrder()

	body := jsonBody(map[string]interface{}{"title": "M1", "price": 100})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones", body, buyerToken)
	rr := newRecorder()
	AddMilestoneHandler(rr, req)

	d := database.GetDB()
	d.Mu.RLock()
	var milestoneID string
	for _, m := range d.Milestones {
		if m.OrderID == orderID {
			milestoneID = m.ID
			break
		}
	}
	d.Mu.RUnlock()

	completeReq := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones/"+milestoneID+"/complete", nil, sellerToken)
	completeRR := newRecorder()
	CompleteMilestoneHandler(completeRR, completeReq)

	approveReq := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones/"+milestoneID+"/approve", nil, buyerToken)
	approveRR := newRecorder()
	ApproveMilestoneHandler(approveRR, approveReq)

	if approveRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", approveRR.Code, approveRR.Body.String())
	}

	d.Mu.RLock()
	for _, m := range d.Milestones {
		if m.ID == milestoneID {
			if m.Status != "approved" {
				t.Errorf("expected approved, got %s", m.Status)
			}
			break
		}
	}
	d.Mu.RUnlock()
}

func TestMilestone_FullLifecycle(t *testing.T) {
	resetDB()
	_, buyerToken, _, sellerToken, orderID := seedMilestoneOrder()

	body := jsonBody(map[string]interface{}{
		"title": "Development",
		"price": 200,
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones", body, buyerToken)
	rr := newRecorder()
	AddMilestoneHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("add: expected 201, got %d", rr.Code)
	}

	d := database.GetDB()
	d.Mu.RLock()
	var milestoneID string
	for _, m := range d.Milestones {
		if m.OrderID == orderID {
			milestoneID = m.ID
			break
		}
	}
	d.Mu.RUnlock()

	completeReq := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones/"+milestoneID+"/complete", nil, sellerToken)
	completeRR := newRecorder()
	CompleteMilestoneHandler(completeRR, completeReq)

	if completeRR.Code != 200 {
		t.Fatalf("complete: expected 200, got %d", completeRR.Code)
	}

	approveReq := authRequest("POST", "/api/v1/orders/"+orderID+"/milestones/"+milestoneID+"/approve", nil, buyerToken)
	approveRR := newRecorder()
	ApproveMilestoneHandler(approveRR, approveReq)

	if approveRR.Code != 200 {
		t.Fatalf("approve: expected 200, got %d", approveRR.Code)
	}
}
