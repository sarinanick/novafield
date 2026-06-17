package handlers

import (
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"testing"
)

func TestCreateOrganization(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"name":        "AI Creative Studio",
		"description": "Professional AI services agency",
	})
	req := authRequest("POST", "/api/v1/organizations", body, token)
	rr := newRecorder()

	CreateOrganizationHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["id"] == nil {
		t.Fatal("expected org id")
	}
}

func TestCreateOrganization_DuplicateSlug(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{"name": "Same Name"})
	req := authRequest("POST", "/api/v1/organizations", body, token)
	rr := newRecorder()
	CreateOrganizationHandler(rr, req)

	body2 := jsonBody(map[string]interface{}{"name": "Same Name"})
	req2 := authRequest("POST", "/api/v1/organizations", body2, token)
	rr2 := newRecorder()
	CreateOrganizationHandler(rr2, req2)

	if rr2.Code != 409 {
		t.Fatalf("expected 409, got %d", rr2.Code)
	}
}

func TestCreateOrganization_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"name": "test"})
	req := authRequest("POST", "/api/v1/organizations", body, "")
	rr := newRecorder()

	CreateOrganizationHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestGetOrganization(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")

	orgID := createTestOrg(user.ID)

	req := authRequest("GET", "/api/v1/organizations/"+orgID, nil, token)
	rr := newRecorder()

	GetOrganizationHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	if result["id"] != orgID {
		t.Errorf("expected id %s, got %v", orgID, result["id"])
	}
}

