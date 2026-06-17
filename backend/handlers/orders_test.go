package handlers

import (
	"encoding/json"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"testing"
)

func seedTestOrder(buyerID, sellerID, gigID, pkgID string) string {
	orderID := store.NewID()

	d := database.GetDB()
	d.Mu.Lock()
	d.Orders = append(d.Orders, models.Order{
		ID:        orderID,
		GigID:     gigID,
		PackageID: pkgID,
		BuyerID:   buyerID,
		SellerID:  sellerID,
		Status:    "active",
		Price:     50,
		EscrowStatus: "held",
		CreatedAt: store.Now(),
	})
	d.Mu.Unlock()

	return orderID
}

func TestCreateOrder_Success(t *testing.T) {
	resetDB()

	_, buyerToken := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)

	body := jsonBody(map[string]interface{}{
		"gigId":        gigID,
		"packageId":    pkgID,
		"requirements": "Do the thing",
	})
	req := authRequest("POST", "/api/v1/orders", body, buyerToken)
	rr := newRecorder()

	CreateOrderHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["id"] == nil {
		t.Fatal("expected order id")
	}
	if result["message"] != "Order placed" {
		t.Errorf("expected 'Order placed', got %v", result["message"])
	}
}

func TestCreateOrder_OwnGig(t *testing.T) {
	resetDB()

	seller, token := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)

	body := jsonBody(map[string]interface{}{
		"gigId":     gigID,
		"packageId": pkgID,
	})
	req := authRequest("POST", "/api/v1/orders", body, token)
	rr := newRecorder()

	CreateOrderHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["error"] != "Cannot order your own gig" {
		t.Errorf("expected own gig error, got %v", result["error"])
	}
}

func TestCreateOrder_InvalidPackage(t *testing.T) {
	resetDB()

	_, token := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, _ := seedTestGig(seller.ID)

	body := jsonBody(map[string]interface{}{
		"gigId":     gigID,
		"packageId": "nonexistent-pkg",
	})
	req := authRequest("POST", "/api/v1/orders", body, token)
	rr := newRecorder()

	CreateOrderHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateOrder_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{
		"gigId":     "x",
		"packageId": "y",
	})
	req := authRequest("POST", "/api/v1/orders", body, "")
	rr := newRecorder()

	CreateOrderHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestGetOrders_AsBuyer(t *testing.T) {
	resetDB()

	buyer, buyerToken := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	req := authRequest("GET", "/api/v1/orders", nil, buyerToken)
	rr := newRecorder()

	GetOrdersHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var orders []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&orders)
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
}

func TestGetOrders_AsSeller(t *testing.T) {
	resetDB()

	buyer, _ := createTestUser("client")
	seller, sellerToken := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	req := authRequest("GET", "/api/v1/orders", nil, sellerToken)
	rr := newRecorder()

	GetOrdersHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var orders []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&orders)
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
}

