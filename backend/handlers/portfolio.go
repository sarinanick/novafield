package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func CreateCaseStudyHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Title        string   `json:"title"`
		ClientName   string   `json:"clientName"`
		Category     string   `json:"category"`
		Challenge    string   `json:"challenge"`
		Approach     string   `json:"approach"`
		Results      string   `json:"results"`
		Images       []string `json:"images"`
		Skills       []string `json:"skills"`
		Duration     string   `json:"duration"`
		BudgetRange  string   `json:"budgetRange"`
		LinkedGigIDs []string `json:"linkedGigIds"`
		IsPublic     bool     `json:"isPublic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		Error(w, 400, "Title is required")
		return
	}

	caseStudy := models.CaseStudy{
		ID:           store.NewID(),
		UserID:       user.ID,
		Title:        req.Title,
		ClientName:   req.ClientName,
		Category:     req.Category,
		Challenge:    req.Challenge,
		Approach:     req.Approach,
		Results:      req.Results,
		Images:       req.Images,
		Skills:       req.Skills,
		Duration:     req.Duration,
		BudgetRange:  req.BudgetRange,
		LinkedGigIDs: req.LinkedGigIDs,
		IsPublic:     req.IsPublic,
		CreatedAt:    store.Now(),
	}

	if caseStudy.Images == nil {
		caseStudy.Images = []string{}
	}
	if caseStudy.Skills == nil {
		caseStudy.Skills = []string{}
	}
	if caseStudy.LinkedGigIDs == nil {
		caseStudy.LinkedGigIDs = []string{}
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.CaseStudies = append(d.CaseStudies, caseStudy)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": caseStudy.ID, "message": "Case study created"})
}

func GetCaseStudyHandler(w http.ResponseWriter, r *http.Request) {
	csID := strings.TrimPrefix(r.URL.Path, "/api/v1/portfolio/cases/")

	user := GetUser(r)

	d := database.GetDB()
	d.Mu.RLock()
	var cs *models.CaseStudy
	for i := range d.CaseStudies {
		if d.CaseStudies[i].ID == csID {
			cs = &d.CaseStudies[i]
			break
		}
	}
	d.Mu.RUnlock()

	if cs == nil {
		Error(w, 404, "Case study not found")
		return
	}

	if !cs.IsPublic && (user == nil || user.ID != cs.UserID) {
		Error(w, 403, "This case study is private")
		return
	}

	JSON(w, 200, cs)
}

func UpdateCaseStudyHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	csID := strings.TrimPrefix(r.URL.Path, "/api/v1/portfolio/cases/")

	var req struct {
		Title        string   `json:"title"`
		ClientName   string   `json:"clientName"`
		Category     string   `json:"category"`
		Challenge    string   `json:"challenge"`
		Approach     string   `json:"approach"`
		Results      string   `json:"results"`
		Images       []string `json:"images"`
		Skills       []string `json:"skills"`
		Duration     string   `json:"duration"`
		BudgetRange  string   `json:"budgetRange"`
		LinkedGigIDs []string `json:"linkedGigIds"`
		IsPublic     *bool    `json:"isPublic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.CaseStudies {
		if d.CaseStudies[i].ID == csID {
			if d.CaseStudies[i].UserID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Not your case study")
				return
			}
			if req.Title != "" {
				d.CaseStudies[i].Title = req.Title
			}
			if req.ClientName != "" {
				d.CaseStudies[i].ClientName = req.ClientName
			}
			if req.Category != "" {
				d.CaseStudies[i].Category = req.Category
			}
			if req.Challenge != "" {
				d.CaseStudies[i].Challenge = req.Challenge
			}
			if req.Approach != "" {
				d.CaseStudies[i].Approach = req.Approach
			}
			if req.Results != "" {
				d.CaseStudies[i].Results = req.Results
			}
			if req.Images != nil {
				d.CaseStudies[i].Images = req.Images
			}
			if req.Skills != nil {
				d.CaseStudies[i].Skills = req.Skills
			}
			if req.Duration != "" {
				d.CaseStudies[i].Duration = req.Duration
			}
			if req.BudgetRange != "" {
				d.CaseStudies[i].BudgetRange = req.BudgetRange
			}
			if req.LinkedGigIDs != nil {
				d.CaseStudies[i].LinkedGigIDs = req.LinkedGigIDs
			}
			if req.IsPublic != nil {
				d.CaseStudies[i].IsPublic = *req.IsPublic
			}
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Case study updated"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Case study not found")
}

func DeleteCaseStudyHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	csID := strings.TrimPrefix(r.URL.Path, "/api/v1/portfolio/cases/")

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.CaseStudies {
		if d.CaseStudies[i].ID == csID {
			if d.CaseStudies[i].UserID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Not your case study")
				return
			}
			d.CaseStudies = append(d.CaseStudies[:i], d.CaseStudies[i+1:]...)
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Case study deleted"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Case study not found")
}

func GetUserPortfolioHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	userID = strings.TrimSuffix(userID, "/portfolio")

	d := database.GetDB()
	d.Mu.RLock()
	var cases []models.CaseStudy
	for _, cs := range d.CaseStudies {
		if cs.UserID == userID && cs.IsPublic {
			cases = append(cases, cs)
		}
	}
	d.Mu.RUnlock()

	if cases == nil {
		cases = []models.CaseStudy{}
	}
	JSON(w, 200, cases)
}

func AttachCaseToGigHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/gigs/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		Error(w, 400, "Invalid path")
		return
	}
	gigID := parts[0]
	csID := parts[2]

	gig := store.FindGigByID(gigID)
	if gig == nil || gig.FreelancerID != user.ID {
		Error(w, 403, "Not your gig")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.CaseStudies {
		if d.CaseStudies[i].ID == csID && d.CaseStudies[i].UserID == user.ID {
			for _, id := range d.CaseStudies[i].LinkedGigIDs {
				if id == gigID {
					d.Mu.Unlock()
					Error(w, 409, "Already attached")
					return
				}
			}
			d.CaseStudies[i].LinkedGigIDs = append(d.CaseStudies[i].LinkedGigIDs, gigID)
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Case study attached to gig"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Case study not found")
}

func DetachCaseFromGigHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/gigs/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		Error(w, 400, "Invalid path")
		return
	}
	gigID := parts[0]
	csID := parts[2]

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.CaseStudies {
		if d.CaseStudies[i].ID == csID && d.CaseStudies[i].UserID == user.ID {
			var updated []string
			for _, id := range d.CaseStudies[i].LinkedGigIDs {
				if id != gigID {
					updated = append(updated, id)
				}
			}
			d.CaseStudies[i].LinkedGigIDs = updated
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Case study detached from gig"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Case study not found")
}

func SubmitTestimonialHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	csID := strings.TrimPrefix(r.URL.Path, "/api/v1/portfolio/cases/")
	csID = strings.TrimSuffix(csID, "/testimonial")

	var req struct {
		Rating int    `json:"rating"`
		Text   string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Rating < 1 || req.Rating > 5 {
		Error(w, 400, "Rating must be 1-5")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var cs *models.CaseStudy
	for i := range d.CaseStudies {
		if d.CaseStudies[i].ID == csID {
			cs = &d.CaseStudies[i]
			break
		}
	}
	d.Mu.RUnlock()

	if cs == nil {
		Error(w, 404, "Case study not found")
		return
	}

	testimonial := models.Testimonial{
		ID:          store.NewID(),
		CaseStudyID: csID,
		AuthorID:    user.ID,
		Rating:      req.Rating,
		Text:        req.Text,
		IsApproved:  false,
		CreatedAt:   store.Now(),
	}

	d.Mu.Lock()
	d.Testimonials = append(d.Testimonials, testimonial)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": testimonial.ID, "message": "Testimonial submitted"})
}
