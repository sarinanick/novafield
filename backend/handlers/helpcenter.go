package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

var defaultArticles = []models.HelpArticle{
	{ID: "help-1", Title: "How to get started", Content: "Welcome to NovaField! Create your profile, browse gigs, and place orders.", Category: "getting-started", Tags: []string{"beginner", "setup"}},
	{ID: "help-2", Title: "How to create a gig", Content: "Go to Create Gig, fill in title, description, pricing packages, and publish.", Category: "gigs", Tags: []string{"gig", "create"}},
	{ID: "help-3", Title: "How payments work", Content: "Payments are held in escrow until the order is approved. Milestones allow partial payments.", Category: "payments", Tags: []string{"payment", "escrow"}},
	{ID: "help-4", Title: "How to resolve disputes", Content: "Open a dispute within 7 days of delivery. Submit evidence and wait for admin review.", Category: "disputes", Tags: []string{"dispute", "refund"}},
	{ID: "help-5", Title: "Verification and trust badges", Content: "Verify your email, phone, and identity to earn trust badges and get more orders.", Category: "trust", Tags: []string{"verification", "trust", "badge"}},
	{ID: "help-6", Title: "Using the virtual office", Content: "Navigate the 2D world, claim desks, join coworking sessions, and collaborate with others.", Category: "virtual-office", Tags: []string{"world", "coworking", "desk"}},
}

func ListHelpArticlesHandler(w http.ResponseWriter, r *http.Request) {
	JSON(w, 200, defaultArticles)
}

func GetHelpArticleHandler(w http.ResponseWriter, r *http.Request) {
	articleID := strings.TrimPrefix(r.URL.Path, "/api/v1/help/articles/")

	for _, a := range defaultArticles {
		if a.ID == articleID {
			JSON(w, 200, a)
			return
		}
	}
	Error(w, 404, "Article not found")
}

func SearchHelpHandler(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	if query == "" {
		JSON(w, 200, defaultArticles)
		return
	}

	var results []models.HelpArticle
	for _, a := range defaultArticles {
		if strings.Contains(strings.ToLower(a.Title), query) ||
			strings.Contains(strings.ToLower(a.Content), query) {
			results = append(results, a)
		}
		for _, tag := range a.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, a)
				break
			}
		}
	}

	if results == nil {
		results = []models.HelpArticle{}
	}
	JSON(w, 200, results)
}

func CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Subject  string `json:"subject"`
		Category string `json:"category"`
		Priority string `json:"priority"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Subject == "" || req.Content == "" {
		Error(w, 400, "Subject and content are required")
		return
	}

	validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "urgent": true}
	if !validPriorities[req.Priority] {
		req.Priority = "medium"
	}

	ticketID := store.NewID()
	ticket := models.SupportTicket{
		ID:        ticketID,
		UserID:    user.ID,
		Subject:   req.Subject,
		Category:  req.Category,
		Priority:  req.Priority,
		Status:    "open",
		CreatedAt: store.Now(),
		Replies: []models.TicketReply{
			{
				ID:        store.NewID(),
				TicketID:  ticketID,
				UserID:    user.ID,
				Content:   req.Content,
				IsStaff:   false,
				CreatedAt: store.Now(),
			},
		},
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.SupportTickets = append(d.SupportTickets, ticket)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": ticketID, "message": "Support ticket created"})
}

func ListMyTicketsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var tickets []models.SupportTicket
	for _, t := range d.SupportTickets {
		if t.UserID == user.ID {
			tickets = append(tickets, t)
		}
	}
	d.Mu.RUnlock()

	if tickets == nil {
		tickets = []models.SupportTicket{}
	}
	JSON(w, 200, tickets)
}

func AdminListTicketsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	tickets := make([]models.SupportTicket, len(d.SupportTickets))
	copy(tickets, d.SupportTickets)
	d.Mu.RUnlock()

	JSON(w, 200, tickets)
}

func AdminUpdateTicketHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	ticketID := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/support/tickets/")

	var req struct {
		Status  string `json:"status"`
		Reply   string `json:"reply"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.SupportTickets {
		if d.SupportTickets[i].ID == ticketID {
			if req.Status != "" {
				d.SupportTickets[i].Status = req.Status
			}
			if req.Reply != "" {
				d.SupportTickets[i].Replies = append(d.SupportTickets[i].Replies, models.TicketReply{
					ID:        store.NewID(),
					TicketID:  ticketID,
					UserID:    user.ID,
					Content:   req.Reply,
					IsStaff:   true,
					CreatedAt: store.Now(),
				})
			}
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Ticket updated"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Ticket not found")
}

func AddWorkspaceCommentHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	orderID := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	orderID = strings.TrimSuffix(orderID, "/workspace/comments")

	order := store.FindOrderByID(orderID)
	if order == nil || (order.BuyerID != user.ID && order.SellerID != user.ID) {
		Error(w, 403, "Not authorized")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		Error(w, 400, "Content is required")
		return
	}

	comment := models.WorkspaceComment{
		ID:        store.NewID(),
		OrderID:   orderID,
		UserID:    user.ID,
		Content:   req.Content,
		CreatedAt: store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.WorkspaceComments = append(d.WorkspaceComments, comment)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": comment.ID, "message": "Comment added"})
}

func AddWorkspaceTaskHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	orderID := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	orderID = strings.TrimSuffix(orderID, "/workspace/tasks")

	order := store.FindOrderByID(orderID)
	if order == nil || (order.BuyerID != user.ID && order.SellerID != user.ID) {
		Error(w, 403, "Not authorized")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		AssignedTo  string `json:"assignedTo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		Error(w, 400, "Title is required")
		return
	}

	task := models.WorkspaceTask{
		ID:          store.NewID(),
		OrderID:     orderID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "todo",
		AssignedTo:  req.AssignedTo,
		CreatedAt:   store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.WorkspaceTasks = append(d.WorkspaceTasks, task)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": task.ID, "message": "Task added"})
}

func UpdateWorkspaceTaskHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		Error(w, 400, "Invalid path")
		return
	}
	orderID := parts[0]
	taskID := parts[3]

	order := store.FindOrderByID(orderID)
	if order == nil || (order.BuyerID != user.ID && order.SellerID != user.ID) {
		Error(w, 403, "Not authorized")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		AssignedTo  string `json:"assignedTo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.WorkspaceTasks {
		if d.WorkspaceTasks[i].ID == taskID && d.WorkspaceTasks[i].OrderID == orderID {
			if req.Title != "" {
				d.WorkspaceTasks[i].Title = req.Title
			}
			if req.Description != "" {
				d.WorkspaceTasks[i].Description = req.Description
			}
			if req.Status != "" {
				d.WorkspaceTasks[i].Status = req.Status
				if req.Status == "done" {
					d.WorkspaceTasks[i].CompletedAt = store.Now()
				}
			}
			if req.AssignedTo != "" {
				d.WorkspaceTasks[i].AssignedTo = req.AssignedTo
			}
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Task updated"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Task not found")
}
