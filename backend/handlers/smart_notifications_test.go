package handlers

import (
	"testing"
)

func TestGetNotificationPreferences(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("GET", "/api/v1/notifications/preferences", nil, token)
	rr := newRecorder()

	GetNotificationPreferencesHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["emailOrders"] != true {
		t.Error("expected emailOrders to be true by default")
	}
	if result["digestFrequency"] != "daily" {
		t.Errorf("expected daily digest, got %v", result["digestFrequency"])
	}
}

func TestGetNotificationPreferences_Unauthorized(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/notifications/preferences", nil, "")
	rr := newRecorder()

	GetNotificationPreferencesHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestUpdateNotificationPreferences(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"emailOrders":     false,
		"emailMessages":   true,
		"emailMarketing":  true,
		"inAppOrders":     true,
		"inAppMessages":   true,
		"quietHoursStart": "23:00",
		"quietHoursEnd":   "07:00",
		"digestFrequency": "weekly",
	})
	req := authRequest("POST", "/api/v1/notifications/preferences", body, token)
	rr := newRecorder()

	UpdateNotificationPreferencesHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	getReq := authRequest("GET", "/api/v1/notifications/preferences", nil, token)
	getRR := newRecorder()
	GetNotificationPreferencesHandler(getRR, getReq)

	result := decodeJSON(getRR)
	if result["emailOrders"] != false {
		t.Error("expected emailOrders to be false")
	}
	if result["digestFrequency"] != "weekly" {
		t.Errorf("expected weekly, got %v", result["digestFrequency"])
	}
}

func TestGetNotificationDigest(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	req := authRequest("GET", "/api/v1/notifications/digest", nil, token)
	rr := newRecorder()

	GetNotificationDigestHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["summary"] == nil {
		t.Fatal("expected summary")
	}
	if result["unreadMessages"].(float64) != 0 {
		t.Error("expected 0 unread messages")
	}
}

func TestGetNotificationDigest_Unauthorized(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/notifications/digest", nil, "")
	rr := newRecorder()

	GetNotificationDigestHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
