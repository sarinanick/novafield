package handlers

import (
	"encoding/json"
	"novafield-api/database"
	"novafield-api/models"
	"testing"
)

func seedInvoiceData() (*models.User, string, *models.User, string) {
	seller, sellerToken := createTestUser("freelancer")
	buyer, buyerToken := createTestUser("client")
	gigID, pkgID := seedTestGig(seller.ID)

	for i := 0; i < 2; i++ {
		orderID := seedTestOrder(buyer.ID, seller.ID, gigID, pkgID)
		d := database.GetDB()
		d.Mu.Lock()
		for j := range d.Orders {
			if d.Orders[j].ID == orderID {
				d.Orders[j].Status = "completed"
				d.Orders[j].Price = 150
				break
			}
		}
		d.Mu.Unlock()
	}

	return seller, sellerToken, buyer, buyerToken
}

func TestFinancialSummary_AsSeller(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedInvoiceData()

	req := authRequest("GET", "/api/v1/financials/summary", nil, sellerToken)
	rr := newRecorder()

	FinancialSummaryHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["totalEarned"].(float64) <= 0 {
		t.Error("expected totalEarned > 0")
	}
	if result["earnedOrders"].(float64) != 2 {
		t.Errorf("expected 2 earned orders, got %v", result["earnedOrders"])
	}

	_ = seller
}

func TestFinancialSummary_AsBuyer(t *testing.T) {
	resetDB()
	_, _, buyer, buyerToken := seedInvoiceData()

	req := authRequest("GET", "/api/v1/financials/summary", nil, buyerToken)
	rr := newRecorder()

	FinancialSummaryHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["totalSpent"].(float64) <= 0 {
		t.Error("expected totalSpent > 0")
	}

	_ = buyer
}

func TestFinancialSummary_Unauthorized(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/financials/summary", nil, "")
	rr := newRecorder()

	FinancialSummaryHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestFinancialByCategory(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedInvoiceData()

	req := authRequest("GET", "/api/v1/financials/by-category", nil, sellerToken)
	rr := newRecorder()

	FinancialByCategoryHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&result)
	if len(result) == 0 {
		t.Error("expected at least 1 category")
	}

	_ = seller
}

func TestFinancialByClient(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedInvoiceData()

	req := authRequest("GET", "/api/v1/financials/by-client", nil, sellerToken)
	rr := newRecorder()

	FinancialByClientHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&result)
	if len(result) != 1 {
		t.Errorf("expected 1 client, got %d", len(result))
	}

	_ = seller
}

func TestFinancialByClient_Client_Forbidden(t *testing.T) {
	resetDB()
	_, _, _, buyerToken := seedInvoiceData()

	req := authRequest("GET", "/api/v1/financials/by-client", nil, buyerToken)
	rr := newRecorder()

	FinancialByClientHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestFinancialExport(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedInvoiceData()

	req := authRequest("GET", "/api/v1/financials/export", nil, sellerToken)
	rr := newRecorder()

	FinancialExportHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if len(body) == 0 {
		t.Fatal("expected CSV content")
	}

	header := "OrderID,Role,Status,Price,Fee,Net,Date\n"
	if body[:len(header)] != header {
		t.Errorf("expected CSV header, got %s", body[:len(header)])
	}

	_ = seller
}

func TestFinancialTaxSummary(t *testing.T) {
	resetDB()
	seller, sellerToken, _, _ := seedInvoiceData()

	req := authRequest("GET", "/api/v1/financials/tax-summary", nil, sellerToken)
	rr := newRecorder()

	FinancialTaxSummaryHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	summaries := result["summaries"].([]interface{})
	if len(summaries) == 0 {
		t.Error("expected at least 1 year summary")
	}

	_ = seller
}

func TestListInvoices_Empty(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("GET", "/api/v1/invoices", nil, token)
	rr := newRecorder()

	ListInvoicesHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var invoices []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&invoices)
	if len(invoices) != 0 {
		t.Errorf("expected 0 invoices, got %d", len(invoices))
	}
}

func TestGetInvoice_NotFound(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("GET", "/api/v1/invoices/nonexistent", nil, token)
	rr := newRecorder()

	GetInvoiceHandler(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestSeedInvoiceData_CreatesOrders(t *testing.T) {
	resetDB()
	seller, _, _, _ := seedInvoiceData()

	d := database.GetDB()
	d.Mu.RLock()
	completed := 0
	for _, o := range d.Orders {
		if o.SellerID == seller.ID && o.Status == "completed" {
			completed++
		}
	}
	d.Mu.RUnlock()

	if completed != 2 {
		t.Errorf("expected 2 completed orders, got %d", completed)
	}

	_ = seller
}
