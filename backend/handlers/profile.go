package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	d := database.GetDB()
	d.Mu.RLock()
	categories := make([]models.Category, len(d.Categories))
	copy(categories, d.Categories)
	d.Mu.RUnlock()
	JSON(w, 200, categories)
}

func GetMeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}
	JSON(w, 200, user)
}

func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	if id == "" || id == r.URL.Path {
		id = strings.TrimPrefix(r.URL.Path, "/api/v1/me")
	}

	user := store.FindUserByID(id)
	if user == nil {
		Error(w, 404, "User not found")
		return
	}

	self := GetUser(r)
	if self != nil && self.ID == user.ID {
		JSON(w, 200, user)
		return
	}
	JSON(w, 200, store.ToPublic(*user))
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Name       string   `json:"name"`
		Bio        string   `json:"bio"`
		Skills     []string `json:"skills"`
		HourlyRate float64  `json:"hourlyRate"`
		Location   string   `json:"location"`
		Website    string   `json:"website"`
		Avatar     string   `json:"avatar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Users {
		if d.Users[i].ID == user.ID {
			d.Users[i].Name = req.Name
			d.Users[i].Bio = req.Bio
			d.Users[i].Skills = req.Skills
			d.Users[i].HourlyRate = req.HourlyRate
			d.Users[i].Location = req.Location
			d.Users[i].Website = req.Website
			d.Users[i].Avatar = req.Avatar
			break
		}
	}
	d.Mu.Unlock()
	d.Save()
	JSON(w, 200, H{"message": "Profile updated"})
}

func GetFreelancersHandler(w http.ResponseWriter, r *http.Request) {
	d := database.GetDB()
	d.Mu.RLock()
	var users []models.User
	for _, u := range d.Users {
		if u.Role == "freelancer" {
			users = append(users, u)
		}
	}
	d.Mu.RUnlock()

	var result []models.UserPublic
	for _, u := range users {
		result = append(result, store.ToPublic(u))
	}
	JSON(w, 200, result)
}
