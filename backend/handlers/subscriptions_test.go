package handlers

import (
	"encoding/json"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"testing"
)

func seedSubscriptionPlan(freelancerID string) (string, string) {
	gigID, _ := seedTestGig(freelancerID)
	planID := store.NewID()

	d := database.GetDB()
	d.Mu.Lock()
	d.SubscriptionPlans = append(d.SubscriptionPlans, models.SubscriptionPlan{
		ID:           planID,
		GigID:        gigID,
		Name:         "Monthly Content",
		Interval:     "monthly",
		Price:        200,
		Deliverables: "4 blog posts per month",
		MaxRevisions: 2,
		IsActive:     true,
		CreatedAt:    store.Now(),
	})
	d.Mu.Unlock()

	return gigID, planID
}

func TestCreateSubscriptionPlan(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	gigID, _ := seedTestGig(user.ID)

	body := jsonBody(map[string]interface{}{
		"name":         "Monthly SEO Package",
		"interval":     "monthly",
		"price":        300,
		"deliverables": "8 articles + keyword research",
		"maxRevisions": 3,
	})
	req := authRequest("POST", "/api/v1/gigs/"+gigID+"/subscriptions", body, token)
	rr := newRecorder()

	CreateSubscriptionPlanHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["id"] == nil {
		t.Fatal("expected plan id")
	}
}

func TestCreateSubscriptionPlan_NonOwner(t *testing.T) {
	resetDB()
	owner, _ := createTestUser("freelancer")
	_, token := createTestUser("freelancer")
	gigID, _ := seedTestGig(owner.ID)

	body := jsonBody(map[string]interface{}{
		"name":  "Should Fail",
		"price": 100,
	})
	req := authRequest("POST", "/api/v1/gigs/"+gigID+"/subscriptions", body, token)
	rr := newRecorder()

	CreateSubscriptionPlanHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestCreateSubscriptionPlan_Client_Forbidden(t *testing.T) {
	resetDB()
	user, _ := createTestUser("freelancer")
	_, token := createTestUser("client")
	gigID, _ := seedTestGig(user.ID)

	body := jsonBody(map[string]interface{}{
		"name":  "Should Fail",
		"price": 100,
	})
	req := authRequest("POST", "/api/v1/gigs/"+gigID+"/subscriptions", body, token)
	rr := newRecorder()

	CreateSubscriptionPlanHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestSubscribe_Success(t *testing.T) {
	resetDB()
	freelancer, _ := createTestUser("freelancer")
	_, planID := seedSubscriptionPlan(freelancer.ID)
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"planId": planID,
	})
	req := authRequest("POST", "/api/v1/subscriptions", body, token)
	rr := newRecorder()

	SubscribeHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSubscribe_OwnPlan(t *testing.T) {
	resetDB()
	freelancer, token := createTestUser("freelancer")
	_, planID := seedSubscriptionPlan(freelancer.ID)

	body := jsonBody(map[string]interface{}{
		"planId": planID,
	})
	req := authRequest("POST", "/api/v1/subscriptions", body, token)
	rr := newRecorder()

	SubscribeHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSubscribe_Duplicate(t *testing.T) {
	resetDB()
	freelancer, _ := createTestUser("freelancer")
	_, planID := seedSubscriptionPlan(freelancer.ID)
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"planId": planID,
	})
	req := authRequest("POST", "/api/v1/subscriptions", body, token)
	rr := newRecorder()
	SubscribeHandler(rr, req)

	body2 := jsonBody(map[string]interface{}{
		"planId": planID,
	})
	req2 := authRequest("POST", "/api/v1/subscriptions", body2, token)
	rr2 := newRecorder()
	SubscribeHandler(rr2, req2)

	if rr2.Code != 409 {
		t.Fatalf("expected 409, got %d: %s", rr2.Code, rr2.Body.String())
	}
}

func TestListSubscriptions(t *testing.T) {
	resetDB()
	freelancer, _ := createTestUser("freelancer")
	_, planID := seedSubscriptionPlan(freelancer.ID)
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"planId": planID})
	req := authRequest("POST", "/api/v1/subscriptions", body, token)
	rr := newRecorder()
	SubscribeHandler(rr, req)

	listReq := authRequest("GET", "/api/v1/subscriptions", nil, token)
	listRR := newRecorder()
	ListSubscriptionsHandler(listRR, listReq)

	if listRR.Code != 200 {
		t.Fatalf("expected 200, got %d", listRR.Code)
	}

	var subs []map[string]interface{}
	json.NewDecoder(listRR.Body).Decode(&subs)
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
}

func TestGetSubscription(t *testing.T) {
	resetDB()
	freelancer, _ := createTestUser("freelancer")
	_, planID := seedSubscriptionPlan(freelancer.ID)
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"planId": planID})
	req := authRequest("POST", "/api/v1/subscriptions", body, token)
	rr := newRecorder()
	SubscribeHandler(rr, req)

	result := decodeJSON(rr)
	subID := result["id"].(string)

	getReq := authRequest("GET", "/api/v1/subscriptions/"+subID, nil, token)
	getRR := newRecorder()
	GetSubscriptionHandler(getRR, getReq)

	if getRR.Code != 200 {
		t.Fatalf("expected 200, got %d", getRR.Code)
	}
}

