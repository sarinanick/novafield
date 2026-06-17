package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"sort"
	"strings"
)

func MatchFreelancersHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Skills      []string `json:"skills"`
		BudgetMin   float64  `json:"budgetMin"`
		BudgetMax   float64  `json:"budgetMax"`
		Timeline    string   `json:"timeline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		Error(w, 400, "Title is required")
		return
	}

	briefID := store.NewID()
	brief := models.ProjectBrief{
		ID:          briefID,
		ClientID:    user.ID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Skills:      req.Skills,
		BudgetMin:   req.BudgetMin,
		BudgetMax:   req.BudgetMax,
		Timeline:    req.Timeline,
		Status:      "active",
		CreatedAt:   store.Now(),
	}

	d := database.GetDB()
	d.Mu.RLock()
	var freelancers []models.User
	for _, u := range d.Users {
		if u.Role == "freelancer" {
			freelancers = append(freelancers, u)
		}
	}

	type scoredFreelancer struct {
		User    models.UserPublic
		Score   float64
		Reasons []models.MatchReason
	}

	var scored []scoredFreelancer
	for _, fl := range freelancers {
		score, reasons := calculateMatchScore(fl, req.Category, req.Skills, req.BudgetMin, req.BudgetMax, d)
		if score > 0.1 {
			scored = append(scored, scoredFreelancer{
				User:    store.ToPublic(fl),
				Score:   score,
				Reasons: reasons,
			})
		}
	}
	d.Mu.RUnlock()

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > 10 {
		scored = scored[:10]
	}

	now := store.Now()
	var results []models.MatchResult
	d.Mu.Lock()
	d.ProjectBriefs = append(d.ProjectBriefs, brief)
	for i, sf := range scored {
		result := models.MatchResult{
			ID:           store.NewID(),
			BriefID:      briefID,
			FreelancerID: sf.User.ID,
			Score:        sf.Score,
			Reasons:      sf.Reasons,
			Rank:         i + 1,
			Status:       "suggested",
			CreatedAt:    now,
		}
		results = append(results, result)
		d.MatchResults = append(d.MatchResults, result)
	}
	d.Mu.Unlock()
	d.Save()

	if results == nil {
		results = []models.MatchResult{}
	}

	JSON(w, 200, H{
		"briefId":   briefID,
		"matches":   results,
		"total":     len(results),
	})
}

func calculateMatchScore(fl models.User, category string, skills []string, budgetMin, budgetMax float64, d *database.FileDB) (float64, []models.MatchReason) {
	var totalWeight float64
	var reasons []models.MatchReason

	skillScore := 0.0
	if len(skills) > 0 {
		matched := 0
		for _, req := range skills {
			reqLower := strings.ToLower(req)
			for _, has := range fl.Skills {
				if strings.ToLower(has) == reqLower {
					matched++
					break
				}
			}
		}
		skillScore = float64(matched) / float64(len(skills))
		reasons = append(reasons, models.MatchReason{
			Factor: "skill_match",
			Weight: skillScore * 0.35,
			Detail: fmt.Sprintf("%d/%d skills matched", matched, len(skills)),
		})
	}
	totalWeight += 0.35

	ratingScore := fl.Rating / 5.0
	reasons = append(reasons, models.MatchReason{
		Factor: "rating",
		Weight: ratingScore * 0.25,
		Detail: "Rating: " + fmt.Sprintf("%.1f", fl.Rating),
	})
	totalWeight += 0.25

	budgetScore := 1.0
	if budgetMax > 0 && fl.HourlyRate > 0 {
		estimatedCost := fl.HourlyRate * 8
		if estimatedCost > budgetMax {
			budgetScore = budgetMax / estimatedCost * 0.5
		} else if estimatedCost >= budgetMin {
			budgetScore = 1.0
		} else {
			budgetScore = 0.7
		}
	}
	reasons = append(reasons, models.MatchReason{
		Factor: "budget_fit",
		Weight: budgetScore * 0.2,
		Detail: "Rate: $" + fmt.Sprintf("%.0f", fl.HourlyRate) + "/hr",
	})
	totalWeight += 0.2

	completedOrders := 0
	totalOrders := 0
	for _, o := range d.Orders {
		if o.SellerID == fl.ID {
			totalOrders++
			if o.Status == "completed" {
				completedOrders++
			}
		}
	}
	completionScore := 0.5
	if totalOrders > 0 {
		completionScore = float64(completedOrders) / float64(totalOrders)
	}
	reasons = append(reasons, models.MatchReason{
		Factor: "completion_rate",
		Weight: completionScore * 0.2,
		Detail: fmt.Sprintf("%d/%d orders completed", completedOrders, totalOrders),
	})
	totalWeight += 0.2

	score := (skillScore*0.35 + ratingScore*0.25 + budgetScore*0.2 + completionScore*0.2)
	score = math.Round(score*100) / 100

	_ = totalWeight
	return score, reasons
}

func GetMatchProjectsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "freelancer" {
		Error(w, 403, "Only freelancers can view matching projects")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var results []map[string]interface{}
	for _, mr := range d.MatchResults {
		if mr.FreelancerID == user.ID && mr.Status == "suggested" {
			var brief *models.ProjectBrief
			for i := range d.ProjectBriefs {
				if d.ProjectBriefs[i].ID == mr.BriefID {
					brief = &d.ProjectBriefs[i]
					break
				}
			}
			if brief != nil && brief.Status == "active" {
				results = append(results, map[string]interface{}{
					"match":  mr,
					"brief":  brief,
				})
			}
		}
	}
	d.Mu.RUnlock()

	if results == nil {
		results = []map[string]interface{}{}
	}
	JSON(w, 200, results)
}

func MatchFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		MatchID string `json:"matchId"`
		Action  string `json:"action"` // "viewed", "contacted", "hired", "dismissed"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MatchID == "" {
		Error(w, 400, "Match ID and action required")
		return
	}

	validActions := map[string]bool{
		"viewed": true, "contacted": true, "hired": true, "dismissed": true,
	}
	if !validActions[req.Action] {
		Error(w, 400, "Invalid action")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.MatchResults {
		if d.MatchResults[i].ID == req.MatchID {
			d.MatchResults[i].Status = req.Action
			break
		}
	}
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Feedback recorded"})
}

func MatchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var results []models.MatchResult
	if user.Role == "admin" {
		results = d.MatchResults
	} else {
		for _, mr := range d.MatchResults {
			brief := store.FindUserByID(mr.FreelancerID)
			if brief != nil && (mr.FreelancerID == user.ID) {
				results = append(results, mr)
			}
		}
	}
	d.Mu.RUnlock()

	if results == nil {
		results = []models.MatchResult{}
	}
	JSON(w, 200, results)
}
