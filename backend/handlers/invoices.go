package handlers

import (
	"fmt"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func ListInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var invoices []models.Invoice
	for _, inv := range d.Invoices {
		if inv.IssuedTo == user.ID || inv.IssuedBy == user.ID {
			invoices = append(invoices, inv)
		}
	}
	d.Mu.RUnlock()

	if invoices == nil {
		invoices = []models.Invoice{}
	}
	JSON(w, 200, invoices)
}

func GetInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	invoiceID := strings.TrimPrefix(r.URL.Path, "/api/v1/invoices/")

	d := database.GetDB()
	d.Mu.RLock()
	var invoice *models.Invoice
	for i := range d.Invoices {
		if d.Invoices[i].ID == invoiceID {
			invoice = &d.Invoices[i]
			break
		}
	}
	d.Mu.RUnlock()

	if invoice == nil {
		Error(w, 404, "Invoice not found")
		return
	}
	if invoice.IssuedTo != user.ID && invoice.IssuedBy != user.ID && user.Role != "admin" {
		Error(w, 403, "Not authorized")
		return
	}

	JSON(w, 200, invoice)
}

func FinancialSummaryHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	var totalEarned, totalSpent, platformFees float64
	var earnedOrders, spentOrders int

	for _, o := range d.Orders {
		if o.Status == "completed" {
			fee := o.Price * 0.1
			if o.SellerID == user.ID {
				totalEarned += o.Price - fee
				platformFees += fee
				earnedOrders++
			}
			if o.BuyerID == user.ID {
				totalSpent += o.Price
				spentOrders++
			}
		}
	}
	d.Mu.RUnlock()

	JSON(w, 200, H{
		"totalEarned":  totalEarned,
		"totalSpent":   totalSpent,
		"platformFees": platformFees,
		"netIncome":    totalEarned - totalSpent,
		"earnedOrders": earnedOrders,
		"spentOrders":  spentOrders,
	})
}

func FinancialByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	categoryData := make(map[string]map[string]float64)
	for _, o := range d.Orders {
		if o.Status == "completed" {
			gig := store.FindGigByID(o.GigID)
			if gig == nil {
				continue
			}
			cat := gig.Category
			if categoryData[cat] == nil {
				categoryData[cat] = map[string]float64{"earned": 0, "spent": 0, "orders": 0}
			}
			if o.SellerID == user.ID {
				categoryData[cat]["earned"] += o.Price
				categoryData[cat]["orders"]++
			}
			if o.BuyerID == user.ID {
				categoryData[cat]["spent"] += o.Price
			}
		}
	}
	d.Mu.RUnlock()

	var result []map[string]interface{}
	for cat, data := range categoryData {
		result = append(result, map[string]interface{}{
			"category": cat,
			"earned":   data["earned"],
			"spent":    data["spent"],
			"orders":   data["orders"],
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}

	JSON(w, 200, result)
}

func FinancialByClientHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "freelancer" {
		Error(w, 403, "Only freelancers can view client breakdown")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	clientData := make(map[string]map[string]float64)
	for _, o := range d.Orders {
		if o.SellerID == user.ID && o.Status == "completed" {
			if clientData[o.BuyerID] == nil {
				clientData[o.BuyerID] = map[string]float64{"revenue": 0, "orders": 0}
			}
			clientData[o.BuyerID]["revenue"] += o.Price
			clientData[o.BuyerID]["orders"]++
		}
	}
	d.Mu.RUnlock()

	var result []map[string]interface{}
	for clientID, data := range clientData {
		client := store.FindUserByID(clientID)
		name := "Unknown"
		if client != nil {
			name = client.Name
		}
		result = append(result, map[string]interface{}{
			"clientId": clientID,
			"name":     name,
			"revenue":  data["revenue"],
			"orders":   data["orders"],
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}

	JSON(w, 200, result)
}

func FinancialExportHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	csv := "OrderID,Role,Status,Price,Fee,Net,Date\n"
	for _, o := range d.Orders {
		if o.SellerID == user.ID || o.BuyerID == user.ID {
			role := "buyer"
			net := o.Price
			fee := 0.0
			if o.SellerID == user.ID {
				role = "seller"
				fee = o.Price * 0.1
				net = o.Price - fee
			}
			csv += fmt.Sprintf("%s,%s,%s,%.2f,%.2f,%.2f,%s\n",
				o.ID, role, o.Status, o.Price, fee, net, o.CreatedAt)
		}
	}
	d.Mu.RUnlock()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=financial_report.csv")
	w.Write([]byte(csv))
}

func FinancialTaxSummaryHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	yearlyData := make(map[string]map[string]float64)
	for _, o := range d.Orders {
		if o.Status == "completed" {
			year := ""
			if len(o.CompletedAt) >= 4 {
				year = o.CompletedAt[:4]
			} else if len(o.CreatedAt) >= 4 {
				year = o.CreatedAt[:4]
			}
			if year == "" {
				continue
			}
			if yearlyData[year] == nil {
				yearlyData[year] = map[string]float64{"gross": 0, "fees": 0, "net": 0, "orders": 0}
			}
			fee := o.Price * 0.1
			if o.SellerID == user.ID {
				yearlyData[year]["gross"] += o.Price
				yearlyData[year]["fees"] += fee
				yearlyData[year]["net"] += o.Price - fee
				yearlyData[year]["orders"]++
			}
		}
	}
	d.Mu.RUnlock()

	var summaries []map[string]interface{}
	for year, data := range yearlyData {
		summaries = append(summaries, map[string]interface{}{
			"year":    year,
			"gross":   data["gross"],
			"fees":    data["fees"],
			"net":     data["net"],
			"orders":  data["orders"],
		})
	}
	if summaries == nil {
		summaries = []map[string]interface{}{}
	}

	JSON(w, 200, H{"summaries": summaries})
}
