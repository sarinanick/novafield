package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
)

func GetNotificationPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, p := range d.NotificationPrefs {
		if p.UserID == user.ID {
			d.Mu.RUnlock()
			JSON(w, 200, p)
			return
		}
	}
	d.Mu.RUnlock()

	defaults := models.NotificationPreference{
		UserID:          user.ID,
		EmailOrders:     true,
		EmailMessages:   true,
		EmailMarketing:  false,
		InAppOrders:     true,
		InAppMessages:   true,
		QuietHoursStart: "22:00",
		QuietHoursEnd:   "08:00",
		DigestFrequency: "daily",
	}
	JSON(w, 200, defaults)
}

func UpdateNotificationPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req models.NotificationPreference
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	req.UserID = user.ID

	d := database.GetDB()
	d.Mu.Lock()
	found := false
	for i := range d.NotificationPrefs {
		if d.NotificationPrefs[i].UserID == user.ID {
			d.NotificationPrefs[i] = req
			found = true
			break
		}
	}
	if !found {
		d.NotificationPrefs = append(d.NotificationPrefs, req)
	}
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Preferences updated"})
}

func GetNotificationDigestHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	var unreadOrders, unreadMessages, unreadNotifications int
	for _, o := range d.Orders {
		if o.BuyerID == user.ID && o.Status == "delivered" {
			unreadOrders++
		}
	}
	for _, m := range d.Messages {
		if m.ReceiverID == user.ID && !m.IsRead {
			unreadMessages++
		}
	}
	for _, n := range d.Notifications {
		if n.UserID == user.ID && !n.IsRead {
			unreadNotifications++
		}
	}

	var recentOrders []map[string]interface{}
	for _, o := range d.Orders {
		if (o.BuyerID == user.ID || o.SellerID == user.ID) && o.Status == "active" {
			recentOrders = append(recentOrders, map[string]interface{}{
				"id":     o.ID,
				"status": o.Status,
				"price":  o.Price,
			})
		}
	}
	d.Mu.RUnlock()

	digest := map[string]interface{}{
		"unreadOrders":       unreadOrders,
		"unreadMessages":     unreadMessages,
		"unreadNotifications": unreadNotifications,
		"recentActiveOrders": recentOrders,
		"summary":            fmt.Sprintf("You have %d unread messages, %d pending orders, and %d notifications", unreadMessages, unreadOrders, unreadNotifications),
	}

	JSON(w, 200, digest)
}