func TestGetOrganization_NotFound(t *testing.T) {
	resetDB()
	_, token := createTestUser("freelancer")

	req := authRequest("GET", "/api/v1/organizations/nonexistent", nil, token)
	rr := newRecorder()

	GetOrganizationHandler(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateOrganization(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	orgID := createTestOrg(user.ID)

	body := jsonBody(map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated description",
	})
	req := authRequest("PUT", "/api/v1/organizations/"+orgID, body, token)
	rr := newRecorder()

	UpdateOrganizationHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateOrganization_NonOwner(t *testing.T) {
	resetDB()
	owner, _ := createTestUser("freelancer")
	_, token := createTestUser("freelancer")
	orgID := createTestOrg(owner.ID)

	body := jsonBody(map[string]interface{}{"name": "Hacked"})
	req := authRequest("PUT", "/api/v1/organizations/"+orgID, body, token)
	rr := newRecorder()

	UpdateOrganizationHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestInviteMember(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	orgID := createTestOrg(user.ID)

	body := jsonBody(map[string]interface{}{
		"email": "newmember@test.com",
		"role":  "member",
	})
	req := authRequest("POST", "/api/v1/organizations/"+orgID+"/members", body, token)
	rr := newRecorder()

	InviteMemberHandler(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	d := database.GetDB()
	d.Mu.RLock()
	if len(d.OrgInvites) != 1 {
		t.Errorf("expected 1 invite, got %d", len(d.OrgInvites))
	}
	d.Mu.RUnlock()
}

func TestInviteMember_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"email": "test@test.com"})
	req := authRequest("POST", "/api/v1/organizations/fake/members", body, "")
	rr := newRecorder()

	InviteMemberHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRemoveMember(t *testing.T) {
	resetDB()
	owner, _ := createTestUser("freelancer")
	member, _ := createTestUser("freelancer")
	orgID := createTestOrg(owner.ID)

	d := database.GetDB()
	d.Mu.Lock()
	d.OrgMembers = append(d.OrgMembers, models.OrgMember{
		ID: store.NewID(), OrgID: orgID, UserID: member.ID,
		Role: "member", JoinedAt: store.Now(),
	})
	for i := range d.Organizations {
		if d.Organizations[i].ID == orgID {
			d.Organizations[i].MemberCount++
		}
	}
	d.Mu.Unlock()

	ownerToken := store.GenerateToken(owner.ID, owner.Email, owner.Role)

	req := authRequest("DELETE", "/api/v1/organizations/"+orgID+"/members/"+member.ID, nil, ownerToken)
	rr := newRecorder()

	RemoveMemberHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d.Mu.RLock()
	for _, m := range d.OrgMembers {
		if m.OrgID == orgID && m.UserID == member.ID {
			t.Error("member should have been removed")
		}
	}
	d.Mu.RUnlock()
}

func TestRemoveMember_Self(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	orgID := createTestOrg(user.ID)

	req := authRequest("DELETE", "/api/v1/organizations/"+orgID+"/members/"+user.ID, nil, token)
	rr := newRecorder()

	RemoveMemberHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestChangeMemberRole(t *testing.T) {
	resetDB()
	owner, _ := createTestUser("freelancer")
	member, _ := createTestUser("freelancer")
	orgID := createTestOrg(owner.ID)

	d := database.GetDB()
	d.Mu.Lock()
	d.OrgMembers = append(d.OrgMembers, models.OrgMember{
		ID: store.NewID(), OrgID: orgID, UserID: member.ID,
		Role: "member", JoinedAt: store.Now(),
	})
	d.Mu.Unlock()

	ownerToken := store.GenerateToken(owner.ID, owner.Email, owner.Role)

	body := jsonBody(map[string]interface{}{"role": "manager"})
	req := authRequest("PUT", "/api/v1/organizations/"+orgID+"/members/"+member.ID+"/role", body, ownerToken)
	rr := newRecorder()

	ChangeMemberRoleHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	d.Mu.RLock()
	for _, m := range d.OrgMembers {
		if m.OrgID == orgID && m.UserID == member.ID {
			if m.Role != "manager" {
				t.Errorf("expected manager, got %s", m.Role)
			}
		}
	}
	d.Mu.RUnlock()
}

func TestListOrgOrders(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	orgID := createTestOrg(user.ID)

	req := authRequest("GET", "/api/v1/organizations/"+orgID+"/orders", nil, token)
	rr := newRecorder()

	ListOrgOrdersHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOrgAnalytics(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	orgID := createTestOrg(user.ID)

	req := authRequest("GET", "/api/v1/organizations/"+orgID+"/analytics", nil, token)
	rr := newRecorder()

	OrgAnalyticsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["memberCount"].(float64) != 1 {
		t.Errorf("expected 1 member, got %v", result["memberCount"])
	}
}

func TestOrganization_FullLifecycle(t *testing.T) {
	resetDB()
	_, ownerToken := createTestUser("freelancer")
	member, _ := createTestUser("freelancer")

	body := jsonBody(map[string]interface{}{
		"name":        "Test Agency",
		"description": "A test agency",
	})
	createReq := authRequest("POST", "/api/v1/organizations", body, ownerToken)
	createRR := newRecorder()
	CreateOrganizationHandler(createRR, createReq)

	if createRR.Code != 201 {
		t.Fatalf("create: expected 201, got %d", createRR.Code)
	}
	orgID := decodeJSON(createRR)["id"].(string)

	inviteBody := jsonBody(map[string]interface{}{
		"email": member.Email,
		"role":  "member",
	})
	inviteReq := authRequest("POST", "/api/v1/organizations/"+orgID+"/members", inviteBody, ownerToken)
	inviteRR := newRecorder()
	InviteMemberHandler(inviteRR, inviteReq)

	if inviteRR.Code != 201 {
		t.Fatalf("invite: expected 201, got %d", inviteRR.Code)
	}

	d := database.GetDB()
	d.Mu.Lock()
	d.OrgMembers = append(d.OrgMembers, models.OrgMember{
		ID: store.NewID(), OrgID: orgID, UserID: member.ID,
		Role: "member", JoinedAt: store.Now(),
	})
	for i := range d.Organizations {
		if d.Organizations[i].ID == orgID {
			d.Organizations[i].MemberCount++
		}
	}
	d.Mu.Unlock()

	changeBody := jsonBody(map[string]interface{}{"role": "manager"})
	changeReq := authRequest("PUT", "/api/v1/organizations/"+orgID+"/members/"+member.ID+"/role", changeBody, ownerToken)
	changeRR := newRecorder()
	ChangeMemberRoleHandler(changeRR, changeReq)

	if changeRR.Code != 200 {
		t.Fatalf("change role: expected 200, got %d", changeRR.Code)
	}

	removeReq := authRequest("DELETE", "/api/v1/organizations/"+orgID+"/members/"+member.ID, nil, ownerToken)
	removeRR := newRecorder()
	RemoveMemberHandler(removeRR, removeReq)

	if removeRR.Code != 200 {
		t.Fatalf("remove: expected 200, got %d", removeRR.Code)
	}
}

func createTestOrg(ownerID string) string {
	orgID := store.NewID()
	d := database.GetDB()
	d.Mu.Lock()
	d.Organizations = append(d.Organizations, models.Organization{
		ID: orgID, Name: "Test Org", Slug: "test-org-"+orgID[:8],
		OwnerID: ownerID, MemberCount: 1, CreatedAt: store.Now(),
	})
	d.OrgMembers = append(d.OrgMembers, models.OrgMember{
		ID: store.NewID(), OrgID: orgID, UserID: ownerID,
		Role: "owner", JoinedAt: store.Now(),
	})
	d.Mu.Unlock()
	return orgID
}
