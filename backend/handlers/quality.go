package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func ScoreGigHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Tags        []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		Error(w, 400, "Title is required")
		return
	}

	titleScore := scoreTitle(req.Title)
	descScore := scoreDescription(req.Description)
	seoScore := scoreSEO(req.Title, req.Description, req.Tags)
	tagsScore := scoreTags(req.Tags, req.Category)

	totalScore := (titleScore*0.25 + descScore*0.35 + seoScore*0.2 + tagsScore*0.2)
	totalScore = math.Round(totalScore*10) / 10

	var suggestions []string
	if titleScore < 7 {
		suggestions = append(suggestions, "Make your title more specific and include key deliverables")
	}
	if descScore < 7 {
		suggestions = append(suggestions, "Add more detail to your description, including process, deliverables, and revisions")
	}
	if seoScore < 7 {
		suggestions = append(suggestions, "Include relevant keywords in your title and description")
	}
	if tagsScore < 7 {
		suggestions = append(suggestions, "Add more relevant tags that match what buyers search for")
	}
	if len(req.Tags) < 3 {
		suggestions = append(suggestions, "Add at least 3-5 tags for better discoverability")
	}

	qualityScore := models.QualityScore{
		GigID:       store.NewID(),
		Score:       totalScore,
		TitleScore:  titleScore,
		DescScore:   descScore,
		SEOScore:    seoScore,
		TagsScore:   tagsScore,
		Suggestions: suggestions,
		ScoredAt:    store.Now(),
	}

	JSON(w, 200, qualityScore)
}

func ScoreExistingGigHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	gigID := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/gig/")
	gigID = strings.TrimSuffix(gigID, "/score")

	gig := store.FindGigByID(gigID)
	if gig == nil {
		Error(w, 404, "Gig not found")
		return
	}
	if gig.FreelancerID != user.ID && user.Role != "admin" {
		Error(w, 403, "Not authorized")
		return
	}

	titleScore := scoreTitle(gig.Title)
	descScore := scoreDescription(gig.Description)
	seoScore := scoreSEO(gig.Title, gig.Description, gig.Tags)
	tagsScore := scoreTags(gig.Tags, gig.Category)

	totalScore := (titleScore*0.25 + descScore*0.35 + seoScore*0.2 + tagsScore*0.2)
	totalScore = math.Round(totalScore*10) / 10

	var suggestions []string
	if titleScore < 7 {
		suggestions = append(suggestions, "Make your title more specific and include key deliverables")
	}
	if descScore < 7 {
		suggestions = append(suggestions, "Add more detail to your description")
	}
	if seoScore < 7 {
		suggestions = append(suggestions, "Include relevant keywords in your title and description")
	}
	if tagsScore < 7 {
		suggestions = append(suggestions, "Add more relevant tags")
	}

	qualityScore := models.QualityScore{
		GigID:       gigID,
		Score:       totalScore,
		TitleScore:  titleScore,
		DescScore:   descScore,
		SEOScore:    seoScore,
		TagsScore:   tagsScore,
		Suggestions: suggestions,
		ScoredAt:    store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.QualityScores = append(d.QualityScores, qualityScore)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, qualityScore)
}

func scoreTitle(title string) float64 {
	score := 5.0
	length := len(title)

	if length >= 20 && length <= 80 {
		score += 2
	} else if length >= 10 {
		score += 1
	}

	lower := strings.ToLower(title)
	keywords := []string{"i will", "create", "design", "build", "professional", "expert", "custom"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			score += 0.5
		}
	}

	if strings.Contains(title, "!") || strings.Contains(title, "|") {
		score += 0.5
	}

	return math.Min(10, score)
}

func scoreDescription(desc string) float64 {
	score := 3.0
	length := len(desc)

	if length >= 200 {
		score += 3
	} else if length >= 100 {
		score += 2
	} else if length >= 50 {
		score += 1
	}

	if strings.Contains(desc, "✅") || strings.Contains(desc, "•") || strings.Contains(desc, "-") {
		score += 1
	}

	if strings.Contains(strings.ToLower(desc), "revision") {
		score += 0.5
	}
	if strings.Contains(strings.ToLower(desc), "delivery") || strings.Contains(strings.ToLower(desc), "deliver") {
		score += 0.5
	}
	if strings.Contains(strings.ToLower(desc), "guarantee") || strings.Contains(strings.ToLower(desc), "satisfaction") {
		score += 0.5
	}

	return math.Min(10, score)
}

func scoreSEO(title, desc string, tags []string) float64 {
	score := 4.0

	lowerTitle := strings.ToLower(title)

	if len(tags) >= 5 {
		score += 2
	} else if len(tags) >= 3 {
		score += 1
	}

	for _, tag := range tags {
		if strings.Contains(lowerTitle, strings.ToLower(tag)) {
			score += 0.5
			break
		}
	}

	if strings.Contains(lowerTitle, "i will") {
		score += 1
	}

	if len(desc) > 100 {
		score += 1
	}

	return math.Min(10, score)
}

func scoreTags(tags []string, category string) float64 {
	score := 3.0

	if len(tags) >= 5 {
		score += 3
	} else if len(tags) >= 3 {
		score += 2
	} else if len(tags) >= 1 {
		score += 1
	}

	if category != "" {
		score += 1
	}

	uniqueTags := make(map[string]bool)
	for _, t := range tags {
		uniqueTags[strings.ToLower(t)] = true
	}
	if len(uniqueTags) == len(tags) && len(tags) > 0 {
		score += 1
	}

	return math.Min(10, score)
}
