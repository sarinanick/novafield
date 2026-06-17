package handlers

import (
	"encoding/json"
	"novafield-api/database"
	"novafield-api/models"
	"testing"
)

func seedDeliveredOrder() (*models.User, string, *models.User, string, string) {
	buyer, buyerToken := createTestUser("client")
	seller, sellerToken := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Orders {
		if d.Orders[i].ID == orderID {
			d.Orders[i].Status = "delivered"
			break
		}
	}
	d.Mu.Unlock()

	return buyer, buyerToken, seller, sellerToken, orderID
}

func TestOpenDispute_AsBuyer(t *testing.T) {
	resetDB()
	buyer, buyerToken, _, _, orderID := seedDeliveredOrder()

	body := jsonBody(map[string]interface{}{
		"reason":      "quality",
		"description": "The deliverable does not match the requirements",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/dispute", body, buyerToken)
	rr := newRecorder()

	OpenDisputeHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["id"] == nil {
		t.Fatal("expected dispute id")
	}

	_ = buyer
}

func TestOpenDispute_AsSeller(t *testing.T) {
	resetDB()
	_, _, _, sellerToken, orderID := seedDeliveredOrder()

	body := jsonBody(map[string]interface{}{
		"reason":      "scope",
		"description": "Client changed requirements after delivery",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/dispute", body, sellerToken)
	rr := newRecorder()

	OpenDisputeHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOpenDispute_Unauthorized(t *testing.T) {
	resetDB()
	_, _, _, _, orderID := seedDeliveredOrder()

	body := jsonBody(map[string]interface{}{
		"reason":      "quality",
		"description": "test",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/dispute", body, "")
	rr := newRecorder()

	OpenDisputeHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestOpenDispute_NotYourOrder(t *testing.T) {
	resetDB()
	_, _, _, _, orderID := seedDeliveredOrder()
	_, outsiderToken := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"reason":      "quality",
		"description": "test",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/dispute", body, outsiderToken)
	rr := newRecorder()

	OpenDisputeHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestOpenDispute_Duplicate(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedDeliveredOrder()

	body := jsonBody(map[string]interface{}{
		"reason":      "quality",
		"description": "First dispute",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/dispute", body, buyerToken)
	rr := newRecorder()
	OpenDisputeHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("first dispute: expected 201, got %d", rr.Code)
	}

	body2 := jsonBody(map[string]interface{}{
		"reason":      "quality",
		"description": "Second dispute",
	})
	req2 := authRequest("POST", "/api/v1/orders/"+orderID+"/dispute", body2, buyerToken)
	rr2 := newRecorder()
	OpenDisputeHandler(rr2, req2)

	if rr2.Code != 409 {
		t.Fatalf("duplicate: expected 409, got %d", rr2.Code)
	}
}

func TestOpenDispute_MissingDescription(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedDeliveredOrder()

	body := jsonBody(map[string]interface{}{
		"reason": "quality",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/dispute", body, buyerToken)
	rr := newRecorder()

	OpenDisputeHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSubmitEvidence(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedDeliveredOrder()

	disputeID := openTestDispute(orderID, buyerToken)

	body := jsonBody(map[string]interface{}{
		"type":    "text",
		"content": "Here is my evidence: the deliverable is incomplete",
	})
	req := authRequest("POST", "/api/v1/disputes/"+disputeID+"/evidence", body, buyerToken)
	rr := newRecorder()

	SubmitEvidenceHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSubmitEvidence_Unauthorized(t *testing.T) {
	resetDB()

	req := authRequest("POST", "/api/v1/disputes/fake/evidence", jsonBody(map[string]interface{}{
		"content": "test",
	}), "")
	rr := newRecorder()

	SubmitEvidenceHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestGetDispute(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedDeliveredOrder()

	disputeID := openTestDispute(orderID, buyerToken)

	req := authRequest("GET", "/api/v1/disputes/"+disputeID, nil, buyerToken)
	rr := newRecorder()

	GetDisputeHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	dispute := result["dispute"].(map[string]interface{})
	if dispute["id"] != disputeID {
		t.Errorf("expected dispute id %s, got %v", disputeID, dispute["id"])
	}
	if dispute["status"] != "open" {
		t.Errorf("expected status open, got %v", dispute["status"])
	}
}

func TestGetDispute_NotFound(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("GET", "/api/v1/disputes/nonexistent", nil, token)
	rr := newRecorder()

	GetDisputeHandler(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestListDisputes(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedDeliveredOrder()
	openTestDispute(orderID, buyerToken)

	req := authRequest("GET", "/api/v1/disputes", nil, buyerToken)
	rr := newRecorder()

	ListDisputesHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var disputes []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&disputes)
	if len(disputes) != 1 {
		t.Fatalf("expected 1 dispute, got %d", len(disputes))
	}
}

func TestResolveDispute_FullRefund(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedDeliveredOrder()
	disputeID := openTestDispute(orderID, buyerToken)

	admin, adminToken := createTestUser("admin")

	body := jsonBody(map[string]interface{}{
		"ruling":    "full_refund",
		"adminNote": "Work was not delivered as described",
	})
	req := authRequest("POST", "/api/v1/disputes/"+disputeID+"/resolve", body, adminToken)
	rr := newRecorder()

	ResolveDisputeHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, dispute := range d.Disputes {
		if dispute.ID == disputeID {
			if dispute.Status != "resolved" {
				t.Errorf("expected status resolved, got %s", dispute.Status)
			}
			if dispute.Resolution.Ruling != "full_refund" {
				t.Errorf("expected ruling full_refund, got %s", dispute.Resolution.Ruling)
			}
			break
		}
	}
	for _, o := range d.Orders {
		if o.ID == orderID {
			if o.Status != "refunded" {
				t.Errorf("expected order status refunded, got %s", o.Status)
			}
			break
		}
	}
	d.Mu.RUnlock()

	_ = admin
}

func TestResolveDispute_FullRelease(t *testing.T) {
	resetDB()
	_, buyerToken, seller, _, orderID := seedDeliveredOrder()
	disputeID := openTestDispute(orderID, buyerToken)

	_, adminToken := createTestUser("admin")

	body := jsonBody(map[string]interface{}{
		"ruling":    "full_release",
		"adminNote": "Freelancer delivered correctly",
	})
	req := authRequest("POST", "/api/v1/disputes/"+disputeID+"/resolve", body, adminToken)
	rr := newRecorder()

	ResolveDisputeHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, o := range d.Orders {
		if o.ID == orderID {
			if o.Status != "completed" {
				t.Errorf("expected order status completed, got %s", o.Status)
			}
			break
		}
	}
	for _, u := range d.Users {
		if u.ID == seller.ID {
			if u.Earnings <= 0 {
				t.Error("expected seller earnings > 0")
			}
			break
		}
	}
	d.Mu.RUnlock()
}

func TestResolveDispute_Split(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedDeliveredOrder()
	disputeID := openTestDispute(orderID, buyerToken)

	_, adminToken := createTestUser("admin")

	body := jsonBody(map[string]interface{}{
		"ruling":      "split",
		"splitClient": 40,
		"adminNote":   "Partial delivery",
	})
	req := authRequest("POST", "/api/v1/disputes/"+disputeID+"/resolve", body, adminToken)
	rr := newRecorder()

	ResolveDisputeHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, o := range d.Orders {
		if o.ID == orderID {
			if o.EscrowStatus != "split" {
				t.Errorf("expected escrow split, got %s", o.EscrowStatus)
			}
			break
		}
	}
	d.Mu.RUnlock()
}

func TestResolveDispute_NonAdmin(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedDeliveredOrder()
	disputeID := openTestDispute(orderID, buyerToken)

	body := jsonBody(map[string]interface{}{
		"ruling": "full_refund",
	})
	req := authRequest("POST", "/api/v1/disputes/"+disputeID+"/resolve", body, buyerToken)
	rr := newRecorder()

	ResolveDisputeHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestAcceptResolution(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedDeliveredOrder()
	disputeID := openTestDispute(orderID, buyerToken)

	_, adminToken := createTestUser("admin")

	resolveBody := jsonBody(map[string]interface{}{
		"ruling":    "full_refund",
		"adminNote": "Resolved",
	})
	resolveReq := authRequest("POST", "/api/v1/disputes/"+disputeID+"/resolve", resolveBody, adminToken)
	resolveRR := newRecorder()
	ResolveDisputeHandler(resolveRR, resolveReq)

	if resolveRR.Code != 200 {
		t.Fatalf("resolve: expected 200, got %d", resolveRR.Code)
	}

	acceptReq := authRequest("POST", "/api/v1/disputes/"+disputeID+"/accept", nil, buyerToken)
	acceptRR := newRecorder()
	AcceptResolutionHandler(acceptRR, acceptReq)

	if acceptRR.Code != 200 {
		t.Fatalf("accept: expected 200, got %d", acceptRR.Code)
	}
}

func TestAcceptResolution_NotResolved(t *testing.T) {
	resetDB()
	_, buyerToken, _, _, orderID := seedDeliveredOrder()
	disputeID := openTestDispute(orderID, buyerToken)

	req := authRequest("POST", "/api/v1/disputes/"+disputeID+"/accept", nil, buyerToken)
	rr := newRecorder()

	AcceptResolutionHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDispute_FullLifecycle(t *testing.T) {
	resetDB()
	_, buyerToken, _, sellerToken, orderID := seedDeliveredOrder()

	disputeID := openTestDispute(orderID, buyerToken)

	evBody := jsonBody(map[string]interface{}{
		"type":    "text",
		"content": "The video quality is 720p, not 1080p as specified",
	})
	evReq := authRequest("POST", "/api/v1/disputes/"+disputeID+"/evidence", evBody, buyerToken)
	evRR := newRecorder()
	SubmitEvidenceHandler(evRR, evReq)

	if evRR.Code != 201 {
		t.Fatalf("buyer evidence: expected 201, got %d", evRR.Code)
	}

	evBody2 := jsonBody(map[string]interface{}{
		"type":    "text",
		"content": "The requirements said 720p is acceptable for the basic package",
	})
	evReq2 := authRequest("POST", "/api/v1/disputes/"+disputeID+"/evidence", evBody2, sellerToken)
	evRR2 := newRecorder()
	SubmitEvidenceHandler(evRR2, evReq2)

	if evRR2.Code != 201 {
		t.Fatalf("seller evidence: expected 201, got %d", evRR2.Code)
	}

	_, adminToken := createTestUser("admin")

	resolveBody := jsonBody(map[string]interface{}{
		"ruling":      "split",
		"splitClient": 30,
		"adminNote":   "Package description was ambiguous",
	})
	resolveReq := authRequest("POST", "/api/v1/disputes/"+disputeID+"/resolve", resolveBody, adminToken)
	resolveRR := newRecorder()
	ResolveDisputeHandler(resolveRR, resolveReq)

	if resolveRR.Code != 200 {
		t.Fatalf("resolve: expected 200, got %d: %s", resolveRR.Code, resolveRR.Body.String())
	}

	acceptReq := authRequest("POST", "/api/v1/disputes/"+disputeID+"/accept", nil, buyerToken)
	acceptRR := newRecorder()
	AcceptResolutionHandler(acceptRR, acceptReq)

	if acceptRR.Code != 200 {
		t.Fatalf("accept: expected 200, got %d", acceptRR.Code)
	}
}

func openTestDispute(orderID, token string) string {
	body := jsonBody(map[string]interface{}{
		"reason":      "quality",
		"description": "Test dispute for testing purposes",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/dispute", body, token)
	rr := newRecorder()
	OpenDisputeHandler(rr, req)

	result := decodeJSON(rr)
	return result["id"].(string)
}
