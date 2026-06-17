package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"regexp"
	"strings"
	"time"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9\-]`)

func CreateOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Logo        string `json:"logo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		Error(w, 400, "Name is required")
		return
	}

	slug := strings.ToLower(req.Name)
	slug = slugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	d := database.GetDB()
	d.Mu.RLock()
	for _, org := range d.Organizations {
		if org.Slug == slug {
			d.Mu.RUnlock()
			Error(w, 409, "Organization with this name already exists")
			return
		}
	}
	d.Mu.RUnlock()

	orgID := store.NewID()
	org := models.Organization{
		ID:          orgID,
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Logo:        req.Logo,
		OwnerID:     user.ID,
		MemberCount: 1,
		CreatedAt:   store.Now(),
	}

	member := models.OrgMember{
		ID:       store.NewID(),
		OrgID:    orgID,
		UserID:   user.ID,
		Role:     "owner",
		JoinedAt: store.Now(),
	}

	d.Mu.Lock()
	d.Organizations = append(d.Organizations, org)
	d.OrgMembers = append(d.OrgMembers, member)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": orgID, "message": "Organization created"})
}

func GetOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	orgID := strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/")

	d := database.GetDB()
	d.Mu.RLock()
	var org *models.Organization
	for i := range d.Organizations {
		if d.Organizations[i].ID == orgID {
			org = &d.Organizations[i]
			break
		}
	}
	d.Mu.RUnlock()

	if org == nil {
		Error(w, 404, "Organization not found")
		return
	}

	JSON(w, 200, org)
}

func UpdateOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	orgID := strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Logo        string `json:"logo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Organizations {
		if d.Organizations[i].ID == orgID {
			if d.Organizations[i].OwnerID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Only the owner can update the organization")
				return
			}
			if req.Name != "" {
				d.Organizations[i].Name = req.Name
			}
			if req.Description != "" {
				d.Organizations[i].Description = req.Description
			}
			if req.Logo != "" {
				d.Organizations[i].Logo = req.Logo
			}
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Organization updated"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Organization not found")
}

func InviteMemberHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	orgID := strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/")
	orgID = strings.TrimSuffix(orgID, "/members")

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		Error(w, 400, "Email is required")
		return
	}

	if req.Role != "manager" && req.Role != "member" {
		req.Role = "member"
	}

	d := database.GetDB()
	d.Mu.RLock()
	var org *models.Organization
	for i := range d.Organizations {
		if d.Organizations[i].ID == orgID {
			org = &d.Organizations[i]
			break
		}
	}
	d.Mu.RUnlock()

	if org == nil {
		Error(w, 404, "Organization not found")
		return
	}

	isMember := false
	for _, m := range d.OrgMembers {
		if m.OrgID == orgID && m.UserID == user.ID && (m.Role == "owner" || m.Role == "manager") {
			isMember = true
			break
		}
	}
	if !isMember {
		Error(w, 403, "Only owners and managers can invite members")
		return
	}

	existing := store.FindUserByEmail(req.Email)
	if existing != nil {
		for _, m := range d.OrgMembers {
			if m.OrgID == orgID && m.UserID == existing.ID {
				Error(w, 409, "User is already a member")
				return
			}
		}
	}

	expires := time.Now().UTC().Add(7 * 24 * time.Hour)
	invite := models.OrgInvite{
		ID:        store.NewID(),
		OrgID:     orgID,
		Email:     req.Email,
		Role:      req.Role,
		InvitedBy: user.ID,
		Status:    "pending",
		ExpiresAt: expires.Format(time.RFC3339),
	}

	d.Mu.Lock()
	d.OrgInvites = append(d.OrgInvites, invite)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": invite.ID, "message": "Invitation sent"})
}

func RemoveMemberHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		Error(w, 400, "Invalid path")
		return
	}
	orgID := parts[0]
	memberUserID := parts[2]

	d := database.GetDB()
	d.Mu.RLock()
	isOwner := false
	for _, m := range d.OrgMembers {
		if m.OrgID == orgID && m.UserID == user.ID && m.Role == "owner" {
			isOwner = true
			break
		}
	}
	d.Mu.RUnlock()

	if !isOwner {
		Error(w, 403, "Only the owner can remove members")
		return
	}

	if memberUserID == user.ID {
		Error(w, 400, "Cannot remove yourself")
		return
	}

	d.Mu.Lock()
	for i := range d.OrgMembers {
		if d.OrgMembers[i].OrgID == orgID && d.OrgMembers[i].UserID == memberUserID {
			d.OrgMembers = append(d.OrgMembers[:i], d.OrgMembers[i+1:]...)
			break
		}
	}
	for i := range d.Organizations {
		if d.Organizations[i].ID == orgID {
			d.Organizations[i].MemberCount--
			break
		}
	}
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Member removed"})
}

func ChangeMemberRoleHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		Error(w, 400, "Invalid path")
		return
	}
	orgID := parts[0]
	memberUserID := parts[2]

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	if req.Role != "owner" && req.Role != "manager" && req.Role != "member" {
		Error(w, 400, "Invalid role")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	isOwner := false
	for _, m := range d.OrgMembers {
		if m.OrgID == orgID && m.UserID == user.ID && m.Role == "owner" {
			isOwner = true
			break
		}
	}
	d.Mu.RUnlock()

	if !isOwner {
		Error(w, 403, "Only the owner can change roles")
		return
	}

	d.Mu.Lock()
	for i := range d.OrgMembers {
		if d.OrgMembers[i].OrgID == orgID && d.OrgMembers[i].UserID == memberUserID {
			d.OrgMembers[i].Role = req.Role
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Role updated"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Member not found")
}

func ListOrgOrdersHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	orgID := strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/")
	orgID = strings.TrimSuffix(orgID, "/orders")

	d := database.GetDB()
	d.Mu.RLock()
	isMember := false
	for _, m := range d.OrgMembers {
		if m.OrgID == orgID && m.UserID == user.ID {
			isMember = true
			break
		}
	}
	d.Mu.RUnlock()

	if !isMember {
		Error(w, 403, "Not a member of this organization")
		return
	}

	var memberIDs []string
	d.Mu.RLock()
	for _, m := range d.OrgMembers {
		if m.OrgID == orgID {
			memberIDs = append(memberIDs, m.UserID)
		}
	}

	var orders []models.Order
	for _, o := range d.Orders {
		for _, id := range memberIDs {
			if o.SellerID == id || o.BuyerID == id {
				orders = append(orders, o)
				break
			}
		}
	}
	d.Mu.RUnlock()

	if orders == nil {
		orders = []models.Order{}
	}
	JSON(w, 200, orders)
}

func OrgAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	orgID := strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/")
	orgID = strings.TrimSuffix(orgID, "/analytics")

	d := database.GetDB()
	d.Mu.RLock()
	isMember := false
	var memberIDs []string
	for _, m := range d.OrgMembers {
		if m.OrgID == orgID {
			memberIDs = append(memberIDs, m.UserID)
			if m.UserID == user.ID {
				isMember = true
			}
		}
	}
	d.Mu.RUnlock()

	if !isMember {
		Error(w, 403, "Not a member of this organization")
		return
	}

	var totalOrders, completed int
	var totalRevenue float64
	d.Mu.RLock()
	for _, o := range d.Orders {
		for _, id := range memberIDs {
			if o.SellerID == id {
				totalOrders++
				if o.Status == "completed" {
					completed++
					totalRevenue += o.Price
				}
				break
			}
		}
	}
	d.Mu.RUnlock()

	JSON(w, 200, H{
		"memberCount":  len(memberIDs),
		"totalOrders":  totalOrders,
		"completed":    completed,
		"totalRevenue": totalRevenue,
	})
}
