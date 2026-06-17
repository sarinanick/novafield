package handlers

import (
	"encoding/json"
	"novafield-api/database"
	"novafield-api/models"
	"testing"
)

func seedAnalyticsData() (*models.User, string, *models.User, string) {
	seller, sellerToken := createTestUser("freelancer")
	buyer, buyerToken := createTestUser("client")
	gigID, pkgID := seedTestGig(seller.ID)

	for i := 0; i < 3; i++ {
		orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)
		d := database.GetDB()
		d.Mu.Lock()
		for j := range d.Orders {
			if d.Orders[j].ID == orderID {
				d.Orders[j].Status = "completed"
				d.Orders[j].Price = 100
				break
			}
		}
		d.Mu.Unlock()
	}

	return seller, sellerToken, buyer, buyerToken
}

func TestAnalyticsEarnings(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedAnalyticsData()

	req := authRequest("GET", "/api/v1/analytics/earnings", nil, sellerToken)
	rr := newRecorder()

	AnalyticsEarningsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["totalEarnings"].(float64) != 300 {
		t.Errorf("expected totalEarnings 300, got %v", result["totalEarnings"])
	}
	if result["completedOrders"].(float64) != 3 {
		t.Errorf("expected completedOrders 3, got %v", result["completedOrders"])
	}

	_ = seller
}

func TestAnalyticsEarnings_Unauthorized(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/analytics/earnings", nil, "")
	rr := newRecorder()

	AnalyticsEarningsHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAnalyticsOrders(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedAnalyticsData()

	req := authRequest("GET", "/api/v1/analytics/orders", nil, sellerToken)
	rr := newRecorder()

	AnalyticsOrdersHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["totalOrders"].(float64) != 3 {
		t.Errorf("expected 3 orders, got %v", result["totalOrders"])
	}

	_ = seller
}

func TestAnalyticsGigs(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedAnalyticsData()

	req := authRequest("GET", "/api/v1/analytics/gigs", nil, sellerToken)
	rr := newRecorder()

	AnalyticsGigsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var gigs []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&gigs)
	if len(gigs) != 1 {
		t.Fatalf("expected 1 gig, got %d", len(gigs))
	}

	_ = seller
}

func TestAnalyticsProfile(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedAnalyticsData()

	req := authRequest("GET", "/api/v1/analytics/profile", nil, sellerToken)
	rr := newRecorder()

	AnalyticsProfileHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["totalGigs"].(float64) != 1 {
		t.Errorf("expected 1 gig, got %v", result["totalGigs"])
	}

	_ = seller
}

func TestAnalyticsClients(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedAnalyticsData()

	req := authRequest("GET", "/api/v1/analytics/clients", nil, sellerToken)
	rr := newRecorder()

	AnalyticsClientsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	clients := result["clients"].([]interface{})
	if len(clients) != 1 {
		t.Errorf("expected 1 client, got %d", len(clients))
	}

	_ = seller
}

func TestAnalyticsInsights(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedAnalyticsData()

	req := authRequest("GET", "/api/v1/analytics/insights", nil, sellerToken)
	rr := newRecorder()

	AnalyticsInsightsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	insights := result["insights"].([]interface{})
	if len(insights) == 0 {
		t.Error("expected at least 1 insight")
	}

	_ = seller
}

func TestAnalyticsExport(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedAnalyticsData()

	req := authRequest("GET", "/api/v1/analytics/export", nil, sellerToken)
	rr := newRecorder()

	AnalyticsExportHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if len(body) == 0 {
		t.Fatal("expected CSV content")
	}

	_ = seller
}

func TestAnalyticsExport_Unauthorized(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/analytics/export", nil, "")
	rr := newRecorder()

	AnalyticsExportHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
