package main

import (
	"fmt"
	"log"
	"net/http"
	"novafield-api/database"
	"novafield-api/handlers"
	"novafield-api/store"
	"os"
	"strings"
)

func main() {
	database.Init()
	database.SeedIfEmpty()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/health", handlers.HealthHandler)
	mux.HandleFunc("/api/v1/auth/register", handlers.RateLimitMiddleware(handlers.LoginLimiter, handlers.RegisterHandler))
	mux.HandleFunc("/api/v1/auth/login", handlers.RateLimitMiddleware(handlers.LoginLimiter, handlers.LoginHandler))
	mux.HandleFunc("/api/v1/categories", handlers.GetCategoriesHandler)
	mux.HandleFunc("/api/v1/freelancers", handlers.GetFreelancersHandler)
	mux.HandleFunc("/api/v1/me", authMiddleware(meRouter))
	mux.HandleFunc("/api/v1/gigs", gigsRouter)
	mux.HandleFunc("/api/v1/gigs/", gigsRouter)
	mux.HandleFunc("/api/v1/users/", userPortfolioRouter)
	mux.HandleFunc("/api/v1/orders", authMiddleware(ordersRouter))
	mux.HandleFunc("/api/v1/orders/", authMiddleware(orderActionRouter))
	mux.HandleFunc("/api/v1/disputes", authMiddleware(disputesRouter))
	mux.HandleFunc("/api/v1/disputes/", authMiddleware(disputeActionRouter))
	mux.HandleFunc("/api/v1/match/freelancers", authMiddleware(handlers.MatchFreelancersHandler))
	mux.HandleFunc("/api/v1/match/projects", authMiddleware(handlers.GetMatchProjectsHandler))
	mux.HandleFunc("/api/v1/match/feedback", authMiddleware(handlers.MatchFeedbackHandler))
	mux.HandleFunc("/api/v1/match/history", authMiddleware(handlers.MatchHistoryHandler))
	mux.HandleFunc("/api/v1/subscriptions", authMiddleware(subscriptionsRouter))
	mux.HandleFunc("/api/v1/subscriptions/", authMiddleware(subscriptionActionRouter))
	mux.HandleFunc("/api/v1/ai/gig/generate", authMiddleware(handlers.GenerateGigHandler))
	mux.HandleFunc("/api/v1/ai/gig/improve", authMiddleware(handlers.ImproveGigHandler))
	mux.HandleFunc("/api/v1/ai/gig/faq", authMiddleware(handlers.GenerateFAQHandler))
	mux.HandleFunc("/api/v1/ai/gig/packages", authMiddleware(handlers.GeneratePackagesHandler))
	mux.HandleFunc("/api/v1/ai/gig/seo", authMiddleware(handlers.SEOSuggestionsHandler))
	mux.HandleFunc("/api/v1/analytics/earnings", authMiddleware(handlers.AnalyticsEarningsHandler))
	mux.HandleFunc("/api/v1/analytics/orders", authMiddleware(handlers.AnalyticsOrdersHandler))
	mux.HandleFunc("/api/v1/analytics/gigs", authMiddleware(handlers.AnalyticsGigsHandler))
	mux.HandleFunc("/api/v1/analytics/profile", authMiddleware(handlers.AnalyticsProfileHandler))
	mux.HandleFunc("/api/v1/analytics/clients", authMiddleware(handlers.AnalyticsClientsHandler))
	mux.HandleFunc("/api/v1/analytics/insights", authMiddleware(handlers.AnalyticsInsightsHandler))
	mux.HandleFunc("/api/v1/analytics/export", authMiddleware(handlers.AnalyticsExportHandler))
	mux.HandleFunc("/api/v1/invoices", authMiddleware(handlers.ListInvoicesHandler))
	mux.HandleFunc("/api/v1/invoices/", authMiddleware(invoiceActionRouter))
	mux.HandleFunc("/api/v1/financials/summary", authMiddleware(handlers.FinancialSummaryHandler))
	mux.HandleFunc("/api/v1/financials/by-category", authMiddleware(handlers.FinancialByCategoryHandler))
	mux.HandleFunc("/api/v1/financials/by-client", authMiddleware(handlers.FinancialByClientHandler))
	mux.HandleFunc("/api/v1/financials/export", authMiddleware(handlers.FinancialExportHandler))
	mux.HandleFunc("/api/v1/financials/tax-summary", authMiddleware(handlers.FinancialTaxSummaryHandler))
	mux.HandleFunc("/api/v1/recommendations/gigs", authMiddleware(handlers.GetRecommendationsHandler))
	mux.HandleFunc("/api/v1/recommendations/similar/", authMiddleware(handlers.GetSimilarGigsHandler))
	mux.HandleFunc("/api/v1/recommendations/feedback", authMiddleware(handlers.RecommendationFeedbackHandler))
	mux.HandleFunc("/api/v1/organizations", authMiddleware(orgsRouter))
	mux.HandleFunc("/api/v1/organizations/", authMiddleware(orgActionRouter))
	mux.HandleFunc("/api/v1/portfolio/cases", authMiddleware(portfolioRouter))
	mux.HandleFunc("/api/v1/portfolio/cases/", authMiddleware(portfolioActionRouter))
	mux.HandleFunc("/api/v1/i18n/languages", handlers.ListLanguagesHandler)
	mux.HandleFunc("/api/v1/i18n/", i18nRouter)
	mux.HandleFunc("/api/v1/users/me/language", authMiddleware(handlers.SetLanguageHandler))
	mux.HandleFunc("/api/v1/conversations", authMiddleware(handlers.GetConversationsHandler))
	mux.HandleFunc("/api/v1/conversations/", authMiddleware(conversationsRouter))
	mux.HandleFunc("/api/v1/messages/unread", authMiddleware(handlers.GetUnreadCountHandler))
	mux.HandleFunc("/api/v1/messages", authMiddleware(handlers.SendMessageHandler))
	mux.HandleFunc("/api/v1/messages/", authMiddleware(handlers.GetMessagesHandler))
	mux.HandleFunc("/api/v1/notifications/read", authMiddleware(handlers.MarkNotificationsReadHandler))
	mux.HandleFunc("/api/v1/notifications", authMiddleware(handlers.GetNotificationsHandler))
	mux.HandleFunc("/api/v1/favorites/", authMiddleware(handlers.ToggleFavoriteHandler))
	mux.HandleFunc("/api/v1/dashboard", authMiddleware(handlers.GetDashboardHandler))
	mux.HandleFunc("/api/v1/earnings", authMiddleware(handlers.GetEarningsHandler))
	mux.HandleFunc("/api/v1/spending", authMiddleware(handlers.GetSpendingHandler))
	mux.HandleFunc("/api/v1/world", handlers.GetWorldStateHandler)
	mux.HandleFunc("/api/v1/world/ws", handlers.HandleWorldWebSocket)
	mux.HandleFunc("/api/v1/realtime/ws", handlers.HandleRealtimeWebSocket)
	mux.HandleFunc("/api/v1/upload", authMiddleware(handlers.UploadHandler))
	mux.HandleFunc("/api/v1/meetings", authMiddleware(meetingsRouter))
	mux.HandleFunc("/api/v1/meetings/", authMiddleware(meetingActionRouter))
	mux.HandleFunc("/api/v1/templates", handlers.TemplatesRouter)
	mux.HandleFunc("/api/v1/templates/", handlers.TemplatesRouter)
	mux.HandleFunc("/api/v1/workspaces/", authMiddleware(handlers.WorkspaceRouter))
	mux.HandleFunc("/api/v1/coworking", authMiddleware(coworkingRouter))
	mux.HandleFunc("/api/v1/coworking/", authMiddleware(coworkingActionRouter))
	mux.HandleFunc("/api/v1/desks", authMiddleware(desksRouter))
	mux.HandleFunc("/api/v1/desks/", authMiddleware(deskActionRouter))
	mux.HandleFunc("/api/v1/floors", authMiddleware(floorsRouter))
	mux.HandleFunc("/api/v1/floors/", authMiddleware(floorActionRouter))
	mux.HandleFunc("/api/v1/spotify/auth", authMiddleware(handlers.SpotifyAuthHandler))
	mux.HandleFunc("/api/v1/spotify/callback", handlers.SpotifyCallbackHandler)
	mux.HandleFunc("/api/v1/spotify/status", authMiddleware(handlers.SpotifyStatusHandler))
	mux.HandleFunc("/api/v1/spotify/share", authMiddleware(handlers.SpotifyShareHandler))
	mux.HandleFunc("/api/v1/spotify/disconnect", authMiddleware(handlers.SpotifyDisconnectHandler))
	mux.HandleFunc("/api/v1/spotify/playback", authMiddleware(handlers.SpotifyPlaybackHandler))
	mux.HandleFunc("/api/v1/spotify/search", authMiddleware(handlers.SpotifySearchHandler))
	mux.HandleFunc("/api/v1/verify/email", authMiddleware(handlers.SendEmailVerificationHandler))
	mux.HandleFunc("/api/v1/verify/email/confirm", authMiddleware(handlers.ConfirmEmailVerificationHandler))
	mux.HandleFunc("/api/v1/verify/phone", authMiddleware(handlers.SendPhoneVerificationHandler))
	mux.HandleFunc("/api/v1/verify/phone/confirm", authMiddleware(handlers.ConfirmPhoneVerificationHandler))
	mux.HandleFunc("/api/v1/verify/identity", authMiddleware(handlers.SubmitIdentityVerificationHandler))
	mux.HandleFunc("/api/v1/verify/status", authMiddleware(handlers.GetVerificationStatusHandler))
	mux.HandleFunc("/api/v1/admin/verify/", authMiddleware(adminVerifyRouter))
	mux.HandleFunc("/api/v1/briefs", authMiddleware(briefsRouter))
	mux.HandleFunc("/api/v1/briefs/", authMiddleware(briefActionRouter))
	mux.HandleFunc("/api/v1/my-briefs", authMiddleware(handlers.ListMyBriefsHandler))
	mux.HandleFunc("/api/v1/referrals/link", authMiddleware(handlers.GetReferralLinkHandler))
	mux.HandleFunc("/api/v1/referrals/apply", authMiddleware(handlers.ApplyReferralCodeHandler))
	mux.HandleFunc("/api/v1/referrals/stats", authMiddleware(handlers.GetReferralStatsHandler))
	mux.HandleFunc("/api/v1/referrals/earnings", authMiddleware(handlers.GetReferralEarningsHandler))
	mux.HandleFunc("/api/v1/assessments", handlers.ListAssessmentsHandler)
	mux.HandleFunc("/api/v1/assessments/", assessmentRouter)
	mux.HandleFunc("/api/v1/searches", authMiddleware(searchesRouter))
	mux.HandleFunc("/api/v1/searches/", authMiddleware(searchActionRouter))
	mux.HandleFunc("/api/v1/searches/trending", handlers.TrendingSearchesHandler)
	mux.HandleFunc("/api/v1/ai/gig/score", authMiddleware(handlers.ScoreGigHandler))
	mux.HandleFunc("/api/v1/help/articles", handlers.ListHelpArticlesHandler)
	mux.HandleFunc("/api/v1/help/articles/", handlers.GetHelpArticleHandler)
	mux.HandleFunc("/api/v1/help/search", handlers.SearchHelpHandler)
	mux.HandleFunc("/api/v1/support/tickets", authMiddleware(supportTicketsRouter))
	mux.HandleFunc("/api/v1/notifications/preferences", authMiddleware(handlers.GetNotificationPreferencesHandler))
	mux.HandleFunc("/api/v1/notifications/digest", authMiddleware(handlers.GetNotificationDigestHandler))
	mux.HandleFunc("/api/v1/admin/support/tickets", authMiddleware(handlers.AdminListTicketsHandler))
	mux.HandleFunc("/api/v1/admin/support/tickets/", authMiddleware(handlers.AdminUpdateTicketHandler))
	mux.HandleFunc("/api/v1/admin", authMiddleware(adminRouter))
	mux.HandleFunc("/api/v1/admin/", authMiddleware(adminActionRouter))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	handler := corsMiddleware(loggingMiddleware(mux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("╔══════════════════════════════════════════╗\n")
	fmt.Printf("║  NovaField AI Marketplace API            ║\n")
	fmt.Printf("║  Running on http://localhost:%s        ║\n", port)
	fmt.Printf("╚══════════════════════════════════════════╝\n")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	origins := map[string]bool{
		"http://localhost:3000":  true,
		"http://localhost:3001":  true,
		"http://127.0.0.1:3000": true,
		"http://127.0.0.1:3001": true,
	}
	if extra := os.Getenv("CORS_ORIGINS"); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			origins[strings.TrimSpace(o)] = true
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		user := store.GetUserByToken(token)
		if user == nil {
			handlers.JSON(w, 401, handlers.H{"error": "Unauthorized"})
			return
		}
		r.Header.Set("X-User-ID", user.ID)
		r.Header.Set("X-User-Role", user.Role)
		next(w, r)
	}
}

func meRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handlers.GetMeHandler(w, r)
	case "PUT":
		handlers.UpdateProfileHandler(w, r)
	default:
		handlers.Error(w, 405, "Method not allowed")
	}
}

func gigsRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/api/v1/gigs" || path == "/api/v1/gigs/" {
		switch r.Method {
		case "GET":
			handlers.GetGigsHandler(w, r)
		case "POST":
			handlers.CreateGigHandler(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}

	// /api/v1/gigs/my-gigs
	trimmed := strings.TrimPrefix(path, "/api/v1/gigs/")
	if trimmed == "my-gigs" {
		handlers.GetMyGigsHandler(w, r)
		return
	}

	// /api/v1/gigs/:id/reviews
	parts := strings.Split(trimmed, "/")
	if len(parts) >= 2 && parts[1] == "reviews" {
		handlers.GetGigReviewsHandler(w, r)
		return
	}

	// /api/v1/gigs/:id/subscriptions
	if len(parts) >= 2 && parts[1] == "subscriptions" {
		if r.Method == "POST" {
			handlers.CreateSubscriptionPlanHandler(w, r)
		} else {
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}

	// /api/v1/gigs/:id/portfolio/:caseId
	if len(parts) >= 3 && parts[1] == "portfolio" {
		if r.Method == "POST" {
			handlers.AttachCaseToGigHandler(w, r)
		} else if r.Method == "DELETE" {
			handlers.DetachCaseFromGigHandler(w, r)
		} else {
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}

	// PUT /api/v1/gigs/:id → edit, DELETE /api/v1/gigs/:id → delete
	if r.Method == "PUT" {
		handlers.EditGigHandler(w, r)
		return
	}
	if r.Method == "DELETE" {
		handlers.DeleteGigHandler(w, r)
		return
	}

	handlers.GetGigHandler(w, r)
}

func ordersRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/orders" || r.URL.Path == "/api/v1/orders/" {
		if r.Method == "POST" {
			handlers.CreateOrderHandler(w, r)
			return
		}
		handlers.GetOrdersHandler(w, r)
		return
	}
}

func orderActionRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/milestones") {
		switch r.Method {
		case "GET":
			handlers.ListMilestonesHandler(w, r)
		case "POST":
			handlers.AddMilestoneHandler(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}

	if strings.Contains(path, "/milestones/") {
		if strings.HasSuffix(path, "/complete") {
			handlers.CompleteMilestoneHandler(w, r)
			return
		}
		if strings.HasSuffix(path, "/approve") {
			handlers.ApproveMilestoneHandler(w, r)
			return
		}
		if r.Method == "PUT" {
			handlers.UpdateMilestoneHandler(w, r)
			return
		}
		handlers.Error(w, 404, "Not found")
		return
	}

	if strings.HasSuffix(path, "/workspace/comments") {
		if r.Method == "POST" {
			handlers.AddWorkspaceCommentHandler(w, r)
		} else {
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}

	if strings.HasSuffix(path, "/workspace/tasks") {
		if r.Method == "POST" {
			handlers.AddWorkspaceTaskHandler(w, r)
		} else {
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}

	if strings.Contains(path, "/workspace/tasks/") {
		if r.Method == "PUT" {
			handlers.UpdateWorkspaceTaskHandler(w, r)
		} else {
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}

	if r.Method != "POST" {
		handlers.Error(w, 405, "Method not allowed")
		return
	}

	if strings.HasSuffix(path, "/deliver") {
		handlers.DeliverOrderHandler(w, r)
	} else if strings.HasSuffix(path, "/approve") {
		handlers.ApproveOrderHandler(w, r)
	} else if strings.HasSuffix(path, "/revision") {
		handlers.RequestRevisionHandler(w, r)
	} else if strings.HasSuffix(path, "/review") {
		handlers.CreateReviewHandler(w, r)
	} else if strings.HasSuffix(path, "/dispute") {
		handlers.OpenDisputeHandler(w, r)
	} else {
		handlers.Error(w, 404, "Not found")
	}
}

func conversationsRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/read") {
		handlers.MarkConversationReadHandler(w, r)
		return
	}
	handlers.Error(w, 404, "Not found")
}

func meetingsRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/meetings" || r.URL.Path == "/api/v1/meetings/" {
		switch r.Method {
		case "GET":
			handlers.GetMeetingsHandler(w, r)
		case "POST":
			handlers.CreateMeetingHandler(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
}

func meetingActionRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	trimmed := strings.TrimPrefix(path, "/api/v1/meetings/")

	if strings.HasSuffix(trimmed, "/join") {
		handlers.JoinMeetingHandler(w, r)
		return
	}
	if strings.HasSuffix(trimmed, "/edit") {
		handlers.UpdateMeetingHandler(w, r)
		return
	}

	switch r.Method {
	case "GET":
		handlers.GetMeetingHandler(w, r)
	case "PUT":
		handlers.UpdateMeetingHandler(w, r)
	case "DELETE":
		handlers.DeleteMeetingHandler(w, r)
	default:
		handlers.Error(w, 405, "Method not allowed")
	}
}

func coworkingRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/coworking" || r.URL.Path == "/api/v1/coworking/" {
		switch r.Method {
		case "GET":
			handlers.ListCoworkingSessions(w, r)
		case "POST":
			handlers.CreateCoworkingSession(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
}

func coworkingActionRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	trimmed := strings.TrimPrefix(path, "/api/v1/coworking/")

	if strings.HasSuffix(trimmed, "/join") {
		handlers.JoinCoworkingSession(w, r)
		return
	}
	if strings.HasSuffix(trimmed, "/leave") {
		handlers.LeaveCoworkingSession(w, r)
		return
	}
	if strings.HasSuffix(trimmed, "/timer") {
		handlers.UpdateCoworkingTimer(w, r)
		return
	}
	if strings.HasSuffix(trimmed, "/end") {
		handlers.EndCoworkingSession(w, r)
		return
	}

	if r.Method == "GET" {
		handlers.GetCoworkingSession(w, r)
		return
	}

	handlers.Error(w, 404, "Not found")
}

func desksRouter(w http.ResponseWriter, r *http.Request) {
	handlers.DesksRouter(w, r)
}

func deskActionRouter(w http.ResponseWriter, r *http.Request) {
	handlers.DeskActionRouter(w, r)
}

func floorsRouter(w http.ResponseWriter, r *http.Request) {
	handlers.FloorsRouter(w, r)
}

func floorActionRouter(w http.ResponseWriter, r *http.Request) {
	handlers.FloorActionRouter(w, r)
}

func adminRouter(w http.ResponseWriter, r *http.Request) {
	handlers.AdminRouter(w, r)
}

func adminActionRouter(w http.ResponseWriter, r *http.Request) {
	handlers.AdminActionRouter(w, r)
}

func disputesRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/disputes" || r.URL.Path == "/api/v1/disputes/" {
		switch r.Method {
		case "GET":
			handlers.ListDisputesHandler(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
}

func disputeActionRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/evidence") {
		if r.Method != "POST" {
			handlers.Error(w, 405, "Method not allowed")
			return
		}
		handlers.SubmitEvidenceHandler(w, r)
		return
	}
	if strings.HasSuffix(path, "/resolve") {
		if r.Method != "POST" {
			handlers.Error(w, 405, "Method not allowed")
			return
		}
		handlers.ResolveDisputeHandler(w, r)
		return
	}
	if strings.HasSuffix(path, "/accept") {
		if r.Method != "POST" {
			handlers.Error(w, 405, "Method not allowed")
			return
		}
		handlers.AcceptResolutionHandler(w, r)
		return
	}

	if strings.Contains(path, "/admin/") {
		handlers.AdminDisputeStatsHandler(w, r)
		return
	}

	switch r.Method {
	case "GET":
		handlers.GetDisputeHandler(w, r)
	default:
		handlers.Error(w, 405, "Method not allowed")
	}
}

func subscriptionsRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/subscriptions" || r.URL.Path == "/api/v1/subscriptions/" {
		switch r.Method {
		case "GET":
			handlers.ListSubscriptionsHandler(w, r)
		case "POST":
			handlers.SubscribeHandler(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
}

func subscriptionActionRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/plan") {
		if r.Method == "PUT" {
			handlers.ChangePlanHandler(w, r)
		} else {
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
	if strings.HasSuffix(path, "/pause") {
		handlers.PauseSubscriptionHandler(w, r)
		return
	}
	if strings.HasSuffix(path, "/cancel") {
		handlers.CancelSubscriptionHandler(w, r)
		return
	}
	if strings.HasSuffix(path, "/deliverables") {
		handlers.GetSubscriptionDeliverablesHandler(w, r)
		return
	}

	switch r.Method {
	case "GET":
		handlers.GetSubscriptionHandler(w, r)
	default:
		handlers.Error(w, 405, "Method not allowed")
	}
}

func invoiceActionRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		handlers.Error(w, 405, "Method not allowed")
		return
	}
	handlers.GetInvoiceHandler(w, r)
}

func orgsRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/organizations" || r.URL.Path == "/api/v1/organizations/" {
		switch r.Method {
		case "POST":
			handlers.CreateOrganizationHandler(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
}

func orgActionRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/members") {
		if r.Method == "POST" {
			handlers.InviteMemberHandler(w, r)
		} else {
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
	if strings.HasSuffix(path, "/orders") {
		handlers.ListOrgOrdersHandler(w, r)
		return
	}
	if strings.HasSuffix(path, "/analytics") {
		handlers.OrgAnalyticsHandler(w, r)
		return
	}

	trimmed := strings.TrimPrefix(path, "/api/v1/organizations/")
	parts := strings.Split(trimmed, "/")
	if len(parts) >= 3 && parts[1] == "members" {
		if len(parts) >= 4 && parts[3] == "role" {
			handlers.ChangeMemberRoleHandler(w, r)
			return
		}
		if r.Method == "DELETE" {
			handlers.RemoveMemberHandler(w, r)
			return
		}
	}

	switch r.Method {
	case "GET":
		handlers.GetOrganizationHandler(w, r)
	case "PUT":
		handlers.UpdateOrganizationHandler(w, r)
	default:
		handlers.Error(w, 405, "Method not allowed")
	}
}

func portfolioRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/portfolio/cases" || r.URL.Path == "/api/v1/portfolio/cases/" {
		switch r.Method {
		case "POST":
			handlers.CreateCaseStudyHandler(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
}

func portfolioActionRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/testimonial") {
		if r.Method == "POST" {
			handlers.SubmitTestimonialHandler(w, r)
		} else {
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}

	switch r.Method {
	case "GET":
		handlers.GetCaseStudyHandler(w, r)
	case "PUT":
		handlers.UpdateCaseStudyHandler(w, r)
	case "DELETE":
		handlers.DeleteCaseStudyHandler(w, r)
	default:
		handlers.Error(w, 405, "Method not allowed")
	}
}

func userPortfolioRouter(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/portfolio") {
		handlers.GetUserPortfolioHandler(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/trust-score") {
		handlers.GetTrustScoreHandler(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/badges") {
		handlers.GetUserBadgesHandler(w, r)
		return
	}
	handlers.GetProfileHandler(w, r)
}

func i18nRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/api/v1/i18n/" || path == "/api/v1/i18n" {
		handlers.ListLanguagesHandler(w, r)
		return
	}
	handlers.GetTranslationsHandler(w, r)
}

func adminVerifyRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/review") {
		handlers.AdminReviewIdentityHandler(w, r)
		return
	}
	handlers.AdminListPendingVerificationsHandler(w, r)
}

func briefsRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/briefs" || r.URL.Path == "/api/v1/briefs/" {
		switch r.Method {
		case "GET":
			handlers.ListBriefsHandler(w, r)
		case "POST":
			handlers.CreateBriefHandler(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
}

func briefActionRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/proposals") {
		switch r.Method {
		case "GET":
			handlers.ListProposalsHandler(w, r)
		case "POST":
			handlers.SubmitProposalHandler(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}

	if strings.Contains(path, "/proposals/") && strings.HasSuffix(path, "/accept") {
		handlers.AcceptProposalHandler(w, r)
		return
	}

	switch r.Method {
	case "GET":
		handlers.GetBriefHandler(w, r)
	default:
		handlers.Error(w, 405, "Method not allowed")
	}
}

func assessmentRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/submit") {
		if r.Method == "POST" {
			handlers.SubmitAssessmentHandler(w, r)
		} else {
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
	if strings.HasSuffix(path, "/start") {
		if r.Method == "POST" {
			handlers.StartAssessmentHandler(w, r)
		} else {
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
	if strings.HasPrefix(path, "/api/v1/assessments/leaderboard/") {
		handlers.GetSkillLeaderboardHandler(w, r)
		return
	}
	if strings.HasSuffix(path, "/results") {
		handlers.GetAssessmentResultsHandler(w, r)
		return
	}

	switch r.Method {
	case "GET":
		handlers.GetAssessmentHandler(w, r)
	default:
		handlers.Error(w, 405, "Method not allowed")
	}
}

func searchesRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/searches" || r.URL.Path == "/api/v1/searches/" {
		switch r.Method {
		case "GET":
			handlers.ListSavedSearchesHandler(w, r)
		case "POST":
			handlers.CreateSavedSearchHandler(w, r)
		default:
			handlers.Error(w, 405, "Method not allowed")
		}
		return
	}
}

func searchActionRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		handlers.DeleteSavedSearchHandler(w, r)
		return
	}
	handlers.Error(w, 405, "Method not allowed")
}

func supportTicketsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handlers.ListMyTicketsHandler(w, r)
	case "POST":
		handlers.CreateTicketHandler(w, r)
	default:
		handlers.Error(w, 405, "Method not allowed")
	}
}
