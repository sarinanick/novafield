package handlers

import (
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func GetFloorsHandler(w http.ResponseWriter, r *http.Request) {
	d := database.GetDB()
	d.Mu.RLock()
	floors := make([]models.Floor, len(d.Floors))
	copy(floors, d.Floors)
	d.Mu.RUnlock()

	if floors == nil {
		floors = []models.Floor{}
	}
	JSON(w, 200, floors)
}

func GetFloorHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/floors/")

	d := database.GetDB()
	d.Mu.RLock()
	defer d.Mu.RUnlock()

	for _, f := range d.Floors {
		if f.ID == id {
			JSON(w, 200, f)
			return
		}
	}
	Error(w, 404, "Floor not found")
}

func CreateFloorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Only admins can create floors")
		return
	}

	var req struct {
		Name   string         `json:"name"`
		Level  int            `json:"level"`
		Zones  []models.Zone  `json:"zones"`
		Width  float64        `json:"width"`
		Height float64        `json:"height"`
	}
	if err := Decode(r, &req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	if req.Name == "" {
		Error(w, 400, "Floor name is required")
		return
	}
	if req.Width == 0 {
		req.Width = 800
	}
	if req.Height == 0 {
		req.Height = 600
	}

	floor := models.Floor{
		ID:        store.NewID(),
		Name:      req.Name,
		Level:     req.Level,
		Zones:     req.Zones,
		Width:     req.Width,
		Height:    req.Height,
		IsDefault: false,
		CreatedAt: store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.Floors = append(d.Floors, floor)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, floor)
}

func UpdateFloorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Only admins can update floors")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/floors/")
	id = strings.TrimSuffix(id, "/")

	var req struct {
		Name   string         `json:"name"`
		Level  int            `json:"level"`
		Zones  []models.Zone  `json:"zones"`
		Width  float64        `json:"width"`
		Height float64        `json:"height"`
	}
	if err := Decode(r, &req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	defer d.Mu.Unlock()

	for i := range d.Floors {
		if d.Floors[i].ID == id {
			if req.Name != "" {
				d.Floors[i].Name = req.Name
			}
			if req.Level != 0 {
				d.Floors[i].Level = req.Level
			}
			if req.Zones != nil {
				d.Floors[i].Zones = req.Zones
			}
			if req.Width != 0 {
				d.Floors[i].Width = req.Width
			}
			if req.Height != 0 {
				d.Floors[i].Height = req.Height
			}
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, d.Floors[i])
			return
		}
	}
	Error(w, 404, "Floor not found")
}

func DeleteFloorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Only admins can delete floors")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/floors/")
	id = strings.TrimSuffix(id, "/")

	d := database.GetDB()
	d.Mu.Lock()
	defer d.Mu.Unlock()

	for i := range d.Floors {
		if d.Floors[i].ID == id {
			if d.Floors[i].IsDefault {
				Error(w, 400, "Cannot delete the default floor")
				return
			}
			d.Floors = append(d.Floors[:i], d.Floors[i+1:]...)
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Floor deleted"})
			return
		}
	}
	Error(w, 404, "Floor not found")
}

func SwitchFloorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		Error(w, 405, "Method not allowed")
		return
	}

	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/floors/")
	id = strings.TrimSuffix(id, "/switch")

	d := database.GetDB()
	d.Mu.RLock()
	floorExists := false
	for _, f := range d.Floors {
		if f.ID == id {
			floorExists = true
			break
		}
	}
	d.Mu.RUnlock()

	if !floorExists {
		Error(w, 404, "Floor not found")
		return
	}

	JSON(w, 200, H{"floorId": id, "message": "Floor switched"})
}

func FloorsRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/floors" || r.URL.Path == "/api/v1/floors/" {
		switch r.Method {
		case "GET":
			GetFloorsHandler(w, r)
		case "POST":
			CreateFloorHandler(w, r)
		default:
			Error(w, 405, "Method not allowed")
		}
		return
	}
}

func FloorActionRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	trimmed := strings.TrimPrefix(path, "/api/v1/floors/")

	if strings.HasSuffix(trimmed, "/switch") {
		SwitchFloorHandler(w, r)
		return
	}

	switch r.Method {
	case "GET":
		GetFloorHandler(w, r)
	case "PUT":
		UpdateFloorHandler(w, r)
	case "DELETE":
		DeleteFloorHandler(w, r)
	default:
		Error(w, 405, "Method not allowed")
	}
}
