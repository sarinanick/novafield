package handlers

import (
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func GetDesksHandler(w http.ResponseWriter, r *http.Request) {
	d := database.GetDB()
	d.Mu.RLock()
	desks := make([]models.Desk, len(d.Desks))
	copy(desks, d.Desks)
	d.Mu.RUnlock()

	for i := range desks {
		if desks[i].OwnerID != "" {
			user := store.FindUserByID(desks[i].OwnerID)
			if user != nil {
				pub := store.ToPublic(*user)
				desks[i].Owner = &pub
			}
		}
		if desks[i].Objects == nil {
			desks[i].Objects = []models.DeskObject{}
		}
	}
	JSON(w, 200, desks)
}

func ClaimDeskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		Error(w, 405, "Method not allowed")
		return
	}
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/desks/")
	id = strings.TrimSuffix(id, "/claim")

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Desks {
		if d.Desks[i].ID == id {
			if d.Desks[i].OwnerID != "" && d.Desks[i].OwnerID != user.ID {
				d.Mu.Unlock()
				Error(w, 409, "Desk is already claimed")
				return
			}
			if d.Desks[i].IsLocked && d.Desks[i].OwnerID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Desk is locked")
				return
			}
			d.Desks[i].OwnerID = user.ID
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Desk claimed"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Desk not found")
}

func UnclaimDeskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		Error(w, 405, "Method not allowed")
		return
	}
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/desks/")
	id = strings.TrimSuffix(id, "/unclaim")

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Desks {
		if d.Desks[i].ID == id {
			if d.Desks[i].OwnerID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Not your desk")
				return
			}
			d.Desks[i].OwnerID = ""
			d.Desks[i].Objects = []models.DeskObject{}
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Desk unclaimed"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Desk not found")
}

func UpdateDeskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		Error(w, 405, "Method not allowed")
		return
	}
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/desks/")

	var req struct {
		Color   string             `json:"color"`
		Objects []models.DeskObject `json:"objects"`
	}
	if err := Decode(r, &req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Desks {
		if d.Desks[i].ID == id {
			if d.Desks[i].OwnerID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Not your desk")
				return
			}
			if req.Color != "" {
				d.Desks[i].Color = req.Color
			}
			if req.Objects != nil {
				d.Desks[i].Objects = req.Objects
			}
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Desk updated"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Desk not found")
}

func LockDeskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		Error(w, 405, "Method not allowed")
		return
	}
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/desks/")
	id = strings.TrimSuffix(id, "/lock")

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Desks {
		if d.Desks[i].ID == id {
			if d.Desks[i].OwnerID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Not your desk")
				return
			}
			d.Desks[i].IsLocked = !d.Desks[i].IsLocked
			locked := d.Desks[i].IsLocked
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"locked": locked})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Desk not found")
}

func DesksRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/desks" || r.URL.Path == "/api/v1/desks/" {
		switch r.Method {
		case "GET":
			GetDesksHandler(w, r)
		default:
			Error(w, 405, "Method not allowed")
		}
		return
	}
}

func DeskActionRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	trimmed := strings.TrimPrefix(path, "/api/v1/desks/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		Error(w, 404, "Not found")
		return
	}

	action := parts[1]
	switch action {
	case "claim":
		ClaimDeskHandler(w, r)
	case "unclaim":
		UnclaimDeskHandler(w, r)
	case "lock":
		LockDeskHandler(w, r)
	default:
		if r.Method == "PUT" {
			UpdateDeskHandler(w, r)
		} else {
			Error(w, 404, "Not found")
		}
	}
}
