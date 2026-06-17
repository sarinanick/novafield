package handlers

import (
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"strings"
)

func GetTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	d := database.GetDB()
	d.Mu.RLock()
	templates := make([]models.OfficeTemplate, len(d.Templates))
	copy(templates, d.Templates)
	d.Mu.RUnlock()
	JSON(w, 200, templates)
}

func GetTemplateHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/templates/")
	if id == "" || id == r.URL.Path {
		id = strings.TrimPrefix(r.URL.Path, "/api/v1/templates")
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, t := range d.Templates {
		if t.ID == id {
			d.Mu.RUnlock()
			JSON(w, 200, t)
			return
		}
	}
	d.Mu.RUnlock()
	Error(w, 404, "Template not found")
}

func ApplyTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	workspaceID := strings.TrimPrefix(r.URL.Path, "/api/v1/workspaces/")
	workspaceID = strings.TrimSuffix(workspaceID, "/apply-template")

	var req struct {
		TemplateID string `json:"templateId"`
	}
	if err := Decode(r, &req); err != nil || req.TemplateID == "" {
		Error(w, 400, "templateId is required")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var template *models.OfficeTemplate
	for _, t := range d.Templates {
		if t.ID == req.TemplateID {
			tt := t
			template = &tt
			break
		}
	}
	d.Mu.RUnlock()

	if template == nil {
		Error(w, 404, "Template not found")
		return
	}

	_ = workspaceID

	JSON(w, 200, H{
		"message":  "Template applied successfully",
		"template": template,
	})
}

func TemplatesRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/v1/templates" || path == "/api/v1/templates/" {
		switch r.Method {
		case "GET":
			GetTemplatesHandler(w, r)
		default:
			Error(w, 405, "Method not allowed")
		}
		return
	}

	id := strings.TrimPrefix(path, "/api/v1/templates/")
	GetTemplateHandler(w, r)
	_ = id
}

func WorkspaceRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	trimmed := strings.TrimPrefix(path, "/api/v1/workspaces/")

	parts := strings.Split(trimmed, "/")
	if len(parts) >= 2 && parts[1] == "apply-template" {
		ApplyTemplateHandler(w, r)
		return
	}

	Error(w, 404, "Not found")
}
