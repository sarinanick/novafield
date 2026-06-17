package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func GetReferralLinkHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var code string
	for _, ref := range d.Referrals {
		if ref.ReferrerID == user.ID {
			code = ref.Code
			break
		}
	}
	d.Mu.RUnlock()

	if code == "" {
		code = "REF-" + user.ID[:8]
	}

	JSON(w, 200, H{
		"code": code,
		"link": "https://novafield.ai/register?ref=" + code,
	})
}

func ApplyReferralCodeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		Error(w, 400, "Referral code is required")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, ref := range d.Referrals {
		if ref.ReferredID == user.ID {
			d.Mu.RUnlock()
			Error(w, 409, "Already used a referral code")
			return
		}
	}
	d.Mu.RUnlock()

	code := strings.TrimPrefix(req.Code, "REF-")
	var referrerID string
	d.Mu.RLock()
	for _, u := range d.Users {
		if u.ID[:8] == code {
			referrerID = u.ID
			break
		}
	}
	d.Mu.RUnlock()

	if referrerID == "" || referrerID == user.ID {
		Error(w, 400, "Invalid referral code")
		return
	}

	referral := models.Referral{
		ID:         store.NewID(),
		ReferrerID: referrerID,
		ReferredID: user.ID,
		Code:       req.Code,
		Status:     "pending",
		CreatedAt:  store.Now(),
	}

	d.Mu.Lock()
	d.Referrals = append(d.Referrals, referral)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Referral code applied"})
}

func GetReferralStatsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	var totalReferrals, completedReferrals int
	var totalEarnings float64

	for _, ref := range d.Referrals {
		if ref.ReferrerID == user.ID {
			totalReferrals++
			if ref.Status == "completed" {
				completedReferrals++
			}
		}
	}

	for _, earning := range d.ReferralEarnings {
		if earning.ReferrerID == user.ID {
			totalEarnings += earning.Amount
		}
	}
	d.Mu.RUnlock()

	JSON(w, 200, H{
		"totalReferrals":     totalReferrals,
		"completedReferrals": completedReferrals,
		"totalEarnings":      totalEarnings,
	})
}

func GetReferralEarningsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var earnings []models.ReferralEarning
	for _, e := range d.ReferralEarnings {
		if e.ReferrerID == user.ID {
			earnings = append(earnings, e)
		}
	}
	d.Mu.RUnlock()

	if earnings == nil {
		earnings = []models.ReferralEarning{}
	}
	JSON(w, 200, earnings)
}
