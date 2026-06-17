package handlers

import (
	"encoding/json"
	"novafield-api/database"
	"testing"
)

func TestSendEmailVerification(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("POST", "/api/v1/verify/email", nil, token)
	rr := newRecorder()

	SendEmailVerificationHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["code"] == nil {
		t.Fatal("expected verification code")
	}
}

func TestSendEmailVerification_Unauthorized(t *testing.T) {
	resetDB()

	req := authRequest("POST", "/api/v1/verify/email", nil, "")
	rr := newRecorder()

	SendEmailVerificationHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestConfirmEmailVerification(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("POST", "/api/v1/verify/email", nil, token)
	rr := newRecorder()
	SendEmailVerificationHandler(rr, req)

	result := decodeJSON(rr)
	code := result["code"].(string)

	body := jsonBody(map[string]interface{}{"code": code})
	confirmReq := authRequest("POST", "/api/v1/verify/email/confirm", body, token)
	confirmRR := newRecorder()
	ConfirmEmailVerificationHandler(confirmRR, confirmReq)

	if confirmRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", confirmRR.Code, confirmRR.Body.String())
	}
}

func TestConfirmEmailVerification_WrongCode(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("POST", "/api/v1/verify/email", nil, token)
	rr := newRecorder()
	SendEmailVerificationHandler(rr, req)

	body := jsonBody(map[string]interface{}{"code": "000000"})
	confirmReq := authRequest("POST", "/api/v1/verify/email/confirm", body, token)
	confirmRR := newRecorder()
	ConfirmEmailVerificationHandler(confirmRR, confirmReq)

	if confirmRR.Code != 400 {
		t.Fatalf("expected 400, got %d", confirmRR.Code)
	}
}

func TestSendPhoneVerification(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"phone": "+1234567890"})
	req := authRequest("POST", "/api/v1/verify/phone", body, token)
	rr := newRecorder()

	SendPhoneVerificationHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestConfirmPhoneVerification(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	phoneBody := jsonBody(map[string]interface{}{"phone": "+1234567890"})
	req := authRequest("POST", "/api/v1/verify/phone", phoneBody, token)
	rr := newRecorder()
	SendPhoneVerificationHandler(rr, req)

	result := decodeJSON(rr)
	code := result["code"].(string)

	body := jsonBody(map[string]interface{}{"code": code})
	confirmReq := authRequest("POST", "/api/v1/verify/phone/confirm", body, token)
	confirmRR := newRecorder()
	ConfirmPhoneVerificationHandler(confirmRR, confirmReq)

	if confirmRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", confirmRR.Code, confirmRR.Body.String())
	}
}

func TestSubmitIdentityVerification(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"docUrl": "/uploads/id-card.jpg"})
	req := authRequest("POST", "/api/v1/verify/identity", body, token)
	rr := newRecorder()

	SubmitIdentityVerificationHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSubmitIdentityVerification_Duplicate(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"docUrl": "/uploads/id-card.jpg"})
	req := authRequest("POST", "/api/v1/verify/identity", body, token)
	rr := newRecorder()
	SubmitIdentityVerificationHandler(rr, req)

	body2 := jsonBody(map[string]interface{}{"docUrl": "/uploads/id-card2.jpg"})
	req2 := authRequest("POST", "/api/v1/verify/identity", body2, token)
	rr2 := newRecorder()
	SubmitIdentityVerificationHandler(rr2, req2)

	if rr2.Code != 409 {
		t.Fatalf("expected 409, got %d", rr2.Code)
	}
}

func TestGetVerificationStatus(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	emailReq := authRequest("POST", "/api/v1/verify/email", nil, token)
	emailRR := newRecorder()
	SendEmailVerificationHandler(emailRR, emailReq)

	req := authRequest("GET", "/api/v1/verify/status", nil, token)
	rr := newRecorder()

	GetVerificationStatusHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["emailVerified"] != false {
		t.Error("email should not be verified yet")
	}
}

func TestGetTrustScore(t *testing.T) {
	resetDB()
	user, _ := createTestUser("freelancer")

	req := authRequest("GET", "/api/v1/users/"+user.ID+"/trust-score", nil, "")
	rr := newRecorder()

	GetTrustScoreHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["userId"] != user.ID {
		t.Errorf("expected userId %s, got %v", user.ID, result["userId"])
	}
	if result["level"] != "new" {
		t.Errorf("expected level 'new', got %v", result["level"])
	}
	if result["emailVerified"] != false {
		t.Error("email should not be verified")
	}
}

func TestGetTrustScore_WithEmailVerified(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")

	emailReq := authRequest("POST", "/api/v1/verify/email", nil, token)
	emailRR := newRecorder()
	SendEmailVerificationHandler(emailRR, emailReq)

	result := decodeJSON(emailRR)
	code := result["code"].(string)

	body := jsonBody(map[string]interface{}{"code": code})
	confirmReq := authRequest("POST", "/api/v1/verify/email/confirm", body, token)
	confirmRR := newRecorder()
	ConfirmEmailVerificationHandler(confirmRR, confirmReq)

	trustReq := authRequest("GET", "/api/v1/users/"+user.ID+"/trust-score", nil, "")
	trustRR := newRecorder()
	GetTrustScoreHandler(trustRR, trustReq)

	trustResult := decodeJSON(trustRR)
	if trustResult["emailVerified"] != true {
		t.Error("email should be verified")
	}
	if trustResult["score"].(float64) <= 20 {
		t.Error("score should be higher than base with email verified")
	}
}

