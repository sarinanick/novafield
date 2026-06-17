package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strconv"
	"strings"
)

func GetGigsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 12
	}

	result := store.SearchGigs(
		q.Get("q"), q.Get("category"), q.Get("minPrice"), q.Get("maxPrice"),
		q.Get("minRating"), q.Get("delivery"), q.Get("sort"), page, limit,
	)
	JSON(w, 200, result)
}

func GetGigHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/gigs/")
	id = strings.Split(id, "/")[0]

	gig := store.FindGigByID(id)
	if gig == nil {
		Error(w, 404, "Gig not found")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Gigs {
		if d.Gigs[i].ID == id {
			d.Gigs[i].Views++
			break
		}
	}
	d.Mu.Unlock()
	d.Save()
	gig.Views++

	user := store.FindUserByID(gig.FreelancerID)
	if user != nil {
		pub := store.ToPublic(*user)
		gig.Freelancer = &pub
	}

	rating, rc := store.CalcGigRating(gig.ID)
	gig.Rating = rating
	gig.ReviewsCount = rc

	d.Mu.RLock()
	var pkgs []models.Package
	for _, p := range d.Packages {
		if p.GigID == gig.ID {
			pkgs = append(pkgs, p)
		}
	}
	d.Mu.RUnlock()
	gig.Packages = pkgs

	JSON(w, 200, gig)
}

func CreateGigHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || (user.Role != "freelancer" && user.Role != "admin") {
		Error(w, 403, "Only freelancers can create gigs")
		return
	}

	var req struct {
		Title        string           `json:"title"`
		Description  string           `json:"description"`
		Category     string           `json:"category"`
		Subcategory  string           `json:"subcategory"`
		Tags         []string         `json:"tags"`
		AITools      []string         `json:"aiTools"`
		PriceType    string           `json:"priceType"`
		DeliveryDays int              `json:"deliveryDays"`
		Revisions    int              `json:"revisions"`
		Images       []string         `json:"images"`
		VideoURL     string           `json:"videoUrl"`
		Packages     []models.Package `json:"packages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	gigID := store.NewID()
	now := store.Now()

	gig := models.Gig{
		ID: gigID, FreelancerID: user.ID, Title: req.Title, Description: req.Description,
		Category: req.Category, Subcategory: req.Subcategory, Tags: req.Tags,
		AITools: req.AITools, PriceType: req.PriceType, DeliveryDays: req.DeliveryDays,
		Revisions: req.Revisions, Images: req.Images, VideoURL: req.VideoURL,
		Status: "active", CreatedAt: now, UpdatedAt: now,
	}

	if len(req.Packages) > 0 {
		gig.Price = req.Packages[0].Price
	}

	d := database.GetDB()
	d.Mu.Lock()

	d.Gigs = append(d.Gigs, gig)

	for _, pkg := range req.Packages {
		pkg.ID = store.NewID()
		pkg.GigID = gigID
		d.Packages = append(d.Packages, pkg)
	}

	for i := range d.Categories {
		if d.Categories[i].Slug == req.Category {
			d.Categories[i].GigsCount++
			break
		}
	}

	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": gigID, "message": "Gig created"})
}

func EditGigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/gigs/")
	id = strings.TrimSuffix(id, "/edit")

	gig := store.FindGigByID(id)
	if gig == nil {
		Error(w, 404, "Gig not found")
		return
	}
	if gig.FreelancerID != user.ID {
		Error(w, 403, "Only the owner can edit this gig")
		return
	}

	var req struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Category     string   `json:"category"`
		Subcategory  string   `json:"subcategory"`
		Tags         []string `json:"tags"`
		AITools      []string `json:"aiTools"`
		PriceType    string   `json:"priceType"`
		Price        float64  `json:"price"`
		DeliveryDays int      `json:"deliveryDays"`
		Revisions    int      `json:"revisions"`
		Images       []string `json:"images"`
		VideoURL     string   `json:"videoUrl"`
		Status       string   `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Gigs {
		if d.Gigs[i].ID == id {
			if req.Title != "" {
				d.Gigs[i].Title = req.Title
			}
			if req.Description != "" {
				d.Gigs[i].Description = req.Description
			}
			if req.Category != "" {
				d.Gigs[i].Category = req.Category
			}
			if req.Subcategory != "" {
				d.Gigs[i].Subcategory = req.Subcategory
			}
			if req.Tags != nil {
				d.Gigs[i].Tags = req.Tags
			}
			if req.AITools != nil {
				d.Gigs[i].AITools = req.AITools
			}
			if req.PriceType != "" {
				d.Gigs[i].PriceType = req.PriceType
			}
			if req.Price > 0 {
				d.Gigs[i].Price = req.Price
			}
			if req.DeliveryDays > 0 {
				d.Gigs[i].DeliveryDays = req.DeliveryDays
			}
			if req.Revisions > 0 {
				d.Gigs[i].Revisions = req.Revisions
			}
			if req.Images != nil {
				d.Gigs[i].Images = req.Images
			}
			if req.VideoURL != "" {
				d.Gigs[i].VideoURL = req.VideoURL
			}
			if req.Status != "" {
				d.Gigs[i].Status = req.Status
			}
			d.Gigs[i].UpdatedAt = store.Now()
			break
		}
	}
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Gig updated"})
}

func DeleteGigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/gigs/")

	gig := store.FindGigByID(id)
	if gig == nil {
		Error(w, 404, "Gig not found")
		return
	}
	if gig.FreelancerID != user.ID {
		Error(w, 403, "Only the owner can delete this gig")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, o := range d.Orders {
		if o.GigID == id && (o.Status == "active" || o.Status == "delivered" || o.Status == "revision") {
			d.Mu.RUnlock()
			Error(w, 400, "Cannot delete gig with active orders")
			return
		}
	}
	d.Mu.RUnlock()

	d.Mu.Lock()
	for i := range d.Gigs {
		if d.Gigs[i].ID == id {
			d.Gigs = append(d.Gigs[:i], d.Gigs[i+1:]...)
			break
		}
	}
	var remaining []models.Package
	for i := range d.Packages {
		if d.Packages[i].GigID != id {
			remaining = append(remaining, d.Packages[i])
		}
	}
	d.Packages = remaining
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Gig deleted"})
}

func GetMyGigsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var gigs []models.Gig
	for _, g := range d.Gigs {
		if g.FreelancerID == user.ID {
			gigs = append(gigs, g)
		}
	}
	d.Mu.RUnlock()
	if gigs == nil {
		gigs = []models.Gig{}
	}
	JSON(w, 200, gigs)
}

func GetGigReviewsHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/gigs/")
	id = strings.TrimSuffix(id, "/reviews")

	d := database.GetDB()
	d.Mu.RLock()
	var reviews []models.Review
	for _, rv := range d.Reviews {
		if rv.GigID == id {
			reviews = append(reviews, rv)
		}
	}
	d.Mu.RUnlock()

	for i := range reviews {
		reviewer := store.FindUserByID(reviews[i].ReviewerID)
		if reviewer != nil {
			pub := store.ToPublic(*reviewer)
			reviews[i].Reviewer = &pub
		}
	}
	if reviews == nil {
		reviews = []models.Review{}
	}
	JSON(w, 200, reviews)
}
