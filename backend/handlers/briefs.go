package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func CreateBriefHandler(w http.ResponseWriter, r *http.Request) {
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

	brief := models.ClientBrief{
		ID:          store.NewID(),
		ClientID:    user.ID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Skills:      req.Skills,
		BudgetMin:   req.BudgetMin,
		BudgetMax:   req.BudgetMax,
		Timeline:    req.Timeline,
		Status:      "open",
		CreatedAt:   store.Now(),
	}

	if brief.Skills == nil {
		brief.Skills = []string{}
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.ClientBriefs = append(d.ClientBriefs, brief)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": brief.ID, "message": "Project brief created"})
}

func ListBriefsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var briefs []models.ClientBrief
	for _, b := range d.ClientBriefs {
		if b.Status == "open" {
			briefs = append(briefs, b)
		}
	}
	d.Mu.RUnlock()

	if briefs == nil {
		briefs = []models.ClientBrief{}
	}
	JSON(w, 200, briefs)
}

func GetBriefHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	briefID := strings.TrimPrefix(r.URL.Path, "/api/v1/briefs/")

	d := database.GetDB()
	d.Mu.RLock()
	var brief *models.ClientBrief
	for i := range d.ClientBriefs {
		if d.ClientBriefs[i].ID == briefID {
			brief = &d.ClientBriefs[i]
			break
		}
	}
	d.Mu.RUnlock()

	if brief == nil {
		Error(w, 404, "Brief not found")
		return
	}

	JSON(w, 200, brief)
}

func SubmitProposalHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "freelancer" {
		Error(w, 403, "Only freelancers can submit proposals")
		return
	}

	briefID := strings.TrimPrefix(r.URL.Path, "/api/v1/briefs/")
	briefID = strings.TrimSuffix(briefID, "/proposals")

	var req struct {
		CoverLetter string  `json:"coverLetter"`
		Price       float64 `json:"price"`
		Timeline    string  `json:"timeline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CoverLetter == "" {
		Error(w, 400, "Cover letter is required")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var brief *models.ClientBrief
	for i := range d.ClientBriefs {
		if d.ClientBriefs[i].ID == briefID {
			brief = &d.ClientBriefs[i]
			break
		}
	}
	d.Mu.RUnlock()

	if brief == nil || brief.Status != "open" {
		Error(w, 404, "Brief not found or not open")
		return
	}

	for _, p := range d.Proposals {
		if p.BriefID == briefID && p.FreelancerID == user.ID {
			Error(w, 409, "Already submitted a proposal")
			return
		}
	}

	proposal := models.Proposal{
		ID:           store.NewID(),
		BriefID:      briefID,
		FreelancerID: user.ID,
		CoverLetter:  req.CoverLetter,
		Price:        req.Price,
		Timeline:     req.Timeline,
		Status:       "pending",
		CreatedAt:    store.Now(),
	}

	d.Mu.Lock()
	d.Proposals = append(d.Proposals, proposal)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": proposal.ID, "message": "Proposal submitted"})
}

func ListProposalsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	briefID := strings.TrimPrefix(r.URL.Path, "/api/v1/briefs/")
	briefID = strings.TrimSuffix(briefID, "/proposals")

	d := database.GetDB()
	d.Mu.RLock()
	var brief *models.ClientBrief
	for i := range d.ClientBriefs {
		if d.ClientBriefs[i].ID == briefID {
			brief = &d.ClientBriefs[i]
			break
		}
	}
	d.Mu.RUnlock()

	if brief == nil || brief.ClientID != user.ID {
		Error(w, 403, "Not authorized")
		return
	}

	d.Mu.RLock()
	var proposals []models.Proposal
	for _, p := range d.Proposals {
		if p.BriefID == briefID {
			proposals = append(proposals, p)
		}
	}
	d.Mu.RUnlock()

	if proposals == nil {
		proposals = []models.Proposal{}
	}
	JSON(w, 200, proposals)
}

func AcceptProposalHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/briefs/")
	path = strings.TrimSuffix(path, "/accept")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		Error(w, 400, "Invalid path")
		return
	}
	briefID := parts[0]
	proposalID := parts[2]

	d := database.GetDB()
	d.Mu.RLock()
	var brief *models.ClientBrief
	for i := range d.ClientBriefs {
		if d.ClientBriefs[i].ID == briefID {
			brief = &d.ClientBriefs[i]
			break
		}
	}
	d.Mu.RUnlock()

	if brief == nil || brief.ClientID != user.ID {
		Error(w, 403, "Not authorized")
		return
	}

	d.Mu.Lock()
	for i := range d.Proposals {
		if d.Proposals[i].ID == proposalID && d.Proposals[i].BriefID == briefID {
			d.Proposals[i].Status = "accepted"

			for j := range d.ClientBriefs {
				if d.ClientBriefs[j].ID == briefID {
					d.ClientBriefs[j].Status = "in_progress"
					break
				}
			}

			for j := range d.Proposals {
				if d.Proposals[j].BriefID == briefID && d.Proposals[j].ID != proposalID {
					d.Proposals[j].Status = "rejected"
				}
			}

			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Proposal accepted", "freelancerId": d.Proposals[i].FreelancerID})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Proposal not found")
}

func ListMyBriefsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var briefs []models.ClientBrief
	for _, b := range d.ClientBriefs {
		if b.ClientID == user.ID {
			briefs = append(briefs, b)
		}
	}
	d.Mu.RUnlock()

	if briefs == nil {
		briefs = []models.ClientBrief{}
	}
	JSON(w, 200, briefs)
}
