package handlers

import (
	"fmt"
	"net/http"
	"novafield-api/database"
	"novafield-api/store"
)

func AnalyticsEarningsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	var totalEarnings, pendingPayouts float64
	var completedOrders, totalOrders int
	monthlyData := make(map[string]float64)

	for _, o := range d.Orders {
		if o.SellerID == user.ID {
			totalOrders++
			if o.Status == "completed" {
				completedOrders++
				totalEarnings += o.Price
				if len(o.CompletedAt) >= 7 {
					month := o.CompletedAt[:7]
					monthlyData[month] += o.Price
				}
			}
			if o.Status == "active" || o.Status == "delivered" {
				pendingPayouts += o.Price
			}
		}
	}
	d.Mu.RUnlock()

	var monthly []map[string]interface{}
	for month, amount := range monthlyData {
		monthly = append(monthly, map[string]interface{}{
			"month":  month,
			"amount": amount,
		})
	}

	JSON(w, 200, H{
		"totalEarnings":   totalEarnings,
		"pendingPayouts":  pendingPayouts,
		"completedOrders": completedOrders,
		"totalOrders":     totalOrders,
		"monthly":         monthly,
	})
}

func AnalyticsOrdersHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	var totalOrders, completed, active, delivered, revision int
	var totalRevenue float64

	for _, o := range d.Orders {
		if o.SellerID == user.ID {
			totalOrders++
			switch o.Status {
			case "completed":
				completed++
				totalRevenue += o.Price
			case "active":
				active++
			case "delivered":
				delivered++
			case "revision":
				revision++
			}
		}
	}
	d.Mu.RUnlock()

	conversionRate := 0.0
	if totalOrders > 0 {
		conversionRate = float64(completed) / float64(totalOrders)
	}

	avgOrderValue := 0.0
	if completed > 0 {
		avgOrderValue = totalRevenue / float64(completed)
	}

	JSON(w, 200, H{
		"totalOrders":    totalOrders,
		"completed":      completed,
		"active":         active,
		"delivered":      delivered,
		"revision":       revision,
		"conversionRate": fmt.Sprintf("%.1f%%", conversionRate*100),
		"avgOrderValue":  avgOrderValue,
		"totalRevenue":   totalRevenue,
	})
}

func AnalyticsGigsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	type gigPerf struct {
		ID             string  `json:"id"`
		Title          string  `json:"title"`
		Views          int     `json:"views"`
		Orders         int     `json:"orders"`
		Revenue        float64 `json:"revenue"`
		ConversionRate float64 `json:"conversionRate"`
		Rating         float64 `json:"rating"`
	}

	var gigs []gigPerf
	for _, g := range d.Gigs {
		if g.FreelancerID == user.ID {
			orders := 0
			revenue := 0.0
			for _, o := range d.Orders {
				if o.GigID == g.ID && o.Status == "completed" {
					orders++
					revenue += o.Price
				}
			}
			convRate := 0.0
			if g.Views > 0 {
				convRate = float64(orders) / float64(g.Views)
			}
			gigs = append(gigs, gigPerf{
				ID: g.ID, Title: g.Title, Views: g.Views,
				Orders: orders, Revenue: revenue,
				ConversionRate: convRate, Rating: g.Rating,
			})
		}
	}
	d.Mu.RUnlock()

	if gigs == nil {
		gigs = []gigPerf{}
	}
	JSON(w, 200, gigs)
}

func AnalyticsProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	var gigCount, totalViews int
	for _, g := range d.Gigs {
		if g.FreelancerID == user.ID {
			gigCount++
			totalViews += g.Views
		}
	}

	var activeClients int
	clientSet := make(map[string]bool)
	for _, o := range d.Orders {
		if o.SellerID == user.ID && o.Status == "completed" {
			if !clientSet[o.BuyerID] {
				clientSet[o.BuyerID] = true
				activeClients++
			}
		}
	}
	d.Mu.RUnlock()

	JSON(w, 200, H{
		"profileViews":  totalViews,
		"totalGigs":     gigCount,
		"activeClients": activeClients,
		"rating":        user.Rating,
		"reviewsCount":  user.ReviewsCount,
	})
}

func AnalyticsClientsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	type clientStat struct {
		ClientID    string  `json:"clientId"`
		ClientName  string  `json:"clientName"`
		OrderCount  int     `json:"orderCount"`
		TotalSpent  float64 `json:"totalSpent"`
		IsRepeat    bool    `json:"isRepeat"`
	}

	clientMap := make(map[string]*clientStat)
	for _, o := range d.Orders {
		if o.SellerID == user.ID && o.Status == "completed" {
			if cs, ok := clientMap[o.BuyerID]; ok {
				cs.OrderCount++
				cs.TotalSpent += o.Price
				cs.IsRepeat = true
			} else {
				buyer := store.FindUserByID(o.BuyerID)
				name := "Unknown"
				if buyer != nil {
					name = buyer.Name
				}
				clientMap[o.BuyerID] = &clientStat{
					ClientID: o.BuyerID, ClientName: name,
					OrderCount: 1, TotalSpent: o.Price,
				}
			}
		}
	}
	d.Mu.RUnlock()

	var clients []clientStat
	repeatCount := 0
	for _, cs := range clientMap {
		clients = append(clients, *cs)
		if cs.IsRepeat {
			repeatCount++
		}
	}

	retentionRate := 0.0
	if len(clientMap) > 0 {
		retentionRate = float64(repeatCount) / float64(len(clientMap))
	}

	if clients == nil {
		clients = []clientStat{}
	}

	JSON(w, 200, H{
		"clients":        clients,
		"totalClients":   len(clientMap),
		"repeatClients":  repeatCount,
		"retentionRate":  fmt.Sprintf("%.1f%%", retentionRate*100),
	})
}

func AnalyticsInsightsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	var insights []map[string]interface{}

	var totalOrders, completed int
	var totalEarnings float64
	for _, o := range d.Orders {
		if o.SellerID == user.ID {
			totalOrders++
			if o.Status == "completed" {
				completed++
				totalEarnings += o.Price
			}
		}
	}

	gigCount := 0
	for _, g := range d.Gigs {
		if g.FreelancerID == user.ID {
			gigCount++
		}
	}

	if totalOrders > 0 {
		rate := float64(completed) / float64(totalOrders) * 100
		if rate >= 90 {
			insights = append(insights, map[string]interface{}{
				"type": "achievement", "title": "Excellent completion rate!",
				"detail": fmt.Sprintf("%.0f%% of your orders are completed successfully", rate),
				"action": "Keep up the great work",
			})
		} else if rate < 70 {
			insights = append(insights, map[string]interface{}{
				"type": "warning", "title": "Completion rate needs attention",
				"detail": fmt.Sprintf("%.0f%% completion rate is below the 80%% target", rate),
				"action": "Review your delivery process and set realistic timelines",
			})
		}
	}

	if gigCount < 3 {
		insights = append(insights, map[string]interface{}{
			"type": "opportunity", "title": "Create more gig listings",
			"detail": "Freelancers with 3+ gigs get 2x more orders",
			"action": "Create additional gigs in related categories",
		})
	}

	if user.Rating >= 4.5 && user.ReviewsCount >= 10 {
		insights = append(insights, map[string]interface{}{
			"type": "achievement", "title": "Top-rated freelancer!",
			"detail": fmt.Sprintf("Rating %.1f with %d reviews", user.Rating, user.ReviewsCount),
			"action": "Consider raising your rates - you've earned it",
		})
	}

	if totalEarnings > 0 && totalOrders > 0 {
		avgOrder := totalEarnings / float64(completed)
		if avgOrder < 50 {
			insights = append(insights, map[string]interface{}{
				"type": "opportunity", "title": "Consider premium packages",
				"detail":  fmt.Sprintf("Your average order value is $%.0f", avgOrder),
				"action": "Add higher-tier packages to increase revenue per order",
			})
		}
	}

	d.Mu.RUnlock()

	if insights == nil {
		insights = []map[string]interface{}{}
	}

	JSON(w, 200, H{"insights": insights})
}

func AnalyticsExportHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	csv := "ID,Gig,Status,Price,Date,Role\n"
	for _, o := range d.Orders {
		if o.SellerID == user.ID || o.BuyerID == user.ID {
			role := "buyer"
			if o.SellerID == user.ID {
				role = "seller"
			}
			csv += fmt.Sprintf("%s,%s,%s,%.2f,%s,%s\n", o.ID, o.GigID, o.Status, o.Price, o.CreatedAt, role)
		}
	}
	d.Mu.RUnlock()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=orders_export.csv")
	w.Write([]byte(csv))
}
