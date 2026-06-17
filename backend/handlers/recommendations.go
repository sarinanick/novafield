package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"sort"
	"strings"
)

func GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	categoryScores := make(map[string]float64)
	for _, o := range d.Orders {
		if o.BuyerID == user.ID {
			gig := store.FindGigByID(o.GigID)
			if gig != nil {
				categoryScores[gig.Category] += 2.0
			}
		}
	}

	for _, g := range d.Gigs {
		categoryScores[g.Category] += 0.1
	}

	type scoredGig struct {
		Gig     models.Gig
		Score   float64
		Reasons []string
	}

	var scored []scoredGig
	for _, g := range d.Gigs {
		if g.Status != "active" || g.FreelancerID == user.ID {
			continue
		}

		ordered := false
		for _, o := range d.Orders {
			if o.BuyerID == user.ID && o.GigID == g.ID {
				ordered = true
				break
			}
		}
		if ordered {
			continue
		}

		score := 0.0
		var reasons []string

		catScore := categoryScores[g.Category]
		if catScore > 0 {
			score += math.Min(catScore/10.0, 0.4)
			reasons = append(reasons, "category_match")
		}

		if g.Rating >= 4.0 {
			score += 0.2
			reasons = append(reasons, "high_rating")
		}

		if g.OrdersCount > 5 {
			score += 0.15
			reasons = append(reasons, "popular")
		}

		if g.Featured {
			score += 0.1
			reasons = append(reasons, "featured")
		}

		rating, _ := store.CalcGigRating(g.ID)
		if rating >= 4.5 {
			score += 0.15
			reasons = append(reasons, "top_rated")
		}

		if score > 0.1 {
			scored = append(scored, scoredGig{Gig: g, Score: math.Round(score*100)/100, Reasons: reasons})
		}
	}
	d.Mu.RUnlock()

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	limit := 12
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := parseLimit(v); err == nil && n > 0 {
			limit = n
		}
	}
	if len(scored) > limit {
		scored = scored[:limit]
	}

	var results []map[string]interface{}
	for _, sg := range scored {
		freelancer := store.FindUserByID(sg.Gig.FreelancerID)
		var pub interface{}
		if freelancer != nil {
			pub = store.ToPublic(*freelancer)
		}
		results = append(results, map[string]interface{}{
			"gig":        sg.Gig,
			"freelancer": pub,
			"score":      sg.Score,
			"reasons":    sg.Reasons,
		})
	}
	if results == nil {
		results = []map[string]interface{}{}
	}

	JSON(w, 200, H{"recommendations": results, "total": len(results)})
}

func parseLimit(v string) (int, error) {
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func GetSimilarGigsHandler(w http.ResponseWriter, r *http.Request) {
	gigID := strings.TrimPrefix(r.URL.Path, "/api/v1/recommendations/similar/")
	gigID = strings.TrimSuffix(gigID, "/")

	gig := store.FindGigByID(gigID)
	if gig == nil {
		Error(w, 404, "Gig not found")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	type scoredGig struct {
		Gig   models.Gig
		Score float64
	}

	var scored []scoredGig
	for _, g := range d.Gigs {
		if g.ID == gigID || g.Status != "active" {
			continue
		}

		score := 0.0
		if g.Category == gig.Category {
			score += 0.5
		}

		for _, tag := range gig.Tags {
			for _, otherTag := range g.Tags {
				if strings.EqualFold(tag, otherTag) {
					score += 0.2
					break
				}
			}
		}

		if g.PriceType == gig.PriceType {
			score += 0.1
		}

		priceDiff := math.Abs(g.Price - gig.Price)
		if priceDiff < gig.Price*0.5 {
			score += 0.2
		}

		if score > 0.3 {
			scored = append(scored, scoredGig{Gig: g, Score: score})
		}
	}
	d.Mu.RUnlock()

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > 6 {
		scored = scored[:6]
	}

	var results []map[string]interface{}
	for _, sg := range scored {
		freelancer := store.FindUserByID(sg.Gig.FreelancerID)
		var pub interface{}
		if freelancer != nil {
			pub = store.ToPublic(*freelancer)
		}
		results = append(results, map[string]interface{}{
			"gig":        sg.Gig,
			"freelancer": pub,
			"score":      sg.Score,
		})
	}
	if results == nil {
		results = []map[string]interface{}{}
	}

	JSON(w, 200, H{"similar": results})
}

func RecommendationFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		GigID     string `json:"gigId"`
		EventType string `json:"eventType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GigID == "" {
		Error(w, 400, "Gig ID is required")
		return
	}

	validEvents := map[string]bool{
		"impression": true, "click": true, "dismiss": true, "order": true,
	}
	if !validEvents[req.EventType] {
		Error(w, 400, "Invalid event type")
		return
	}

	feedback := models.RecommendationFeedback{
		ID:        store.NewID(),
		UserID:    user.ID,
		GigID:     req.GigID,
		EventType: req.EventType,
		CreatedAt: store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.RecFeedback = append(d.RecFeedback, feedback)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Feedback recorded"})
}