func TestPauseSubscription(t *testing.T) {
	resetDB()
	freelancer, _ := createTestUser("freelancer")
	_, planID := seedSubscriptionPlan(freelancer.ID)
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"planId": planID})
	req := authRequest("POST", "/api/v1/subscriptions", body, token)
	rr := newRecorder()
	SubscribeHandler(rr, req)

	result := decodeJSON(rr)
	subID := result["id"].(string)

	pauseReq := authRequest("POST", "/api/v1/subscriptions/"+subID+"/pause", nil, token)
	pauseRR := newRecorder()
	PauseSubscriptionHandler(pauseRR, pauseReq)

	if pauseRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", pauseRR.Code, pauseRR.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, s := range d.Subscriptions {
		if s.ID == subID {
			if s.Status != "paused" {
				t.Errorf("expected paused, got %s", s.Status)
			}
			break
		}
	}
	d.Mu.RUnlock()
}

func TestCancelSubscription(t *testing.T) {
	resetDB()
	freelancer, _ := createTestUser("freelancer")
	_, planID := seedSubscriptionPlan(freelancer.ID)
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"planId": planID})
	req := authRequest("POST", "/api/v1/subscriptions", body, token)
	rr := newRecorder()
	SubscribeHandler(rr, req)

	result := decodeJSON(rr)
	subID := result["id"].(string)

	cancelReq := authRequest("POST", "/api/v1/subscriptions/"+subID+"/cancel", nil, token)
	cancelRR := newRecorder()
	CancelSubscriptionHandler(cancelRR, cancelReq)

	if cancelRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", cancelRR.Code, cancelRR.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	for _, s := range d.Subscriptions {
		if s.ID == subID {
			if s.Status != "cancelled" {
				t.Errorf("expected cancelled, got %s", s.Status)
			}
			if s.CancelledAt == "" {
				t.Error("expected cancelledAt to be set")
			}
			break
		}
	}
	d.Mu.RUnlock()
}

func TestChangePlan(t *testing.T) {
	resetDB()
	freelancer, _ := createTestUser("freelancer")
	gigID, planID := seedSubscriptionPlan(freelancer.ID)

	newPlanID := store.NewID()
	d := database.GetDB()
	d.Mu.Lock()
	d.SubscriptionPlans = append(d.SubscriptionPlans, models.SubscriptionPlan{
		ID: newPlanID, GigID: gigID, Name: "Quarterly", Interval: "quarterly",
		Price: 500, IsActive: true, CreatedAt: store.Now(),
	})
	d.Mu.Unlock()

	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{"planId": planID})
	req := authRequest("POST", "/api/v1/subscriptions", body, token)
	rr := newRecorder()
	SubscribeHandler(rr, req)

	result := decodeJSON(rr)
	subID := result["id"].(string)

	changeBody := jsonBody(map[string]interface{}{"newPlanId": newPlanID})
	changeReq := authRequest("PUT", "/api/v1/subscriptions/"+subID+"/plan", changeBody, token)
	changeRR := newRecorder()
	ChangePlanHandler(changeRR, changeReq)

	if changeRR.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", changeRR.Code, changeRR.Body.String())
	}

	d.Mu.RLock()
	for _, s := range d.Subscriptions {
		if s.ID == subID {
			if s.PlanID != newPlanID {
				t.Errorf("expected plan %s, got %s", newPlanID, s.PlanID)
			}
			break
		}
	}
	d.Mu.RUnlock()
}

func TestSubscription_FullLifecycle(t *testing.T) {
	resetDB()
	freelancer, _ := createTestUser("freelancer")
	_, planID := seedSubscriptionPlan(freelancer.ID)
	_, token := createTestUser("client")

	subBody := jsonBody(map[string]interface{}{"planId": planID})
	subReq := authRequest("POST", "/api/v1/subscriptions", subBody, token)
	subRR := newRecorder()
	SubscribeHandler(subRR, subReq)

	if subRR.Code != 201 {
		t.Fatalf("subscribe: expected 201, got %d", subRR.Code)
	}

	result := decodeJSON(subRR)
	subID := result["id"].(string)

	getReq := authRequest("GET", "/api/v1/subscriptions/"+subID, nil, token)
	getRR := newRecorder()
	GetSubscriptionHandler(getRR, getReq)

	if getRR.Code != 200 {
		t.Fatalf("get: expected 200, got %d", getRR.Code)
	}

	pauseReq := authRequest("POST", "/api/v1/subscriptions/"+subID+"/pause", nil, token)
	pauseRR := newRecorder()
	PauseSubscriptionHandler(pauseRR, pauseReq)

	if pauseRR.Code != 200 {
		t.Fatalf("pause: expected 200, got %d", pauseRR.Code)
	}
}