func TestDeliverOrder_AsSeller(t *testing.T) {
	resetDB()

	buyer, _ := createTestUser("client")
	seller, sellerToken := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	body := jsonBody(map[string]interface{}{
		"file":  "/uploads/delivery.zip",
		"notes": "Here you go",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/deliver", body, sellerToken)
	rr := newRecorder()

	DeliverOrderHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeliverOrder_AsBuyer_Forbidden(t *testing.T) {
	resetDB()

	buyer, buyerToken := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	body := jsonBody(map[string]interface{}{"notes": "nope"})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/deliver", body, buyerToken)
	rr := newRecorder()

	DeliverOrderHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestApproveOrder_AsBuyer(t *testing.T) {
	resetDB()

	buyer, buyerToken := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	req := authRequest("POST", "/api/v1/orders/"+orderID+"/approve", nil, buyerToken)
	rr := newRecorder()

	ApproveOrderHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, o := range d.Orders {
		if o.ID == orderID {
			if o.Status != "completed" {
				t.Errorf("expected status completed, got %s", o.Status)
			}
			if o.EscrowStatus != "released" {
				t.Errorf("expected escrow released, got %s", o.EscrowStatus)
			}
			break
		}
	}
	d.Mu.RUnlock()
}

func TestApproveOrder_AsSeller_Forbidden(t *testing.T) {
	resetDB()

	buyer, _ := createTestUser("client")
	seller, sellerToken := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	req := authRequest("POST", "/api/v1/orders/"+orderID+"/approve", nil, sellerToken)
	rr := newRecorder()

	ApproveOrderHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestRequestRevision_AsBuyer(t *testing.T) {
	resetDB()

	buyer, buyerToken := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	body := jsonBody(map[string]interface{}{
		"message": "Please revise the colors",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/revision", body, buyerToken)
	rr := newRecorder()

	RequestRevisionHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, o := range d.Orders {
		if o.ID == orderID {
			if o.Status != "revision" {
				t.Errorf("expected status revision, got %s", o.Status)
			}
			break
		}
	}
	d.Mu.RUnlock()
}

func TestCreateReview_AfterCompletion(t *testing.T) {
	resetDB()

	buyer, buyerToken := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Orders {
		if d.Orders[i].ID == orderID {
			d.Orders[i].Status = "completed"
			break
		}
	}
	d.Mu.Unlock()

	body := jsonBody(map[string]interface{}{
		"rating":  5,
		"comment": "Amazing work!",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/review", body, buyerToken)
	rr := newRecorder()

	CreateReviewHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["id"] == nil {
		t.Fatal("expected review id")
	}
}

func TestCreateReview_InvalidRating(t *testing.T) {
	resetDB()

	buyer, buyerToken := createTestUser("client")
	seller, _ := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)
	orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)

	body := jsonBody(map[string]interface{}{
		"rating":  0,
		"comment": "Bad rating",
	})
	req := authRequest("POST", "/api/v1/orders/"+orderID+"/review", body, buyerToken)
	rr := newRecorder()

	CreateReviewHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestFullOrderLifecycle(t *testing.T) {
	resetDB()

	_, buyerToken := createTestUser("client")
	seller, sellerToken := createTestUser("freelancer")
	gigID, pkgID := seedTestGig(seller.ID)

	body := jsonBody(map[string]interface{}{
		"gigId":        gigID,
		"packageId":    pkgID,
		"requirements": "Build me something cool",
	})
	req := authRequest("POST", "/api/v1/orders", body, buyerToken)
	rr := newRecorder()
	CreateOrderHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("create order: expected 201, got %d", rr.Code)
	}
	orderResult := decodeJSON(rr)
	orderID := orderResult["id"].(string)

	deliverBody := jsonBody(map[string]interface{}{
		"file":  "/uploads/result.zip",
		"notes": "Done!",
	})
	req2 := authRequest("POST", "/api/v1/orders/"+orderID+"/deliver", deliverBody, sellerToken)
	rr2 := newRecorder()
	DeliverOrderHandler(rr2, req2)

	if rr2.Code != 200 {
		t.Fatalf("deliver: expected 200, got %d", rr2.Code)
	}

	req3 := authRequest("POST", "/api/v1/orders/"+orderID+"/approve", nil, buyerToken)
	rr3 := newRecorder()
	ApproveOrderHandler(rr3, req3)

	if rr3.Code != 200 {
		t.Fatalf("approve: expected 200, got %d", rr3.Code)
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, o := range d.Orders {
		if o.ID == orderID {
			if o.Status != "completed" {
				t.Errorf("final status: expected completed, got %s", o.Status)
			}
			break
		}
	}
	d.Mu.RUnlock()

	reviewBody := jsonBody(map[string]interface{}{
		"rating":  5,
		"comment": "Excellent!",
	})
	req4 := authRequest("POST", "/api/v1/orders/"+orderID+"/review", reviewBody, buyerToken)
	rr4 := newRecorder()
	CreateReviewHandler(rr4, req4)

	if rr4.Code != 201 {
		t.Fatalf("review: expected 201, got %d", rr4.Code)
	}
}
