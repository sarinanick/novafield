package handlers

import (
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"sort"
	"strings"
)

func GetDashboardHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var stats models.DashboardStats

	d := database.GetDB()
	d.Mu.RLock()

	var orders []models.Order
	for _, o := range d.Orders {
		if o.BuyerID == user.ID || o.SellerID == user.ID {
			orders = append(orders, o)
		}
	}

	for _, o := range orders {
		if user.Role == "freelancer" {
			if o.SellerID == user.ID {
				stats.TotalOrders++
				if o.Status == "active" || o.Status == "revision" {
					stats.ActiveOrders++
				}
				if o.Status == "completed" {
					stats.CompletedOrders++
					stats.TotalEarnings += o.Price
				}
				if o.Status == "active" || o.Status == "delivered" {
					stats.PendingPayouts += o.Price
				}
			}
		} else {
			if o.BuyerID == user.ID {
				stats.TotalOrders++
				if o.Status == "active" || o.Status == "revision" {
					stats.ActiveOrders++
				}
				if o.Status == "completed" {
					stats.CompletedOrders++
					stats.TotalSpent += o.Price
				}
			}
		}
	}

	stats.AvgRating = user.Rating
	stats.TotalReviews = user.ReviewsCount

	var gigCount int
	var totalViews int
	for _, g := range d.Gigs {
		if g.FreelancerID == user.ID {
			gigCount++
			totalViews += g.Views
		}
	}
	stats.TotalGigs = gigCount
	stats.TotalViews = totalViews

	if stats.TotalOrders > 0 {
		stats.ConversionRate = float64(stats.CompletedOrders) / float64(stats.TotalOrders)
	}

	var recentOrders []models.Order
	for _, o := range d.Orders {
		if user.Role == "freelancer" {
			if o.SellerID == user.ID {
				recentOrders = append(recentOrders, o)
			}
		} else {
			if o.BuyerID == user.ID {
				recentOrders = append(recentOrders, o)
			}
		}
	}
	d.Mu.RUnlock()

	sort.Slice(recentOrders, func(i, j int) bool {
		return recentOrders[i].CreatedAt > recentOrders[j].CreatedAt
	})
	if len(recentOrders) > 10 {
		recentOrders = recentOrders[:10]
	}

	for i := range recentOrders {
		gig := store.FindGigByID(recentOrders[i].GigID)
		if gig != nil {
			recentOrders[i].Gig = gig
		}
		buyer := store.FindUserByID(recentOrders[i].BuyerID)
		if buyer != nil {
			pub := store.ToPublic(*buyer)
			recentOrders[i].Buyer = &pub
		}
		seller := store.FindUserByID(recentOrders[i].SellerID)
		if seller != nil {
			pub := store.ToPublic(*seller)
			recentOrders[i].Seller = &pub
		}
	}
	stats.RecentOrders = recentOrders
	JSON(w, 200, stats)
}

func GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var notifs []models.Notification
	for _, n := range d.Notifications {
		if n.UserID == user.ID {
			notifs = append(notifs, n)
		}
	}
	d.Mu.RUnlock()

	sort.Slice(notifs, func(i, j int) bool {
		return notifs[i].CreatedAt > notifs[j].CreatedAt
	})
	if len(notifs) > 50 {
		notifs = notifs[:50]
	}

	if notifs == nil {
		notifs = []models.Notification{}
	}
	JSON(w, 200, notifs)
}

func MarkNotificationsReadHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Notifications {
		if d.Notifications[i].UserID == user.ID {
			d.Notifications[i].IsRead = true
		}
	}
	d.Mu.Unlock()
	d.Save()
	JSON(w, 200, H{"message": "Notifications marked as read"})
}

func ToggleFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}
	gigID := strings.TrimPrefix(r.URL.Path, "/api/v1/favorites/")

	store.DB.Mu.Lock()
	defer store.DB.Mu.Unlock()

	if store.DB.Favorites[user.ID] == nil {
		store.DB.Favorites[user.ID] = make(map[string]bool)
	}

	if store.DB.Favorites[user.ID][gigID] {
		delete(store.DB.Favorites[user.ID], gigID)
		JSON(w, 200, H{"favorited": false})
		return
	}
	store.DB.Favorites[user.ID][gigID] = true
	JSON(w, 200, H{"favorited": true})
}

func GetEarningsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var totalEarnings float64
	var pendingPayouts float64
	var completedOrders int

	d := database.GetDB()
	d.Mu.RLock()
	for _, o := range d.Orders {
		if o.SellerID == user.ID {
			if o.Status == "completed" {
				totalEarnings += o.Price
				completedOrders++
			}
			if o.Status == "active" || o.Status == "delivered" {
				pendingPayouts += o.Price
			}
		}
	}
	d.Mu.RUnlock()

	JSON(w, 200, H{
		"totalEarnings":   totalEarnings,
		"pendingPayouts":  pendingPayouts,
		"completedOrders": completedOrders,
	})
}

func GetSpendingHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var totalSpent float64
	var activeOrders int
	var completedOrders int

	d := database.GetDB()
	d.Mu.RLock()
	for _, o := range d.Orders {
		if o.BuyerID == user.ID {
			if o.Status == "completed" {
				totalSpent += o.Price
				completedOrders++
			}
			if o.Status == "active" || o.Status == "revision" {
				activeOrders++
			}
		}
	}
	d.Mu.RUnlock()

	JSON(w, 200, H{
		"totalSpent":      totalSpent,
		"activeOrders":    activeOrders,
		"completedOrders": completedOrders,
	})
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	JSON(w, 200, H{"status": "ok", "service": "novafield-api", "version": "1.0.0"})
}
