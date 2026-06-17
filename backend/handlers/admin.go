package handlers

import (
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

type WorkspaceSettings struct {
	Name            string `json:"name"`
	DefaultLayout   string `json:"defaultLayout"`
	MaxMembers      int    `json:"maxMembers"`
	GuestAccess     bool   `json:"guestAccess"`
	EmailNotify     bool   `json:"emailNotify"`
}

var workspaceSettings = WorkspaceSettings{
	Name:          "NovaField AI",
	DefaultLayout: "open",
	MaxMembers:    100,
	GuestAccess:   true,
	EmailNotify:   true,
}

func AdminRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		AdminGetMembersHandler(w, r)
	default:
		Error(w, 405, "Method not allowed")
	}
}

func AdminActionRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/")

	if path == "members/invite" || path == "members/invite/" {
		if r.Method != "POST" {
			Error(w, 405, "Method not allowed")
			return
		}
		AdminInviteMemberHandler(w, r)
		return
	}

	if path == "settings" || path == "settings/" {
		switch r.Method {
		case "GET":
			AdminGetSettingsHandler(w, r)
		case "PUT":
			AdminUpdateSettingsHandler(w, r)
		default:
			Error(w, 405, "Method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "members/") {
		trimmed := strings.TrimPrefix(path, "members/")
		parts := strings.Split(trimmed, "/")
		if len(parts) >= 2 {
			id := parts[0]
			action := parts[1]
			switch action {
			case "role":
				AdminChangeRoleHandler(w, r, id)
				return
			case "demote":
				AdminDemoteMemberHandler(w, r, id)
				return
			}
		}
		if len(parts) == 1 && r.Method == "DELETE" {
			AdminRemoveMemberHandler(w, r, parts[0])
			return
		}
	}

	Error(w, 404, "Not found")
}

func AdminGetMembersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	members := make([]models.UserPublic, 0, len(d.Users))
	for _, u := range d.Users {
		members = append(members, store.ToPublic(u))
	}
	d.Mu.RUnlock()

	JSON(w, 200, H{
		"members": members,
		"total":   len(members),
	})
}

func AdminInviteMemberHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := Decode(r, &req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	if req.Email == "" {
		Error(w, 400, "Email is required")
		return
	}

	if req.Role == "" {
		req.Role = "client"
	}
	if req.Role != "admin" && req.Role != "client" && req.Role != "freelancer" {
		Error(w, 400, "Invalid role. Must be admin, client, or freelancer")
		return
	}

	existing := store.FindUserByEmail(req.Email)
	if existing != nil {
		Error(w, 409, "User with this email already exists")
		return
	}

	newUser := models.User{
		ID:           store.NewID(),
		Email:        req.Email,
		PasswordHash: store.HashPassword("changeme123"),
		Name:         strings.Split(req.Email, "@")[0],
		Role:         req.Role,
		JoinedAt:     store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.Users = append(d.Users, newUser)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{
		"message": "Member invited successfully",
		"user":    store.ToPublic(newUser),
	})
}

func AdminChangeRoleHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != "PUT" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := Decode(r, &req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	if req.Role != "admin" && req.Role != "client" && req.Role != "freelancer" {
		Error(w, 400, "Invalid role")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Users {
		if d.Users[i].ID == id {
			d.Users[i].Role = req.Role
			pub := store.ToPublic(d.Users[i])
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{
				"message": "Role updated",
				"user":    pub,
			})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Member not found")
}

func AdminRemoveMemberHandler(w http.ResponseWriter, r *http.Request, id string) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	if id == user.ID {
		Error(w, 400, "Cannot remove yourself")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Users {
		if d.Users[i].ID == id {
			d.Users = append(d.Users[:i], d.Users[i+1:]...)
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Member removed"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Member not found")
}

func AdminDemoteMemberHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != "POST" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Users {
		if d.Users[i].ID == id {
			d.Users[i].Role = "client"
			pub := store.ToPublic(d.Users[i])
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{
				"message": "Member demoted to client",
				"user":    pub,
			})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Member not found")
}

func AdminGetSettingsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	JSON(w, 200, workspaceSettings)
}

func AdminUpdateSettingsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	var req WorkspaceSettings
	if err := Decode(r, &req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	if req.Name != "" {
		workspaceSettings.Name = req.Name
	}
	if req.DefaultLayout != "" {
		workspaceSettings.DefaultLayout = req.DefaultLayout
	}
	if req.MaxMembers > 0 {
		workspaceSettings.MaxMembers = req.MaxMembers
	}
	workspaceSettings.GuestAccess = req.GuestAccess
	workspaceSettings.EmailNotify = req.EmailNotify

	JSON(w, 200, workspaceSettings)
}
