package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"strings"
)

var supportedLanguages = []models.Language{
	{Code: "en", Name: "English", NativeName: "English", IsActive: true},
	{Code: "es", Name: "Spanish", NativeName: "Español", IsActive: true},
	{Code: "pt", Name: "Portuguese", NativeName: "Português", IsActive: true},
	{Code: "ar", Name: "Arabic", NativeName: "العربية", IsRTL: true, IsActive: true},
	{Code: "fr", Name: "French", NativeName: "Français", IsActive: true},
}

var defaultTranslations = map[string]map[string]string{
	"en": {
		"nav.home":           "Home",
		"nav.gigs":           "Gigs",
		"nav.orders":         "Orders",
		"nav.messages":       "Messages",
		"nav.profile":        "Profile",
		"nav.dashboard":      "Dashboard",
		"gig.create":         "Create Gig",
		"gig.title":          "Title",
		"gig.description":    "Description",
		"gig.price":          "Price",
		"gig.delivery":       "Delivery Time",
		"gig.revisions":      "Revisions",
		"order.place":        "Place Order",
		"order.deliver":      "Deliver",
		"order.approve":      "Approve",
		"order.request_rev":  "Request Revision",
		"auth.login":         "Login",
		"auth.register":      "Register",
		"auth.email":         "Email",
		"auth.password":      "Password",
		"auth.name":          "Name",
		"common.search":      "Search",
		"common.filter":      "Filter",
		"common.save":        "Save",
		"common.cancel":      "Cancel",
		"common.delete":      "Delete",
		"common.edit":        "Edit",
		"common.submit":      "Submit",
		"common.loading":     "Loading...",
		"common.no_results":  "No results found",
		"common.error":       "An error occurred",
		"common.success":     "Success",
	},
	"es": {
		"nav.home":           "Inicio",
		"nav.gigs":           "Servicios",
		"nav.orders":         "Pedidos",
		"nav.messages":       "Mensajes",
		"nav.profile":        "Perfil",
		"nav.dashboard":      "Panel",
		"gig.create":         "Crear Servicio",
		"gig.title":          "Título",
		"gig.description":    "Descripción",
		"gig.price":          "Precio",
		"gig.delivery":       "Tiempo de Entrega",
		"gig.revisions":      "Revisiones",
		"order.place":        "Realizar Pedido",
		"order.deliver":      "Entregar",
		"order.approve":      "Aprobar",
		"order.request_rev":  "Solicitar Revisión",
		"auth.login":         "Iniciar Sesión",
		"auth.register":      "Registrarse",
		"auth.email":         "Correo Electrónico",
		"auth.password":      "Contraseña",
		"auth.name":          "Nombre",
		"common.search":      "Buscar",
		"common.filter":      "Filtrar",
		"common.save":        "Guardar",
		"common.cancel":      "Cancelar",
		"common.delete":      "Eliminar",
		"common.edit":        "Editar",
		"common.submit":      "Enviar",
		"common.loading":     "Cargando...",
		"common.no_results":  "No se encontraron resultados",
		"common.error":       "Ocurrió un error",
		"common.success":     "Éxito",
	},
	"pt": {
		"nav.home":           "Início",
		"nav.gigs":           "Serviços",
		"nav.orders":         "Pedidos",
		"nav.messages":       "Mensagens",
		"nav.profile":        "Perfil",
		"nav.dashboard":      "Painel",
		"gig.create":         "Criar Serviço",
		"auth.login":         "Entrar",
		"auth.register":      "Registrar",
		"common.search":      "Pesquisar",
		"common.save":        "Salvar",
		"common.cancel":      "Cancelar",
	},
	"ar": {
		"nav.home":           "الرئيسية",
		"nav.gigs":           "الخدمات",
		"nav.orders":         "الطلبات",
		"nav.messages":       "الرسائل",
		"nav.profile":        "الملف الشخصي",
		"nav.dashboard":      "لوحة التحكم",
		"gig.create":         "إنشاء خدمة",
		"auth.login":         "تسجيل الدخول",
		"auth.register":      "إنشاء حساب",
		"common.search":      "بحث",
		"common.save":        "حفظ",
		"common.cancel":      "إلغاء",
	},
	"fr": {
		"nav.home":           "Accueil",
		"nav.gigs":           "Services",
		"nav.orders":         "Commandes",
		"nav.messages":       "Messages",
		"nav.profile":        "Profil",
		"nav.dashboard":      "Tableau de bord",
		"gig.create":         "Créer un Service",
		"auth.login":         "Connexion",
		"auth.register":      "S'inscrire",
		"common.search":      "Rechercher",
		"common.save":        "Enregistrer",
		"common.cancel":      "Annuler",
	},
}

func GetTranslationsHandler(w http.ResponseWriter, r *http.Request) {
	lang := strings.TrimPrefix(r.URL.Path, "/api/v1/i18n/")
	if lang == "" || lang == r.URL.Path {
		Error(w, 400, "Language code required")
		return
	}

	translations, ok := defaultTranslations[lang]
	if !ok {
		Error(w, 404, "Language not supported")
		return
	}

	JSON(w, 200, H{
		"lang":      lang,
		"version":   1,
		"strings":   translations,
		"updatedAt": store.Now(),
	})
}

func ListLanguagesHandler(w http.ResponseWriter, r *http.Request) {
	JSON(w, 200, supportedLanguages)
}

func SetLanguageHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Language string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Language == "" {
		Error(w, 400, "Language code required")
		return
	}

	supported := false
	for _, lang := range supportedLanguages {
		if lang.Code == req.Language && lang.IsActive {
			supported = true
			break
		}
	}
	if !supported {
		Error(w, 400, "Language not supported")
		return
	}

	d := database.GetDB()
	d.Mu.Lock()
	for i := range d.Users {
		if d.Users[i].ID == user.ID {
			d.Users[i].Language = req.Language
			break
		}
	}
	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"language": req.Language, "message": "Language preference updated"})
}

func TranslateGigHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	gigID := strings.TrimPrefix(r.URL.Path, "/api/v1/gigs/")
	gigID = strings.TrimSuffix(gigID, "/translate")

	var req struct {
		TargetLang string `json:"targetLang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TargetLang == "" {
		Error(w, 400, "Target language required")
		return
	}

	gig := store.FindGigByID(gigID)
	if gig == nil {
		Error(w, 404, "Gig not found")
		return
	}

	translations, ok := defaultTranslations[req.TargetLang]
	if !ok {
		Error(w, 400, "Target language not supported")
		return
	}

	titlePrefix := translations["gig.create"]
	if titlePrefix == "" {
		titlePrefix = "Service"
	}

	JSON(w, 200, H{
		"gigId":      gigID,
		"targetLang": req.TargetLang,
		"translation": H{
			"title":       gig.Title,
			"description": gig.Description,
			"note":        "Full AI translation requires an external translation service",
		},
		"availableStrings": translations,
	})
}

func AdminUpdateTranslationHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || user.Role != "admin" {
		Error(w, 403, "Admin access required")
		return
	}

	lang := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/i18n/")
	if lang == "" {
		Error(w, 400, "Language code required")
		return
	}

	var req struct {
		Strings map[string]string `json:"strings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, 400, "Invalid request")
		return
	}

	if defaultTranslations[lang] == nil {
		defaultTranslations[lang] = make(map[string]string)
	}
	for k, v := range req.Strings {
		defaultTranslations[lang][k] = v
	}

	JSON(w, 200, H{"message": "Translations updated", "lang": lang, "count": len(req.Strings)})
}
