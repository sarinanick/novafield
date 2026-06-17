package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"sort"
	"strings"
)

func CreateMeetingHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Title          string   `json:"title"`
		Description    string   `json:"description"`
		RoomID         string   `json:"roomId"`
		StartTime      string   `json:"startTime"`
		EndTime        string   `json:"endTime"`
		Recurring      bool     `json:"recurring"`
		RecurrenceRule string   `json:"recurrenceRule"`
		AttendeeIDs    []string `json:"attendeeIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		Error(w, 400, "Title is required")
		return
	}

	if req.AttendeeIDs == nil {
		req.AttendeeIDs = []string{}
	}

	meeting := models.Meeting{
		ID:             store.NewID(),
		Title:          req.Title,
		Description:    req.Description,
		OrganizerID:    user.ID,
		RoomID:         req.RoomID,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Recurring:      req.Recurring,
		RecurrenceRule: req.RecurrenceRule,
		AttendeeIDs:    req.AttendeeIDs,
		Status:         "scheduled",
		CreatedAt:      store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.Meetings = append(d.Meetings, meeting)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"id": meeting.ID, "message": "Meeting created"})
}

func GetMeetingsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var meetings []models.Meeting
	for _, m := range d.Meetings {
		if m.OrganizerID == user.ID || containsString(m.AttendeeIDs, user.ID) {
			meetings = append(meetings, m)
		}
	}
	d.Mu.RUnlock()

	for i := range meetings {
		organizer := store.FindUserByID(meetings[i].OrganizerID)
		if organizer != nil {
			pub := store.ToPublic(*organizer)
			meetings[i].Organizer = &pub
		}
	}

	sort.Slice(meetings, func(i, j int) bool {
		return meetings[i].StartTime > meetings[j].StartTime
	})

	if meetings == nil {
		meetings = []models.Meeting{}
	}
	JSON(w, 200, meetings)
}

func GetMeetingHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	id = strings.Split(id, "/")[0]

	d := database.GetDB()
	d.Mu.RLock()
	var found *models.Meeting
	for _, m := range d.Meetings {
		if m.ID == id {
			f := m
			found = &f
			break
		}
	}
	d.Mu.RUnlock()

	if found == nil {
		Error(w, 404, "Meeting not found")
		return
	}

	if found.OrganizerID != user.ID && !containsString(found.AttendeeIDs, user.ID) {
		Error(w, 403, "Not authorized")
		return
	}

	organizer := store.FindUserByID(found.OrganizerID)
	if organizer != nil {
		pub := store.ToPublic(*organizer)
		found.Organizer = &pub
	}

	JSON(w, 200, found)
}

func UpdateMeetingHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	id = strings.TrimSuffix(id, "/edit")

	d := database.GetDB()
	d.Mu.Lock()
	found := false
	for i := range d.Meetings {
		if d.Meetings[i].ID == id {
			if d.Meetings[i].OrganizerID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Only the organizer can update")
				return
			}
			var req struct {
				Title          string   `json:"title"`
				Description    string   `json:"description"`
				RoomID         string   `json:"roomId"`
				StartTime      string   `json:"startTime"`
				EndTime        string   `json:"endTime"`
				Recurring      *bool    `json:"recurring"`
				RecurrenceRule string   `json:"recurrenceRule"`
				AttendeeIDs    []string `json:"attendeeIds"`
				Status         string   `json:"status"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				d.Mu.Unlock()
				Error(w, 400, "Invalid request")
				return
			}
			if req.Title != "" {
				d.Meetings[i].Title = req.Title
			}
			if req.Description != "" {
				d.Meetings[i].Description = req.Description
			}
			if req.RoomID != "" {
				d.Meetings[i].RoomID = req.RoomID
			}
			if req.StartTime != "" {
				d.Meetings[i].StartTime = req.StartTime
			}
			if req.EndTime != "" {
				d.Meetings[i].EndTime = req.EndTime
			}
			if req.Recurring != nil {
				d.Meetings[i].Recurring = *req.Recurring
			}
			if req.RecurrenceRule != "" {
				d.Meetings[i].RecurrenceRule = req.RecurrenceRule
			}
			if req.AttendeeIDs != nil {
				d.Meetings[i].AttendeeIDs = req.AttendeeIDs
			}
			if req.Status != "" {
				d.Meetings[i].Status = req.Status
			}
			found = true
			break
		}
	}
	d.Mu.Unlock()

	if !found {
		Error(w, 404, "Meeting not found")
		return
	}
	d.Save()
	JSON(w, 200, H{"message": "Meeting updated"})
}

func DeleteMeetingHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")

	d := database.GetDB()
	d.Mu.Lock()
	found := false
	for i := range d.Meetings {
		if d.Meetings[i].ID == id {
			if d.Meetings[i].OrganizerID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Only the organizer can cancel")
				return
			}
			d.Meetings[i].Status = "cancelled"
			found = true
			break
		}
	}
	d.Mu.Unlock()

	if !found {
		Error(w, 404, "Meeting not found")
		return
	}
	d.Save()
	JSON(w, 200, H{"message": "Meeting cancelled"})
}

func JoinMeetingHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	id = strings.TrimSuffix(id, "/join")

	d := database.GetDB()
	d.Mu.Lock()
	found := false
	var meeting models.Meeting
	for i := range d.Meetings {
		if d.Meetings[i].ID == id {
			if d.Meetings[i].OrganizerID != user.ID && !containsString(d.Meetings[i].AttendeeIDs, user.ID) {
				d.Mu.Unlock()
				Error(w, 403, "You are not invited to this meeting")
				return
			}
			if d.Meetings[i].Status == "cancelled" {
				d.Mu.Unlock()
				Error(w, 400, "Meeting is cancelled")
				return
			}
			d.Meetings[i].Status = "active"
			meeting = d.Meetings[i]
			found = true
			break
		}
	}
	d.Mu.Unlock()

	if !found {
		Error(w, 404, "Meeting not found")
		return
	}
	d.Save()
	JSON(w, 200, H{"message": "Meeting joined", "roomId": meeting.RoomID})
}
