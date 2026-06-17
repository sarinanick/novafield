package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func AddMilestoneHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	orderID := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	orderID = strings.TrimSuffix(orderID, "/milestones")

	order := store.FindOrderByID(orderID)
	if order == nil {
		Error(w, 404, "Order not found")
		return
	}
	if order.BuyerID != user.ID && order.SellerID != user.ID {
		Error(w, 403, "Not authorized")
		return
	}

	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		DueDate     string  `json:"dueDate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.Price <= 0 {
		Error(w, 400, "Title and positive price are required")
		return
	}

	milestone := models.Milestone{
		ID:          store.NewID(),
		OrderID:     orderID,
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Status:      "pending",
		DueDate:     req.DueDate,
		CreatedAt:   store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.Milestones = append(d.Milestones, milestone)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": milestone.ID, "message": "Milestone added"})
}

func ListMilestonesHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	orderID := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	orderID = strings.TrimSuffix(orderID, "/milestones")

	order := store.FindOrderByID(orderID)
	if order == nil {
		Error(w, 404, "Order not found")
		return
	}
	if order.BuyerID != user.ID && order.SellerID != user.ID {
		Error(w, 403, "Not authorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var milestones []models.Milestone
	for _, m := range d.Milestones {
		if m.OrderID == orderID {
			milestones = append(milestones, m)
		}
	}
	d.Mu.RUnlock()

	if milestones == nil {
		milestones = []models.Milestone{}
	}
	JSON(w, 200, milestones)
}

func UpdateMilestoneHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		Error(w, 400, "Invalid path")
		return
	}
	orderID := parts[0]
	milestoneID := parts[2]

	order := store.FindOrderByID(orderID)
	if order == nil || (order.BuyerID != user.ID && order.SellerID != user.ID) {
		Error(w, 403, "Not authorized")
		return
	}

	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		DueDate     string  `json:"dueDate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Milestones {
		if d.Milestones[i].ID == milestoneID && d.Milestones[i].OrderID == orderID {
			if d.Milestones[i].Status != "pending" {
				d.Mu.Unlock()
				Error(w, 400, "Can only edit pending milestones")
				return
			}
			if req.Title != "" {
				d.Milestones[i].Title = req.Title
			}
			if req.Description != "" {
				d.Milestones[i].Description = req.Description
			}
			if req.Price > 0 {
				d.Milestones[i].Price = req.Price
			}
			if req.DueDate != "" {
				d.Milestones[i].DueDate = req.DueDate
			}
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Milestone updated"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Milestone not found")
}

func CompleteMilestoneHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	path = strings.TrimSuffix(path, "/complete")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		Error(w, 400, "Invalid path")
		return
	}
	orderID := parts[0]
	milestoneID := parts[2]

	order := store.FindOrderByID(orderID)
	if order == nil || order.SellerID != user.ID {
		Error(w, 403, "Only the seller can complete milestones")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Milestones {
		if d.Milestones[i].ID == milestoneID && d.Milestones[i].OrderID == orderID {
			if d.Milestones[i].Status != "pending" && d.Milestones[i].Status != "in_progress" {
				d.Mu.Unlock()
				Error(w, 400, "Milestone cannot be completed")
				return
			}
			d.Milestones[i].Status = "completed"
			d.Milestones[i].CompletedAt = store.Now()
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Milestone marked as completed"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Milestone not found")
}

func ApproveMilestoneHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	path = strings.TrimSuffix(path, "/approve")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		Error(w, 400, "Invalid path")
		return
	}
	orderID := parts[0]
	milestoneID := parts[2]

	order := store.FindOrderByID(orderID)
	if order == nil || order.BuyerID != user.ID {
		Error(w, 403, "Only the buyer can approve milestones")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Milestones {
		if d.Milestones[i].ID == milestoneID && d.Milestones[i].OrderID == orderID {
			if d.Milestones[i].Status != "completed" {
				d.Mu.Unlock()
				Error(w, 400, "Milestone must be completed before approval")
				return
			}
			d.Milestones[i].Status = "approved"
			d.Milestones[i].ApprovedAt = store.Now()

			for j := range d.Users {
				if d.Users[j].ID == order.SellerID {
					d.Users[j].Earnings += d.Milestones[i].Price
					break
				}
			}
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Milestone approved, payment released"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Milestone not found")
}