func TestAdminReviewIdentity(t *testing.T) {
	resetDB()
	user, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"docUrl": "/uploads/id.jpg"})
	req := authRequest("POST", "/api/v1/verify/identity", body, token)
	rr := newRecorder()
	SubmitIdentityVerificationHandler(rr, req)

	d := database.GetDB()
	d.Mu.RLock()
	var verificationID string
	for _, v := range d.Verifications {
		if v.UserID == user.ID && v.Type == "identity" {
			verificationID = v.ID
			break
		}
	}
	d.Mu.RUnlock()

	_, adminToken := createTestUser("admin")

	reviewBody := jsonBody(map[string]interface{}{"approved": true, "note": "Looks good"})
	reviewReq := authRequest("POST", "/api/v1/admin/verify/"+verificationID+"/review", reviewBody, adminToken)
	reviewRR := newRecorder()
	AdminReviewIdentityHandler(reviewRR, reviewReq)

	if reviewRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", reviewRR.Code, reviewRR.Body.String())
	}
}

func TestAdminReviewIdentity_NonAdmin(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	reviewBody := jsonBody(map[string]interface{}{"approved": true})
	reviewReq := authRequest("POST", "/api/v1/admin/verify/fake/review", reviewBody, token)
	reviewRR := newRecorder()
	AdminReviewIdentityHandler(reviewRR, reviewReq)

	if reviewRR.Code != 403 {
		t.Fatalf("expected 403, got %d", reviewRR.Code)
	}
}

func TestAdminListPendingVerifications(t *testing.T) {
	resetDB()
	user, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"docUrl": "/uploads/id.jpg"})
	req := authRequest("POST", "/api/v1/verify/identity", body, token)
	rr := newRecorder()
	SubmitIdentityVerificationHandler(rr, req)

	_, adminToken := createTestUser("admin")

	listReq := authRequest("GET", "/api/v1/admin/verify/", nil, adminToken)
	listRR := newRecorder()
	AdminListPendingVerificationsHandler(listRR, listReq)

	if listRR.Code != 200 {
		t.Fatalf("expected 200, got %d", listRR.Code)
	}

	var pending []map[string]interface{}
	json.NewDecoder(listRR.Body).Decode(&pending)
	if len(pending) != 1 {
		t.Errorf("expected 1 pending, got %d", len(pending))
	}

	_ = user
}

func TestVerification_FullLifecycle(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")

	emailReq := authRequest("POST", "/api/v1/verify/email", nil, token)
	emailRR := newRecorder()
	SendEmailVerificationHandler(emailRR, emailReq)

	if emailRR.Code != 200 {
		t.Fatalf("send email: expected 200, got %d", emailRR.Code)
	}
	emailCode := decodeJSON(emailRR)["code"].(string)

	confirmBody := jsonBody(map[string]interface{}{"code": emailCode})
	confirmReq := authRequest("POST", "/api/v1/verify/email/confirm", confirmBody, token)
	confirmRR := newRecorder()
	ConfirmEmailVerificationHandler(confirmRR, confirmReq)

	if confirmRR.Code != 200 {
		t.Fatalf("confirm email: expected 200, got %d", confirmRR.Code)
	}

	phoneBody := jsonBody(map[string]interface{}{"phone": "+1234567890"})
	phoneReq := authRequest("POST", "/api/v1/verify/phone", phoneBody, token)
	phoneRR := newRecorder()
	SendPhoneVerificationHandler(phoneRR, phoneReq)

	phoneCode := decodeJSON(phoneRR)["code"].(string)
	phoneConfirmBody := jsonBody(map[string]interface{}{"code": phoneCode})
	phoneConfirmReq := authRequest("POST", "/api/v1/verify/phone/confirm", phoneConfirmBody, token)
	phoneConfirmRR := newRecorder()
	ConfirmPhoneVerificationHandler(phoneConfirmRR, phoneConfirmReq)

	if phoneConfirmRR.Code != 200 {
		t.Fatalf("confirm phone: expected 200, got %d", phoneConfirmRR.Code)
	}

	identityBody := jsonBody(map[string]interface{}{"docUrl": "/uploads/passport.jpg"})
	identityReq := authRequest("POST", "/api/v1/verify/identity", identityBody, token)
	identityRR := newRecorder()
	SubmitIdentityVerificationHandler(identityRR, identityReq)

	if identityRR.Code != 201 {
		t.Fatalf("submit identity: expected 201, got %d", identityRR.Code)
	}

	statusReq := authRequest("GET", "/api/v1/verify/status", nil, token)
	statusRR := newRecorder()
	GetVerificationStatusHandler(statusRR, statusReq)

	status := decodeJSON(statusRR)
	if status["emailVerified"] != true {
		t.Error("email should be verified")
	}
	if status["phoneVerified"] != true {
		t.Error("phone should be verified")
	}
	if status["identityVerified"] != false {
		t.Error("identity should not be verified yet")
	}
	if status["identityPending"] != true {
		t.Error("identity should be pending")
	}

	trustReq := authRequest("GET", "/api/v1/users/"+user.ID+"/trust-score", nil, "")
	trustRR := newRecorder()
	GetTrustScoreHandler(trustRR, trustReq)

	trust := decodeJSON(trustRR)
	if trust["emailVerified"] != true {
		t.Error("trust score should show email verified")
	}
	if trust["phoneVerified"] != true {
		t.Error("trust score should show phone verified")
	}
	if trust["score"].(float64) <= 20 {
		t.Error("trust score should be higher than base")
	}
}
