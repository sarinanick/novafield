package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

var defaultAssessments = []models.Assessment{
	{
		ID: "assess-gpt4", Category: "ai-chatbots", Title: "GPT-4 & Prompt Engineering",
		Description: "Test your knowledge of GPT-4, prompt engineering, and AI chatbot development",
		PassScore: 70,
		Questions: []models.AssessmentQuestion{
			{ID: "q1", Question: "What is the maximum context window of GPT-4 Turbo?", Options: []string{"8K tokens", "32K tokens", "128K tokens", "256K tokens"}, Correct: 2},
			{ID: "q2", Question: "Which technique helps reduce hallucinations in LLM outputs?", Options: []string{"Chain-of-thought", "Temperature=2", "Top-p=0.1", "Max tokens=100"}, Correct: 0},
			{ID: "q3", Question: "What is the purpose of a system prompt?", Options: []string{"Speed up inference", "Define model behavior and context", "Reduce token usage", "Enable function calling"}, Correct: 1},
			{ID: "q4", Question: "Which is NOT a valid OpenAI API parameter?", Options: []string{"temperature", "top_p", "frequency_penalty", "creativity_boost"}, Correct: 3},
			{ID: "q5", Question: "What is Retrieval-Augmented Generation (RAG)?", Options: []string{"A new model architecture", "Combining LLMs with external knowledge", "A training technique", "A prompt template"}, Correct: 1},
		},
	},
	{
		ID: "assess-midjourney", Category: "ai-image", Title: "Midjourney & AI Image Generation",
		Description: "Test your expertise in AI image generation tools and techniques",
		PassScore: 70,
		Questions: []models.AssessmentQuestion{
			{ID: "q1", Question: "What does the --ar parameter control in Midjourney?", Options: []string{"Art style", "Aspect ratio", "Animation rate", "Alpha rendering"}, Correct: 1},
			{ID: "q2", Question: "Which is a valid Midjourney version flag?", Options: []string{"--v 5", "--version 5", "--model v5", "--gen 5"}, Correct: 0},
			{ID: "q3", Question: "What does --stylize control?", Options: []string{"Image resolution", "How strongly Midjourney's style is applied", "Color saturation", "Noise level"}, Correct: 1},
			{ID: "q4", Question: "Which format is best for AI-generated images with transparency?", Options: []string{"JPEG", "PNG", "BMP", "TIFF"}, Correct: 1},
			{ID: "q5", Question: "What is img2img?", Options: []string{"Converting images to text", "Using an image as a reference for generation", "Image compression", "Image annotation"}, Correct: 1},
		},
	},
	{
		ID: "assess-sora", Category: "ai-video", Title: "Sora & AI Video Generation",
		Description: "Test your knowledge of AI video generation tools",
		PassScore: 70,
		Questions: []models.AssessmentQuestion{
			{ID: "q1", Question: "What is Sora primarily designed for?", Options: []string{"Image generation", "Text generation", "Video generation", "Audio generation"}, Correct: 2},
			{ID: "q2", Question: "What is the maximum video length Sora can typically generate?", Options: []string{"5 seconds", "30 seconds", "60 seconds", "5 minutes"}, Correct: 2},
			{ID: "q3", Question: "Which is a key challenge in AI video generation?", Options: []string{"Color accuracy", "Temporal consistency between frames", "File size", "Audio sync"}, Correct: 1},
			{ID: "q4", Question: "What is prompt engineering for video?", Options: []string{"Editing video code", "Crafting detailed text descriptions for video generation", "Video compression", "Video annotation"}, Correct: 1},
			{ID: "q5", Question: "Which factor most affects AI video quality?", Options: []string{"GPU speed", "Prompt detail and specificity", "File format", "Internet speed"}, Correct: 1},
		},
	},
}

func ListAssessmentsHandler(w http.ResponseWriter, r *http.Request) {
	JSON(w, 200, defaultAssessments)
}

func GetAssessmentHandler(w http.ResponseWriter, r *http.Request) {
	assessmentID := strings.TrimPrefix(r.URL.Path, "/api/v1/assessments/")

	for _, a := range defaultAssessments {
		if a.ID == assessmentID {
			JSON(w, 200, a)
			return
		}
	}
	Error(w, 404, "Assessment not found")
}

func StartAssessmentHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	assessmentID := strings.TrimPrefix(r.URL.Path, "/api/v1/assessments/")
	assessmentID = strings.TrimSuffix(assessmentID, "/start")

	var assessment *models.Assessment
	for i := range defaultAssessments {
		if defaultAssessments[i].ID == assessmentID {
			assessment = &defaultAssessments[i]
			break
		}
	}
	if assessment == nil {
		Error(w, 404, "Assessment not found")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, result := range d.AssessmentResults {
		if result.UserID == user.ID && result.AssessmentID == assessmentID && result.Passed {
			d.Mu.RUnlock()
			JSON(w, 200, H{"message": "Already passed this assessment", "score": result.Score})
			return
		}
	}
	d.Mu.RUnlock()

	questions := make([]map[string]interface{}, len(assessment.Questions))
	for i, q := range assessment.Questions {
		questions[i] = map[string]interface{}{
			"id":       q.ID,
			"question": q.Question,
			"options":  q.Options,
		}
	}

	JSON(w, 200, H{
		"assessmentId": assessment.ID,
		"title":        assessment.Title,
		"questions":    questions,
	})
}

func SubmitAssessmentHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	assessmentID := strings.TrimPrefix(r.URL.Path, "/api/v1/assessments/")
	assessmentID = strings.TrimSuffix(assessmentID, "/submit")

	var assessment *models.Assessment
	for i := range defaultAssessments {
		if defaultAssessments[i].ID == assessmentID {
			assessment = &defaultAssessments[i]
			break
		}
	}
	if assessment == nil {
		Error(w, 404, "Assessment not found")
		return
	}

	var req struct {
		Answers []int `json:"answers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Answers) != len(assessment.Questions) {
		Error(w, 400, "Must provide answers for all questions")
		return
	}

	correct := 0
	for i, q := range assessment.Questions {
		if req.Answers[i] == q.Correct {
			correct++
		}
	}

	score := correct * 100 / len(assessment.Questions)
	passed := score >= assessment.PassScore

	result := models.AssessmentResult{
		ID:           store.NewID(),
		UserID:       user.ID,
		AssessmentID: assessmentID,
		Score:        score,
		Passed:       passed,
		Answers:      req.Answers,
		CompletedAt:  store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.AssessmentResults = append(d.AssessmentResults, result)

	if passed {
		level := "verified"
		if score >= 90 {
			level = "expert"
		}
		badge := models.Badge{
			ID:       store.NewID(),
			UserID:   user.ID,
			Skill:    assessment.Category,
			Level:    level,
			Score:    score,
			EarnedAt: store.Now(),
		}
		d.Badges = append(d.Badges, badge)
	}
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{
		"score":  score,
		"passed": passed,
		"correct": correct,
		"total":  len(assessment.Questions),
	})
}

func GetAssessmentResultsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var results []models.AssessmentResult
	for _, result := range d.AssessmentResults {
		if result.UserID == user.ID {
			results = append(results, result)
		}
	}
	d.Mu.RUnlock()

	if results == nil {
		results = []models.AssessmentResult{}
	}
	JSON(w, 200, results)
}

func GetUserBadgesHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	userID = strings.TrimSuffix(userID, "/badges")

	d := database.GetDB()
	d.Mu.RLock()
	var badges []models.Badge
	for _, b := range d.Badges {
		if b.UserID == userID {
			badges = append(badges, b)
		}
	}
	d.Mu.RUnlock()

	if badges == nil {
		badges = []models.Badge{}
	}
	JSON(w, 200, badges)
}

func GetSkillLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimPrefix(r.URL.Path, "/api/v1/assessments/leaderboard/")

	d := database.GetDB()
	d.Mu.RLock()

	type leaderboardEntry struct {
		UserID string `json:"userId"`
		Name   string `json:"name"`
		Score  int    `json:"score"`
		Level  string `json:"level"`
	}

	var entries []leaderboardEntry
	for _, b := range d.Badges {
		if category == "" || b.Skill == category {
			user := store.FindUserByID(b.UserID)
			name := "Unknown"
			if user != nil {
				name = user.Name
			}
			entries = append(entries, leaderboardEntry{
				UserID: b.UserID,
				Name:   name,
				Score:  b.Score,
				Level:  b.Level,
			})
		}
	}
	d.Mu.RUnlock()

	if entries == nil {
		entries = []leaderboardEntry{}
	}
	JSON(w, 200, entries)
}
