package handlers

import (
	crand "crypto/rand"
	"encoding/json"
	"math"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

func SendEmailVerificationHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, v := range d.Verifications {
		if v.UserID == user.ID && v.Type == "email" && v.Status == "verified" {
			d.Mu.RUnlock()
			JSON(w, 200, H{"message": "Email already verified"})
			return
		}
	}
	d.Mu.RUnlock()

	code := generateVerificationCode()

	verification := models.Verification{
		ID:        store.NewID(),
		UserID:    user.ID,
		Type:      "email",
		Status:    "pending",
		Code:      code,
		CreatedAt: store.Now(),
	}

	d.Mu.Lock()
	d.Verifications = append(d.Verifications, verification)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Verification code sent to " + user.Email})
}

func ConfirmEmailVerificationHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		Error(w, 400, "Code is required")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Verifications {
		if d.Verifications[i].UserID == user.ID && d.Verifications[i].Type == "email" && d.Verifications[i].Status == "pending" {
			if d.Verifications[i].Code == req.Code {
				d.Verifications[i].Status = "verified"
				d.Verifications[i].VerifiedAt = store.Now()
				d.Mu.Unlock()
				d.Save()
				JSON(w, 200, H{"message": "Email verified successfully"})
				return
			}
			d.Mu.Unlock()
			Error(w, 400, "Invalid code")
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "No pending verification found")
}

func SendPhoneVerificationHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" {
		Error(w, 400, "Phone number is required")
		return
	}

	code := generateVerificationCode()

	verification := models.Verification{
		ID:        store.NewID(),
		UserID:    user.ID,
		Type:      "phone",
		Status:    "pending",
		Code:      code,
		CreatedAt: store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.Verifications = append(d.Verifications, verification)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Verification code sent to " + req.Phone})
}

func ConfirmPhoneVerificationHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		Error(w, 400, "Code is required")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Verifications {
		if d.Verifications[i].UserID == user.ID && d.Verifications[i].Type == "phone" && d.Verifications[i].Status == "pending" {
			if d.Verifications[i].Code == req.Code {
				d.Verifications[i].Status = "verified"
				d.Verifications[i].VerifiedAt = store.Now()
				d.Mu.Unlock()
				d.Save()
				JSON(w, 200, H{"message": "Phone verified successfully"})
				return
			}
			d.Mu.Unlock()
			Error(w, 400, "Invalid code")
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "No pending verification found")
}

func SubmitIdentityVerificationHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		DocURL string `json:"docUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DocURL == "" {
		Error(w, 400, "Document URL is required")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, v := range d.Verifications {
		if v.UserID == user.ID && v.Type == "identity" && (v.Status == "pending" || v.Status == "verified") {
			d.Mu.RUnlock()
			Error(w, 409, "Identity verification already submitted")
			return
		}
	}
	d.Mu.RUnlock()

	verification := models.Verification{
		ID:        store.NewID(),
		UserID:    user.ID,
		Type:      "identity",
		Status:    "pending",
		DocURL:    req.DocURL,
		CreatedAt: store.Now(),
	}

	d.Mu.Lock()
	d.Verifications = append(d.Verifications, verification)
	d.Mu.Unlock()
	d.Save()

	JSON(w, 201, H{"message": "Identity verification submitted for review"})
}

func GetVerificationStatusHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()

	var emailVerified, phoneVerified, identityVerified bool
	var identityPending bool

	for _, v := range d.Verifications {
		if v.UserID == user.ID {
			switch v.Type {
			case "email":
				emailVerified = v.Status == "verified"
			case "phone":
				phoneVerified = v.Status == "verified"
			case "identity":
				identityVerified = v.Status == "verified"
				identityPending = v.Status == "pending"
			}
		}
	}
	d.Mu.RUnlock()

	JSON(w, 200, H{
		"emailVerified":    emailVerified,
		"phoneVerified":    phoneVerified,
		"identityVerified": identityVerified,
		"identityPending":  identityPending,
	})
}

func GetTrustScoreHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	userID = strings.TrimSuffix(userID, "/trust-score")

	d := database.GetDB()
	d.Mu.RLock()

	var emailVerified, phoneVerified, identityVerified bool
	for _, v := range d.Verifications {
		if v.UserID == userID && v.Status == "verified" {
			switch v.Type {
			case "email":
				emailVerified = true
			case "phone":
				phoneVerified = true
			case "identity":
				identityVerified = true
			}
		}
	}

	var totalOrders, completed int
	var totalRating float64
	var reviewCount int
	for _, o := range d.Orders {
		if o.SellerID == userID {
			totalOrders++
			if o.Status == "completed" {
				completed++
			}
		}
	}
	for _, rv := range d.Reviews {
		if rv.RevieweeID == userID {
			totalRating += float64(rv.Rating)
			reviewCount++
		}
	}

	var user *models.User
	for i := range d.Users {
		if d.Users[i].ID == userID {
			user = &d.Users[i]
			break
		}
	}
	d.Mu.RUnlock()

	score := 20.0
	if emailVerified {
		score += 15
	}
	if phoneVerified {
		score += 10
	}
	if identityVerified {
		score += 20
	}
	if totalOrders > 0 {
		completionRate := float64(completed) / float64(totalOrders)
		score += completionRate * 15
	}
	if reviewCount > 0 {
		avgRating := totalRating / float64(reviewCount)
		score += (avgRating / 5.0) * 20
	}
	score = math.Min(100, math.Round(score*10)/10)

	level := "new"
	if score >= 80 {
		level = "elite"
	} else if score >= 60 {
		level = "verified"
	} else if score >= 40 {
		level = "trusted"
	}

	avgRating := 0.0
	if reviewCount > 0 {
		avgRating = math.Round(totalRating/float64(reviewCount)*10) / 10
	}

	completionRate := 0.0
	if totalOrders > 0 {
		completionRate = math.Round(float64(completed)/float64(totalOrders)*100) / 100
	}

	var responseTime float64
	if user != nil {
		responseTime = 2.0
	}

	JSON(w, 200, models.TrustScore{
		UserID:           userID,
		Score:            score,
		EmailVerified:    emailVerified,
		PhoneVerified:    phoneVerified,
		IdentityVerified: identityVerified,
		SkillsVerified:   0,
		CompletionRate:   completionRate,
		AvgRating:        avgRating,
		ResponseTime:     responseTime,
		Level:            level,
	})
}

func AdminReviewIdentityHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	verificationID := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/verify/")
	verificationID = strings.TrimSuffix(verificationID, "/review")

	var req struct {
		Approved bool   `json:"approved"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Verifications {
		if d.Verifications[i].ID == verificationID && d.Verifications[i].Type == "identity" {
			if req.Approved {
				d.Verifications[i].Status = "verified"
				d.Verifications[i].VerifiedAt = store.Now()
			} else {
				d.Verifications[i].Status = "rejected"
			}
			d.Verifications[i].ReviewedBy = user.ID
			d.Mu.Unlock()
			d.Save()
			JSON(w, 200, H{"message": "Identity verification reviewed"})
			return
		}
	}
	d.Mu.Unlock()
	Error(w, 404, "Verification not found")
}

func AdminListPendingVerificationsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var pending []models.Verification
	for _, v := range d.Verifications {
		if v.Status == "pending" && v.Type == "identity" {
			pending = append(pending, v)
		}
	}
	d.Mu.RUnlock()

	if pending == nil {
		pending = []models.Verification{}
	}
	JSON(w, 200, pending)
}

func generateVerificationCode() string {
	b := make([]byte, 6)
	crand.Read(b)
	code := ""
	for i := 0; i < 6; i++ {
		code += string(rune('0' + int(b[i]%10)))
	}
	return code
}
