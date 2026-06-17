package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
	"time"
)

func CreateSubscriptionPlanHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || (user.Role != "freelancer" && user.Role != "admin") {
		Error(w, 403, "Only freelancers can create subscription plans")
		return
	}

	gigID := strings.TrimPrefix(r.URL.Path, "/api/v1/gigs/")
	gigID = strings.TrimSuffix(gigID, "/subscriptions")

	gig := store.FindGigByID(gigID)
	if gig == nil {
		Error(w, 404, "Gig not found")
		return
	}
	if gig.FreelancerID != user.ID {
		Error(w, 403, "Only the gig owner can create subscription plans")
		return
	}

	var req struct {
		Name         string  `json:"name"`
		Interval     string  `json:"interval"`
		Price        float64 `json:"price"`
		Deliverables string  `json:"deliverables"`
		MaxRevisions int     `json:"maxRevisions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.Price <= 0 {
		Error(w, 400, "Name and positive price are required")
		return
	}

	if req.Interval != "monthly" && req.Interval != "quarterly" {
		req.Interval = "monthly"
	}
	if req.MaxRevisions <= 0 {
		req.MaxRevisions = 2
	}

	plan := models.SubscriptionPlan{
		ID:           store.NewID(),
		GigID:        gigID,
		Name:         req.Name,
		Interval:     req.Interval,
		Price:        req.Price,
		Deliverables: req.Deliverables,
		MaxRevisions: req.MaxRevisions,
		IsActive:     true,
		CreatedAt:    store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.SubscriptionPlans = append(d.SubscriptionPlans, plan)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": plan.ID, "message": "Subscription plan created"})
}

func SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		PlanID string `json:"planId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlanID == "" {
		Error(w, 400, "Plan ID is required")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var plan *models.SubscriptionPlan
	for i := range d.SubscriptionPlans {
		if d.SubscriptionPlans[i].ID == req.PlanID {
			plan = &d.SubscriptionPlans[i]
			break
		}
	}
	d.Mu.RUnlock()

	if plan == nil || !plan.IsActive {
		Error(w, 404, "Plan not found or inactive")
		return
	}

	gig := store.FindGigByID(plan.GigID)
	if gig == nil {
		Error(w, 404, "Associated gig not found")
		return
	}
	if gig.FreelancerID == user.ID {
		Error(w, 400, "Cannot subscribe to your own plan")
		return
	}

	for _, sub := range d.Subscriptions {
		if sub.ClientID == user.ID && sub.PlanID == req.PlanID && sub.Status == "active" {
			Error(w, 409, "Already subscribed to this plan")
			return
		}
	}

	now := time.Now().UTC()
	var periodEnd time.Time
	if plan.Interval == "quarterly" {
		periodEnd = now.AddDate(0, 3, 0)
	} else {
		periodEnd = now.AddDate(0, 1, 0)
	}

	subID := store.NewID()
	sub := models.Subscription{
		ID:                 subID,
		PlanID:             req.PlanID,
		ClientID:           user.ID,
		FreelancerID:       gig.FreelancerID,
		Status:             "active",
		CurrentPeriodStart: now.Format(time.RFC3339),
		CurrentPeriodEnd:   periodEnd.Format(time.RFC3339),
		NextBillingDate:    periodEnd.Format(time.RFC3339),
		CreatedAt:          store.Now(),
	}

	d.Mu.Lock()
	d.Subscriptions = append(d.Subscriptions, sub)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": subID, "message": "Subscribed successfully"})
}

func ListSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var subs []models.Subscription
	for _, s := range d.Subscriptions {
		if s.ClientID == user.ID || s.FreelancerID == user.ID {
			subs = append(subs, s)
		}
	}
	d.Mu.RUnlock()

	if subs == nil {
		subs = []models.Subscription{}
	}
	JSON(w, 200, subs)
}

func GetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	subID := strings.TrimPrefix(r.URL.Path, "/api/v1/subscriptions/")

	d := database.GetDB()
	d.Mu.RLock()
	var sub *models.Subscription
	for i := range d.Subscriptions {
		if d.Subscriptions[i].ID == subID {
			sub = &d.Subscriptions[i]
			break
		}
	}
	d.Mu.RUnlock()

	if sub == nil {
		Error(w, 404, "Subscription not found")
		return
	}
	if sub.ClientID != user.ID && sub.FreelancerID != user.ID && user.Role != "admin" {
		Error(w, 403, "Not authorized")
		return
	}

	JSON(w, 200, sub)
}

func ChangePlanHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	subID := strings.TrimPrefix(r.URL.Path, "/api/v1/subscriptions/")
	subID = strings.TrimSuffix(subID, "/plan")

	var req struct {
		NewPlanID string `json:"newPlanId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewPlanID == "" {
		Error(w, 400, "New plan ID required")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	found := false
	for i := range d.Subscriptions {
		if d.Subscriptions[i].ID == subID && d.Subscriptions[i].ClientID == user.ID {
			if d.Subscriptions[i].Status != "active" {
				d.Mu.Unlock()
				Error(w, 400, "Can only change active subscriptions")
				return
			}
			validPlan := false
			for _, p := range d.SubscriptionPlans {
				if p.ID == req.NewPlanID && p.IsActive {
					validPlan = true
					break
				}
			}
			if !validPlan {
				d.Mu.Unlock()
				Error(w, 404, "New plan not found or inactive")
				return
			}
			d.Subscriptions[i].PlanID = req.NewPlanID
			found = true
			break
		}
	}
	d.Mu.Unlock()

	if !found {
		Error(w, 404, "Subscription not found")
		return
	}
	d.Save()

	JSON(w, 200, H{"message": "Plan changed"})
}

func PauseSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	updateSubscriptionStatus(w, r, "paused")
}

func CancelSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	updateSubscriptionStatus(w, r, "cancelled")
}

func updateSubscriptionStatus(w http.ResponseWriter, r *http.Request, newStatus string) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	subID := strings.TrimPrefix(r.URL.Path, "/api/v1/subscriptions/")
	subID = strings.TrimSuffix(subID, "/pause")
	subID = strings.TrimSuffix(subID, "/cancel")

	d := database.GetDB()
	d.Mu.Lock()
	found := false
	for i := range d.Subscriptions {
		if d.Subscriptions[i].ID == subID && d.Subscriptions[i].ClientID == user.ID {
			if d.Subscriptions[i].Status != "active" {
				d.Mu.Unlock()
				Error(w, 400, "Can only pause/cancel active subscriptions")
				return
			}
			d.Subscriptions[i].Status = newStatus
			if newStatus == "cancelled" {
				d.Subscriptions[i].CancelledAt = store.Now()
			}
			found = true
			break
		}
	}
	d.Mu.Unlock()

	if !found {
		Error(w, 404, "Subscription not found")
		return
	}
	d.Save()

	JSON(w, 200, H{"message": "Subscription " + newStatus})
}

func GetSubscriptionDeliverablesHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	subID := strings.TrimPrefix(r.URL.Path, "/api/v1/subscriptions/")
	subID = strings.TrimSuffix(subID, "/deliverables")

	d := database.GetDB()
	d.Mu.RLock()
	var sub *models.Subscription
	for i := range d.Subscriptions {
		if d.Subscriptions[i].ID == subID {
			sub = &d.Subscriptions[i]
			break
		}
	}
	d.Mu.RUnlock()

	if sub == nil {
		Error(w, 404, "Subscription not found")
		return
	}
	if sub.ClientID != user.ID && sub.FreelancerID != user.ID {
		Error(w, 403, "Not authorized")
		return
	}

	d.Mu.RLock()
	var deliverables []models.SubscriptionDeliverable
	for _, del := range d.SubDeliverables {
		if del.SubscriptionID == subID {
			deliverables = append(deliverables, del)
		}
	}
	d.Mu.RUnlock()

	if deliverables == nil {
		deliverables = []models.SubscriptionDeliverable{}
	}
	JSON(w, 200, deliverables)
}
