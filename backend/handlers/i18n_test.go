package handlers

import (
	"encoding/json"
	"testing"
)

func TestGetTranslations_English(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/i18n/en", nil, "")
	rr := newRecorder()

	GetTranslationsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["lang"] != "en" {
		t.Errorf("expected lang en, got %v", result["lang"])
	}

	strings := result["strings"].(map[string]interface{})
	if strings["nav.home"] != "Home" {
		t.Errorf("expected 'Home', got %v", strings["nav.home"])
	}
}

func TestGetTranslations_Spanish(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/i18n/es", nil, "")
	rr := newRecorder()

	GetTranslationsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	strings := result["strings"].(map[string]interface{})
	if strings["nav.home"] != "Inicio" {
		t.Errorf("expected 'Inicio', got %v", strings["nav.home"])
	}
	if strings["auth.login"] != "Iniciar Sesión" {
		t.Errorf("expected 'Iniciar Sesión', got %v", strings["auth.login"])
	}
}

func TestGetTranslations_Arabic(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/i18n/ar", nil, "")
	rr := newRecorder()

	GetTranslationsHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	result := decodeJSON(rr)
	strings := result["strings"].(map[string]interface{})
	if strings["nav.home"] != "الرئيسية" {
		t.Errorf("expected 'الرئيسية', got %v", strings["nav.home"])
	}
}

func TestGetTranslations_Unsupported(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/i18n/xx", nil, "")
	rr := newRecorder()

	GetTranslationsHandler(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestListLanguages(t *testing.T) {
	resetDB()

	req := authRequest("GET", "/api/v1/i18n/languages", nil, "")
	rr := newRecorder()

	ListLanguagesHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var langs []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&langs)
	if len(langs) < 5 {
		t.Fatalf("expected at least 5 languages, got %d", len(langs))
	}

	foundArabic := false
	for _, l := range langs {
		if l["code"] == "ar" {
			foundArabic = true
			if l["isRtl"] != true {
				t.Error("Arabic should be RTL")
			}
		}
	}
	if !foundArabic {
		t.Error("expected Arabic in language list")
	}
}

func TestSetLanguage(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"language": "es",
	})
	req := authRequest("POST", "/api/v1/users/me/language", body, token)
	rr := newRecorder()

	SetLanguageHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["language"] != "es" {
		t.Errorf("expected es, got %v", result["language"])
	}
}

func TestSetLanguage_Unsupported(t *testing.T) {
	resetDB()
	_, token := createTestUser("client")

	body := jsonBody(map[string]interface{}{
		"language": "xx",
	})
	req := authRequest("POST", "/api/v1/users/me/language", body, token)
	rr := newRecorder()

	SetLanguageHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSetLanguage_Unauthorized(t *testing.T) {
	resetDB()

	body := jsonBody(map[string]interface{}{"language": "es"})
	req := authRequest("POST", "/api/v1/users/me/language", body, "")
	rr := newRecorder()

	SetLanguageHandler(rr, req)

	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestTranslateGig(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	gigID, _ := seedTestGig(user.ID)

	body := jsonBody(map[string]interface{}{
		"targetLang": "es",
	})
	req := authRequest("POST", "/api/v1/gigs/"+gigID+"/translate", body, token)
	rr := newRecorder()

	TranslateGigHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	result := decodeJSON(rr)
	if result["gigId"] != gigID {
		t.Errorf("expected gigId %s, got %v", gigID, result["gigId"])
	}
	if result["targetLang"] != "es" {
		t.Errorf("expected es, got %v", result["targetLang"])
	}
}

func TestTranslateGig_UnsupportedLang(t *testing.T) {
	resetDB()
	user, token := createTestUser("freelancer")
	gigID, _ := seedTestGig(user.ID)

	body := jsonBody(map[string]interface{}{
		"targetLang": "xx",
	})
	req := authRequest("POST", "/api/v1/gigs/"+gigID+"/translate", body, token)
	rr := newRecorder()

	TranslateGigHandler(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAllLanguages_HaveCommonKeys(t *testing.T) {
	requiredKeys := []string{"nav.home", "auth.login", "common.search"}

	for lang, translations := range defaultTranslations {
		for _, key := range requiredKeys {
			if _, ok := translations[key]; !ok {
				t.Errorf("language %s missing required key %s", lang, key)
			}
		}
	}
}
