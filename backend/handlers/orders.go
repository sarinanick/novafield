package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		GigID        string `json:"gigId"`
		PackageID    string `json:"packageId"`
		Requirements string `json:"requirements"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	pkg := store.FindPackageByID(req.PackageID)
	gig := store.FindGigByID(req.GigID)
	if pkg == nil || gig == nil || pkg.GigID != req.GigID {
		Error(w, 400, "Invalid package or gig")
		return
	}
	if gig.FreelancerID == user.ID {
		Error(w, 400, "Cannot order your own gig")
		return
	}

	orderID := store.NewID()
	order := models.Order{
		ID: orderID, GigID: req.GigID, PackageID: req.PackageID,
		BuyerID: user.ID, SellerID: gig.FreelancerID, Status: "active",
		Requirements: req.Requirements, Price: pkg.Price, EscrowStatus: "held",
		CreatedAt: store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()

	d.Orders = append(d.Orders, order)

	for i := range d.Gigs {
		if d.Gigs[i].ID == req.GigID {
			d.Gigs[i].OrdersCount++
			break
		}
	}

	for i := range d.Users {
		if d.Users[i].ID == user.ID {
			d.Users[i].Spent += pkg.Price
			break
		}
	}

	orderNotification := models.Notification{
		ID: store.NewID(), UserID: gig.FreelancerID, Type: "order",
		Title: "New Order", Message: "You have a new order",
		Link: "/orders/" + orderID, CreatedAt: store.Now(),
	}
	d.Notifications = append(d.Notifications, orderNotification)

	d.Mu.Unlock()
	d.Save()

	BroadcastNotification(gig.FreelancerID, orderNotification)

	JSON(w, 201, H{"id": orderID, "message": "Order placed"})
}

func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	statusFilter := r.URL.Query().Get("status")

	d := database.GetDB()
	d.Mu.RLock()
	var orders []models.Order
	for _, o := range d.Orders {
		if o.BuyerID == user.ID || o.SellerID == user.ID {
			if statusFilter == "" || o.Status == statusFilter {
				orders = append(orders, o)
			}
		}
	}
	d.Mu.RUnlock()

	for i := range orders {
		gig := store.FindGigByID(orders[i].GigID)
		if gig != nil {
			orders[i].Gig = gig
		}
		buyer := store.FindUserByID(orders[i].BuyerID)
		if buyer != nil {
			pub := store.ToPublic(*buyer)
			orders[i].Buyer = &pub
		}
		seller := store.FindUserByID(orders[i].SellerID)
		if seller != nil {
			pub := store.ToPublic(*seller)
			orders[i].Seller = &pub
		}
	}
	if orders == nil {
		orders = []models.Order{}
	}
	JSON(w, 200, orders)
}

func DeliverOrderHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	id = strings.TrimSuffix(id, "/deliver")

	order := store.FindOrderByID(id)
	if order == nil || order.SellerID != user.ID {
		Error(w, 403, "Not authorized")
		return
	}

	var req struct {
		File  string `json:"file"`
		Notes string `json:"notes"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Orders {
		if d.Orders[i].ID == id {
			d.Orders[i].Status = "delivered"
			d.Orders[i].DeliveryFile = req.File
			d.Orders[i].DeliveryNotes = req.Notes
			d.Orders[i].DeliveredAt = store.Now()
			break
		}
	}

	deliveryNotification := models.Notification{
		ID: store.NewID(), UserID: order.BuyerID, Type: "delivery",
		Title: "Delivery Ready", Message: "Your order has been delivered!",
		Link: "/orders/" + id, CreatedAt: store.Now(),
	}
	d.Notifications = append(d.Notifications, deliveryNotification)

	d.Mu.Unlock()
	d.Save()

	BroadcastNotification(order.BuyerID, deliveryNotification)

	JSON(w, 200, H{"message": "Order delivered"})
}

func ApproveOrderHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	id = strings.TrimSuffix(id, "/approve")

	order := store.FindOrderByID(id)
	if order == nil || order.BuyerID != user.ID {
		Error(w, 403, "Not authorized")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Orders {
		if d.Orders[i].ID == id {
			d.Orders[i].Status = "completed"
			d.Orders[i].EscrowStatus = "released"
			d.Orders[i].CompletedAt = store.Now()
			break
		}
	}

	for i := range d.Users {
		if d.Users[i].ID == order.SellerID {
			d.Users[i].Earnings += order.Price
			break
		}
	}

	paymentNotification := models.Notification{
		ID: store.NewID(), UserID: order.SellerID, Type: "payment",
		Title: "Payment Released", Message: "Payment has been released to your account",
		Link: "/orders/" + id, CreatedAt: store.Now(),
	}
	d.Notifications = append(d.Notifications, paymentNotification)

	d.Mu.Unlock()
	d.Save()

	BroadcastNotification(order.SellerID, paymentNotification)

	JSON(w, 200, H{"message": "Order completed, payment released"})
}

func RequestRevisionHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	id = strings.TrimSuffix(id, "/revision")

	order := store.FindOrderByID(id)
	if order == nil || order.BuyerID != user.ID {
		Error(w, 403, "Not authorized")
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Orders {
		if d.Orders[i].ID == id {
			d.Orders[i].Status = "revision"
			d.Orders[i].DeliveryNotes = req.Message
			break
		}
	}

	revisionNotification := models.Notification{
		ID: store.NewID(), UserID: order.SellerID, Type: "revision",
		Title: "Revision Requested", Message: req.Message,
		Link: "/orders/" + id, CreatedAt: store.Now(),
	}
	d.Notifications = append(d.Notifications, revisionNotification)

	d.Mu.Unlock()
	d.Save()

	BroadcastNotification(order.SellerID, revisionNotification)

	JSON(w, 200, H{"message": "Revision requested"})
}

func CreateReviewHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	id = strings.TrimSuffix(id, "/review")

	order := store.FindOrderByID(id)
	if order == nil || (order.BuyerID != user.ID && order.SellerID != user.ID) {
		Error(w, 403, "Not authorized")
		return
	}

	var req struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Rating < 1 || req.Rating > 5 {
		Error(w, 400, "Rating must be 1-5")
		return
	}

	revieweeID := order.SellerID
	if user.ID == order.SellerID {
		revieweeID = order.BuyerID
	}

	review := models.Review{
		ID: store.NewID(), OrderID: id, GigID: order.GigID,
		ReviewerID: user.ID, RevieweeID: revieweeID,
		Rating: req.Rating, Comment: req.Comment, CreatedAt: store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.Reviews = append(d.Reviews, review)

	var sum float64
	var count int
	for _, r := range d.Reviews {
		if r.GigID == order.GigID {
			sum += float64(r.Rating)
			count++
		}
	}
	if count > 0 {
		for i := range d.Gigs {
			if d.Gigs[i].ID == order.GigID {
				d.Gigs[i].Rating = math.Round(sum/float64(count)*10) / 10
				break
			}
		}
	}
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": review.ID, "message": "Review submitted"})
}
