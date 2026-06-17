package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Password == "" || req.Name == "" {
		Error(w, 400, "Email, password, and name are required")
		return
	}

	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		Error(w, 400, "Invalid email format")
		return
	}

	if len(req.Password) < 6 {
		Error(w, 400, "Password must be at least 6 characters")
		return
	}

	if store.FindUserByEmail(req.Email) != nil {
		Error(w, 409, "Email already registered")
		return
	}

	role := "client"
	if strings.Contains(r.URL.RawQuery, "role=freelancer") {
		role = "freelancer"
	}

	user := models.User{
		ID: store.NewID(), Email: req.Email, PasswordHash: store.HashPassword(req.Password),
		Name: req.Name, Role: role, JoinedAt: store.Now(), Skills: []string{},
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.Users = append(d.Users, user)
	d.Mu.Unlock()
	d.Save()

	token := store.GenerateToken(user.ID, user.Email, user.Role)
	pub := store.ToPublic(user)
	JSON(w, 201, H{"token": token, "user": pub})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	user := store.FindUserByEmail(req.Email)

	if user == nil || !store.CheckPassword(user.PasswordHash, req.Password) {
		Error(w, 401, "Invalid credentials")
		return
	}

	token := store.GenerateToken(user.ID, user.Email, user.Role)
	pub := store.ToPublic(*user)
	JSON(w, 200, H{"token": token, "user": pub})
}
