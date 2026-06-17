package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func OpenDisputeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	orderID := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	orderID = strings.TrimSuffix(orderID, "/dispute")

	order := store.FindOrderByID(orderID)
	if order == nil {
		Error(w, 404, "Order not found")
		return
	}

	if order.BuyerID != user.ID && order.SellerID != user.ID {
		Error(w, 403, "Not authorized")
		return
	}

	if order.Status != "delivered" && order.Status != "active" && order.Status != "revision" {
		Error(w, 400, "Can only dispute active or delivered orders")
		return
	}

	if order.DisputeID != "" {
		Error(w, 409, "Dispute already exists for this order")
		return
	}

	var req struct {
		Reason      string `json:"reason"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Description == "" {
		Error(w, 400, "Reason and description are required")
		return
	}

	validReasons := map[string]bool{
		"quality": true, "scope": true, "communication": true, "deadline": true, "other": true,
	}
	if !validReasons[req.Reason] {
		req.Reason = "other"
	}

	role := "client"
	if user.ID == order.SellerID {
		role = "freelancer"
	}

	disputeID := store.NewID()
	dispute := models.Dispute{
		ID:          disputeID,
		OrderID:     orderID,
		OpenedBy:    user.ID,
		OpenedRole:  role,
		Reason:      req.Reason,
		Description: req.Description,
		Status:      "open",
		CreatedAt:   store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.Disputes = append(d.Disputes, dispute)
	for i := range d.Orders {
		if d.Orders[i].ID == orderID {
			d.Orders[i].DisputeID = disputeID
			break
		}
	}
	d.Mu.Unlock()
	d.Save()

	otherID := order.BuyerID
	if user.ID == order.BuyerID {
		otherID = order.SellerID
	}
	notif := models.Notification{
		ID: store.NewID(), UserID: otherID, Type: "dispute",
		Title: "Dispute Opened", Message: "A dispute has been opened on order #" + orderID[:8],
		Link: "/orders/" + orderID, CreatedAt: store.Now(),
	}
	d.Mu.Lock()
	d.Notifications = append(d.Notifications, notif)
	d.Mu.Unlock()
	d.Save()
	BroadcastNotification(otherID, notif)

	JSON(w, 201, H{"id": disputeID, "message": "Dispute opened"})
}

func SubmitEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	disputeID := strings.TrimPrefix(r.URL.Path, "/api/v1/disputes/")
	disputeID = strings.TrimSuffix(disputeID, "/evidence")

	d := database.GetDB()
	d.Mu.RLock()
	var dispute *models.Dispute
	for i := range d.Disputes {
		if d.Disputes[i].ID == disputeID {
			dispute = &d.Disputes[i]
			break
		}
	}
	d.Mu.RUnlock()

	if dispute == nil {
		Error(w, 404, "Dispute not found")
		return
	}

	if dispute.Status == "resolved" {
		Error(w, 400, "Dispute is already resolved")
		return
	}

	order := store.FindOrderByID(dispute.OrderID)
	if order == nil || (order.BuyerID != user.ID && order.SellerID != user.ID) {
		Error(w, 403, "Not authorized")
		return
	}

	var req struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		Error(w, 400, "Content is required")
		return
	}

	if req.Type != "file" {
		req.Type = "text"
	}

	evidence := models.DisputeEvidence{
		ID:        store.NewID(),
		DisputeID: disputeID,
		UserID:    user.ID,
		Type:      req.Type,
		Content:   req.Content,
		CreatedAt: store.Now(),
	}

	d.Mu.Lock()
	d.DisputeEvidences = append(d.DisputeEvidences, evidence)
	for i := range d.Disputes {
		if d.Disputes[i].ID == disputeID && d.Disputes[i].Status == "open" {
			d.Disputes[i].Status = "evidence_pending"
			break
		}
	}
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": evidence.ID, "message": "Evidence submitted"})
}

func ResolveDisputeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	disputeID := strings.TrimPrefix(r.URL.Path, "/api/v1/disputes/")
	disputeID = strings.TrimSuffix(disputeID, "/resolve")

	var req struct {
		Ruling      string  `json:"ruling"`
		SplitClient float64 `json:"splitClient"`
		AdminNote   string  `json:"adminNote"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	validRulings := map[string]bool{
		"full_refund": true, "full_release": true, "split": true,
	}
	if !validRulings[req.Ruling] {
		Error(w, 400, "Ruling must be: full_refund, full_release, or split")
		return
	}

	if req.Ruling == "split" && (req.SplitClient < 0 || req.SplitClient > 100) {
		Error(w, 400, "Split percentage must be 0-100")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	found := false
	for i := range d.Disputes {
		if d.Disputes[i].ID == disputeID {
			if d.Disputes[i].Status == "resolved" {
				d.Mu.Unlock()
				Error(w, 400, "Dispute already resolved")
				return
			}
			d.Disputes[i].Status = "resolved"
			d.Disputes[i].Resolution = &models.Resolution{
				Ruling:      req.Ruling,
				SplitClient: req.SplitClient,
				AdminNote:   req.AdminNote,
				ResolvedBy:  user.ID,
			}
			d.Disputes[i].ResolvedAt = store.Now()

			orderID := d.Disputes[i].OrderID
			for j := range d.Orders {
				if d.Orders[j].ID == orderID {
					switch req.Ruling {
					case "full_refund":
						d.Orders[j].Status = "refunded"
						d.Orders[j].EscrowStatus = "refunded"
					case "full_release":
						d.Orders[j].Status = "completed"
						d.Orders[j].EscrowStatus = "released"
						for k := range d.Users {
							if d.Users[k].ID == d.Orders[j].SellerID {
								d.Users[k].Earnings += d.Orders[j].Price
								break
							}
						}
					case "split":
						d.Orders[j].Status = "completed"
						d.Orders[j].EscrowStatus = "split"
						amount := d.Orders[j].Price * req.SplitClient / 100
						for k := range d.Users {
							if d.Users[k].ID == d.Orders[j].SellerID {
								d.Users[k].Earnings += d.Orders[j].Price - amount
								break
							}
						}
					}
					break
				}
			}
			found = true
			break
		}
	}
	d.Mu.Unlock()

	if !found {
		Error(w, 404, "Dispute not found")
		return
	}
	d.Save()

	JSON(w, 200, H{"message": "Dispute resolved", "ruling": req.Ruling})
}

func AcceptResolutionHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	disputeID := strings.TrimPrefix(r.URL.Path, "/api/v1/disputes/")
	disputeID = strings.TrimSuffix(disputeID, "/accept")

	d := database.GetDB()
	d.Mu.RLock()
	var dispute *models.Dispute
	for i := range d.Disputes {
		if d.Disputes[i].ID == disputeID {
			dispute = &d.Disputes[i]
			break
		}
	}
	d.Mu.RUnlock()

	if dispute == nil {
		Error(w, 404, "Dispute not found")
		return
	}

	if dispute.Status != "resolved" {
		Error(w, 400, "Dispute is not resolved yet")
		return
	}

	order := store.FindOrderByID(dispute.OrderID)
	if order == nil || (order.BuyerID != user.ID && order.SellerID != user.ID) {
		Error(w, 403, "Not authorized")
		return
	}

	JSON(w, 200, H{"message": "Resolution accepted"})
}

func GetDisputeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	disputeID := strings.TrimPrefix(r.URL.Path, "/api/v1/disputes/")

	d := database.GetDB()
	d.Mu.RLock()
	var dispute *models.Dispute
	for i := range d.Disputes {
		if d.Disputes[i].ID == disputeID {
			dispute = &d.Disputes[i]
			break
		}
	}
	d.Mu.RUnlock()

	if dispute == nil {
		Error(w, 404, "Dispute not found")
		return
	}

	order := store.FindOrderByID(dispute.OrderID)
	if order == nil || (order.BuyerID != user.ID && order.SellerID != user.ID && user.Role != "admin") {
		Error(w, 403, "Not authorized")
		return
	}

	d.Mu.RLock()
	var evidence []models.DisputeEvidence
	for _, e := range d.DisputeEvidences {
		if e.DisputeID == disputeID {
			evidence = append(evidence, e)
		}
	}
	d.Mu.RUnlock()

	if evidence == nil {
		evidence = []models.DisputeEvidence{}
	}

	JSON(w, 200, H{"dispute": dispute, "evidence": evidence})
}

func ListDisputesHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var disputes []models.Dispute
	for _, dispute := range d.Disputes {
		if user.Role == "admin" {
			disputes = append(disputes, dispute)
		} else {
			order := store.FindOrderByID(dispute.OrderID)
			if order != nil && (order.BuyerID == user.ID || order.SellerID == user.ID) {
				disputes = append(disputes, dispute)
			}
		}
	}
	d.Mu.RUnlock()

	if disputes == nil {
		disputes = []models.Dispute{}
	}
	JSON(w, 200, disputes)
}

func AdminDisputeStatsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var open, pending, resolved int
	for _, dispute := range d.Disputes {
		switch dispute.Status {
		case "open":
			open++
		case "evidence_pending", "under_review":
			pending++
		case "resolved":
			resolved++
		}
	}
	d.Mu.RUnlock()

	JSON(w, 200, H{
		"open":     open,
		"pending":  pending,
		"resolved": resolved,
		"total":    open + pending + resolved,
	})
}
