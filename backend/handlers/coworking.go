package handlers

import (
	"log"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
	"time"
)

const (
	maxCoworkingParticipants = 20
	pomodoroWorkMinutes      = 25
	pomodoroBreakMinutes     = 5
)

func CreateCoworkingSession(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Type     string `json:"type"`
		Title    string `json:"title"`
		ZoneID   string `json:"zoneId"`
		Duration int    `json:"duration"`
	}
	if err := Decode(r, &req); err != nil || req.Title == "" {
		Error(w, 400, "Title is required")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if len(req.Title) > 100 {
		Error(w, 400, "Title must be 100 characters or less")
		return
	}

	switch req.Type {
	case "focused", "pomodoro", "casual":
	default:
		req.Type = "focused"
	}
	if req.Duration <= 0 || req.Duration > 480 {
		req.Duration = 25
	}
	if req.ZoneID == "" {
		req.ZoneID = "work"
	}

	totalSeconds := req.Duration * 60
	phase := "work"
	if req.Type == "pomodoro" {
		totalSeconds = pomodoroWorkMinutes * 60
	}

	session := models.CoworkingSession{
		ID:             store.NewID(),
		HostID:         user.ID,
		Type:           req.Type,
		Title:          req.Title,
		ZoneID:         req.ZoneID,
		StartTime:      store.Now(),
		Duration:       req.Duration,
		ParticipantIDs: []string{},
		Status:         "active",
		MaxParticipants: maxCoworkingParticipants,
		TimerState: &models.TimerState{
			Remaining: totalSeconds,
			IsPaused:  false,
			Phase:     phase,
		},
		CreatedAt: store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.CoworkingSessions = append(d.CoworkingSessions, session)
	d.Mu.Unlock()
	d.Save()

	pub := store.ToPublic(*user)
	session.Host = &pub

	BroadcastCoworkingEvent("session-created", session)
	JSON(w, 201, session)
}

func ListCoworkingSessions(w http.ResponseWriter, r *http.Request) {
	d := database.GetDB()
	d.Mu.RLock()
	var sessions []models.CoworkingSession
	for _, s := range d.CoworkingSessions {
		if s.Status == "active" {
			sessions = append(sessions, s)
		}
	}
	d.Mu.RUnlock()

	for i := range sessions {
		host := store.FindUserByID(sessions[i].HostID)
		if host != nil {
			pub := store.ToPublic(*host)
			sessions[i].Host = &pub
		}
	}

	if sessions == nil {
		sessions = []models.CoworkingSession{}
	}
	JSON(w, 200, sessions)
}

func JoinCoworkingSession(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := extractCoworkingID(r.URL.Path, "/join")

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.CoworkingSessions {
		if d.CoworkingSessions[i].ID == id {
			s := &d.CoworkingSessions[i]
			if s.Status != "active" {
				d.Mu.Unlock()
				Error(w, 400, "Session is not active")
				return
			}
			if s.HostID == user.ID {
				d.Mu.Unlock()
				Error(w, 400, "You are the host")
				return
			}
			for _, pid := range s.ParticipantIDs {
				if pid == user.ID {
					d.Mu.Unlock()
					Error(w, 400, "Already joined")
					return
				}
			}
			if s.MaxParticipants > 0 && len(s.ParticipantIDs)+1 >= s.MaxParticipants {
				d.Mu.Unlock()
				Error(w, 400, "Session is full")
				return
			}
			s.ParticipantIDs = append(s.ParticipantIDs, user.ID)
			d.Mu.Unlock()
			d.Save()

			BroadcastCoworkingEvent("participant-joined", *s)
			JSON(w, 200, H{"message": "Joined session"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Session not found")
}

func LeaveCoworkingSession(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := extractCoworkingID(r.URL.Path, "/leave")

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.CoworkingSessions {
		if d.CoworkingSessions[i].ID == id {
			s := &d.CoworkingSessions[i]
			var updated []string
			for _, pid := range s.ParticipantIDs {
				if pid != user.ID {
					updated = append(updated, pid)
				}
			}
			s.ParticipantIDs = updated
			d.Mu.Unlock()
			d.Save()

			BroadcastCoworkingEvent("participant-left", *s)
			JSON(w, 200, H{"message": "Left session"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Session not found")
}

func UpdateCoworkingTimer(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := extractCoworkingID(r.URL.Path, "/timer")

	var req struct {
		IsPaused *bool  `json:"isPaused"`
		Phase    string `json:"phase"`
	}
	if err := Decode(r, &req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.CoworkingSessions {
		if d.CoworkingSessions[i].ID == id {
			s := &d.CoworkingSessions[i]
			if s.HostID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Only the host can update the timer")
				return
			}
			if s.TimerState == nil {
				s.TimerState = &models.TimerState{}
			}
			if req.IsPaused != nil {
				s.TimerState.IsPaused = *req.IsPaused
				if !*req.IsPaused {
					s.TimerState.LastResumed = store.Now()
				}
			}
			if req.Phase != "" {
				switch req.Phase {
				case "work", "break":
					s.TimerState.Phase = req.Phase
					if req.Phase == "break" {
						s.TimerState.Remaining = pomodoroBreakMinutes * 60
					} else {
						s.TimerState.Remaining = pomodoroWorkMinutes * 60
					}
					s.TimerState.LastResumed = store.Now()
				}
			}
			d.Mu.Unlock()
			d.Save()

			BroadcastCoworkingEvent("timer-updated", *s)
			JSON(w, 200, s.TimerState)
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Session not found")
}

func EndCoworkingSession(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	id := extractCoworkingID(r.URL.Path, "/end")

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.CoworkingSessions {
		if d.CoworkingSessions[i].ID == id {
			s := &d.CoworkingSessions[i]
			if s.HostID != user.ID {
				d.Mu.Unlock()
				Error(w, 403, "Only the host can end the session")
				return
			}
			s.Status = "completed"
			participantIDs := make([]string, len(s.ParticipantIDs))
			copy(participantIDs, s.ParticipantIDs)
			d.Mu.Unlock()
			d.Save()

			BroadcastCoworkingEvent("session-ended", *s)
			for _, pid := range participantIDs {
				createCoworkingNotification(pid, s.Title, "ended")
			}
			JSON(w, 200, H{"message": "Session ended"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Session not found")
}

func BroadcastCoworkingEvent(eventType string, session models.CoworkingSession) {
	realtimeHub.Broadcast(RealtimeMessage{
		Type:    "coworking-" + eventType,
		Payload: session,
	}, "")
}

func createCoworkingNotification(userID, sessionTitle, action string) {
	d := database.GetDB()
	notif := models.Notification{
		ID:        store.NewID(),
		UserID:    userID,
		Type:      "coworking",
		Title:     "Coworking Session " + strings.ToUpper(action[:1]) + action[1:],
		Message:   "The session \"" + sessionTitle + "\" has been " + action + ".",
		Link:      "/world",
		IsRead:    false,
		CreatedAt: store.Now(),
	}
	d.Mu.Lock()
	d.Notifications = append(d.Notifications, notif)
	d.Mu.Unlock()
	d.Save()

	BroadcastNotification(userID, notif)
}

func ExpireStaleCoworkingSessions() {
	d := database.GetDB()
	d.Mu.Lock()
	now := time.Now().UTC()
	changed := false
	for i := range d.CoworkingSessions {
		s := &d.CoworkingSessions[i]
		if s.Status != "active" {
			continue
		}
		start, err := time.Parse(time.RFC3339, s.StartTime)
		if err != nil {
			continue
		}
		maxDuration := time.Duration(s.Duration)*time.Minute + 30*time.Minute
		if now.Sub(start) > maxDuration {
			s.Status = "completed"
			changed = true
			log.Printf("Auto-expired coworking session %s (started %s)", s.ID, s.StartTime)
			go func(participants []string, title string) {
				for _, pid := range participants {
					createCoworkingNotification(pid, title, "expired")
				}
			}(append([]string{}, s.ParticipantIDs...), s.Title)
		}
	}
	d.Mu.Unlock()
	if changed {
		d.Save()
	}
}

func SyncCoworkingTimers() {
	d := database.GetDB()
	d.Mu.Lock()
	now := time.Now().UTC()
	changed := false
	for i := range d.CoworkingSessions {
		s := &d.CoworkingSessions[i]
		if s.Status != "active" || s.TimerState == nil || s.TimerState.IsPaused {
			continue
		}
		if s.TimerState.LastResumed == "" {
			continue
		}
		resumed, err := time.Parse(time.RFC3339, s.TimerState.LastResumed)
		if err != nil {
			continue
		}
		elapsed := int(now.Sub(resumed).Seconds())
		newRemaining := s.TimerState.Remaining
		if elapsed > 0 {
			newRemaining = s.TimerState.Remaining - elapsed
			if newRemaining < 0 {
				newRemaining = 0
			}
			if newRemaining != s.TimerState.Remaining {
				s.TimerState.Remaining = newRemaining
				changed = true
			}
		}
		if newRemaining <= 0 && s.Type == "pomodoro" && s.TimerState.Phase == "work" {
			s.TimerState.Phase = "break"
			s.TimerState.Remaining = pomodoroBreakMinutes * 60
			s.TimerState.LastResumed = store.Now()
			changed = true
		} else if newRemaining <= 0 && s.Type == "pomodoro" && s.TimerState.Phase == "break" {
			s.TimerState.Phase = "work"
			s.TimerState.Remaining = pomodoroWorkMinutes * 60
			s.TimerState.LastResumed = store.Now()
			changed = true
		} else if newRemaining <= 0 && s.Type != "pomodoro" {
			s.Status = "completed"
			changed = true
		}
	}
	d.Mu.Unlock()
	if changed {
		d.Save()
	}
}

func GetCoworkingSession(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/coworking/")
	id = strings.TrimSuffix(id, "/")

	d := database.GetDB()
	d.Mu.RLock()
	for _, s := range d.CoworkingSessions {
		if s.ID == id {
			d.Mu.RUnlock()
			host := store.FindUserByID(s.HostID)
			if host != nil {
				pub := store.ToPublic(*host)
				s.Host = &pub
			}
			JSON(w, 200, s)
			return
		}
	}
	d.Mu.RUnlock()
	Error(w, 404, "Session not found")
}

func extractCoworkingID(path, suffix string) string {
	trimmed := strings.TrimPrefix(path, "/api/v1/coworking/")
	trimmed = strings.TrimSuffix(trimmed, suffix)
	trimmed = strings.TrimSuffix(trimmed, "/")
	return trimmed
}

func init() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			SyncCoworkingTimers()
			ExpireStaleCoworkingSessions()
		}
	}()
}
