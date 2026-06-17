package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func CreateSavedSearchHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Query      string   `json:"query"`
		Category   string   `json:"category"`
		MinPrice   float64  `json:"minPrice"`
		MaxPrice   float64  `json:"maxPrice"`
		MinRating  float64  `json:"minRating"`
		Skills     []string `json:"skills"`
		AlertEmail bool     `json:"alertEmail"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	search := models.SavedSearch{
		ID:         store.NewID(),
		UserID:     user.ID,
		Query:      req.Query,
		Category:   req.Category,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		MinRating:  req.MinRating,
		Skills:     req.Skills,
		AlertEmail: req.AlertEmail,
		CreatedAt:  store.Now(),
	}
	if search.Skills == nil {
		search.Skills = []string{}
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.SavedSearches = append(d.SavedSearches, search)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": search.ID, "message": "Search saved"})
}

func ListSavedSearchesHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var searches []models.SavedSearch
	for _, s := range d.SavedSearches {
		if s.UserID == user.ID {
			searches = append(searches, s)
		}
	}
	d.Mu.RUnlock()

	if searches == nil {
		searches = []models.SavedSearch{}
	}
	JSON(w, 200, searches)
}

func DeleteSavedSearchHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	searchID := strings.TrimPrefix(r.URL.Path, "/api/v1/searches/")

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.SavedSearches {
		if d.SavedSearches[i].ID == searchID && d.SavedSearches[i].UserID == user.ID {
			d.SavedSearches = append(d.SavedSearches[:i], d.SavedSearches[i+1:]...)
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Search deleted"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Search not found")
}

func TrendingSearchesHandler(w http.ResponseWriter, r *http.Request) {
	d := database.GetDB()
	d.Mu.RLock()

	queryCount := make(map[string]int)
	for _, s := range d.SavedSearches {
		if s.Query != "" {
			queryCount[s.Query]++
		}
	}
	d.Mu.RUnlock()

	type trending struct {
		Query string `json:"query"`
		Count int    `json:"count"`
	}

	var results []trending
	for q, c := range queryCount {
		results = append(results, trending{Query: q, Count: c})
	}

	if results == nil {
		results = []trending{}
	}
	JSON(w, 200, results)
}
