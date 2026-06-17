package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func GenerateGigHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || (user.Role != "freelancer" && user.Role != "admin") {
		Error(w, 403, "Only freelancers can generate gig content")
		return
	}

	var req struct {
		ServiceType    string   `json:"serviceType"`
		Skills         []string `json:"skills"`
		Experience     string   `json:"experience"`
		TargetAudience string   `json:"targetAudience"`
		Tone           string   `json:"tone"`
		PriceMin       float64  `json:"priceMin"`
		PriceMax       float64  `json:"priceMax"`
		UniqueSelling  string   `json:"uniqueSelling"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ServiceType == "" {
		Error(w, 400, "Service type is required")
		return
	}

	if req.Experience == "" {
		req.Experience = "intermediate"
	}
	if req.Tone == "" {
		req.Tone = "professional"
	}
	if req.TargetAudience == "" {
		req.TargetAudience = "businesses"
	}
	if req.PriceMin <= 0 {
		req.PriceMin = 50
	}
	if req.PriceMax <= 0 {
		req.PriceMax = req.PriceMin * 4
	}

	result := generateGigContent(req.ServiceType, req.Skills, req.Experience,
		req.TargetAudience, req.Tone, req.PriceMin, req.PriceMax, req.UniqueSelling)

	JSON(w, 200, result)
}

func ImproveGigHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || (user.Role != "freelancer" && user.Role != "admin") {
		Error(w, 403, "Only freelancers can improve gig content")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Description == "" {
		Error(w, 400, "Description is required")
		return
	}

	suggestions := generateImprovements(req.Title, req.Description)

	JSON(w, 200, H{
		"suggestions": suggestions,
		"improvedTitle":       suggestions["title"],
		"improvedDescription": suggestions["description"],
	})
}

func GenerateFAQHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || (user.Role != "freelancer" && user.Role != "admin") {
		Error(w, 403, "Only freelancers can generate FAQ")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		Error(w, 400, "Title is required")
		return
	}

	faq := generateFAQ(req.Title, req.Description, req.Category)

	JSON(w, 200, H{"faq": faq})
}

func GeneratePackagesHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil || (user.Role != "freelancer" && user.Role != "admin") {
		Error(w, 403, "Only freelancers can generate packages")
		return
	}

	var req struct {
		ServiceType string  `json:"serviceType"`
		PriceMin    float64 `json:"priceMin"`
		PriceMax    float64 `json:"priceMax"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ServiceType == "" {
		Error(w, 400, "Service type is required")
		return
	}

	if req.PriceMin <= 0 {
		req.PriceMin = 25
	}
	if req.PriceMax <= 0 {
		req.PriceMax = req.PriceMin * 5
	}

	packages := generatePackages(req.ServiceType, req.PriceMin, req.PriceMax)

	JSON(w, 200, H{"packages": packages})
}

func SEOSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		Title    string `json:"title"`
		Category string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		Error(w, 400, "Title is required")
		return
	}

	tips := generateSEOTips(req.Title, req.Category)

	JSON(w, 200, H{"tips": tips})
}

func generateGigContent(serviceType string, skills []string, experience, audience, tone string, priceMin, priceMax float64, unique string) map[string]interface{} {
	skillStr := "AI tools"
	if len(skills) > 0 {
		skillStr = strings.Join(skills, ", ")
	}

	title := fmt.Sprintf("I will create professional %s using %s", strings.ToLower(serviceType), skillStr)
	if len(title) > 80 {
		title = title[:77] + "..."
	}

	desc := fmt.Sprintf(
		"Looking for expert %s? You're in the right place!\n\n"+
			"With %s experience in %s, I deliver high-quality results that exceed expectations.\n\n"+
			"What you get:\n"+
			"✅ Professional %s tailored to your needs\n"+
			"✅ Fast turnaround with unlimited revisions\n"+
			"✅ 100%% satisfaction guarantee\n"+
			"✅ Source files included\n\n"+
			"I work with: %s\n\n"+
			"Perfect for %s looking for quality %s work.",
		serviceType, experience, serviceType, serviceType, skillStr, audience, serviceType,
	)
	if unique != "" {
		desc += "\n\nWhy choose me: " + unique
	}

	basicPrice := priceMin
	standardPrice := (priceMin + priceMax) / 2
	premiumPrice := priceMax

	packages := []map[string]interface{}{
		{
			"tier": "basic", "name": "Basic",
			"description":  fmt.Sprintf("Essential %s package", serviceType),
			"price": basicPrice, "deliveryDays": 3, "revisions": 1,
		},
		{
			"tier": "standard", "name": "Standard",
			"description":  fmt.Sprintf("Complete %s with extras", serviceType),
			"price": standardPrice, "deliveryDays": 5, "revisions": 3,
		},
		{
			"tier": "premium", "name": "Premium",
			"description":  fmt.Sprintf("Full %s package with priority support", serviceType),
			"price": premiumPrice, "deliveryDays": 7, "revisions": 5,
		},
	}

	faq := []map[string]string{
		{"question": fmt.Sprintf("What do I need to get started with %s?", serviceType),
			"answer": "Just provide your requirements, brand guidelines (if any), and any reference materials. I'll handle the rest!"},
		{"question": "How many revisions are included?",
			"answer": "Revisions depend on the package you choose. Basic includes 1, Standard includes 3, and Premium includes unlimited revisions."},
		{"question": "What is your typical turnaround time?",
			"answer": "Basic orders are delivered in 3 days, Standard in 5 days, and Premium in 7 days. Rush delivery is available for an additional fee."},
	}

	tags := append(skills, serviceType, "professional", "quality")
	if len(tags) > 7 {
		tags = tags[:7]
	}

	return map[string]interface{}{
		"title":       title,
		"description": desc,
		"faq":         faq,
		"packages":    packages,
		"tags":        tags,
		"seoTips": []string{
			"Use your main keyword in the title and first paragraph",
			"Include specific deliverables in your description",
			"Add relevant tags that match search queries",
			"Mention the tools and software you use",
			"Include a clear call-to-action",
		},
	}
}

func generateImprovements(title, description string) map[string]string {
	improvedTitle := title
	if !strings.HasPrefix(strings.ToLower(title), "i will") {
		improvedTitle = "I will " + strings.ToLower(title)
	}
	if len(improvedTitle) > 80 {
		improvedTitle = improvedTitle[:77] + "..."
	}

	improvedDesc := description
	if !strings.Contains(description, "✅") && !strings.Contains(description, "What you get") {
		improvedDesc += "\n\nWhat you get:\n✅ Professional quality\n✅ Fast delivery\n✅ Unlimited revisions\n✅ 100% satisfaction guarantee"
	}

	return map[string]string{
		"title":       improvedTitle,
		"description": improvedDesc,
		"titleTip":    "Start with 'I will' for better search visibility",
		"descTip":     "Add bullet points with checkmarks for scannability",
	}
}

func generateFAQ(title, description, category string) []map[string]string {
	return []map[string]string{
		{"question": "What information do you need from me?",
			"answer": "Please provide your project requirements, any brand guidelines, preferred style, and deadline. The more details, the better the result!"},
		{"question": "How many revisions are included?",
			"answer": "It depends on the package. Basic includes 1 revision, Standard includes 3, and Premium includes unlimited revisions."},
		{"question": "What is your refund policy?",
			"answer": "If I haven't started working on your order, you can cancel for a full refund. Once work begins, partial refunds may apply."},
		{"question": "Can I see samples of your previous work?",
			"answer": "Yes! Check my portfolio and gig images for examples. Feel free to message me for specific samples related to your project."},
	}
}

func generatePackages(serviceType string, priceMin, priceMax float64) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"tier": "basic", "name": "Starter",
			"description":  fmt.Sprintf("Essential %s - perfect for trying out my services", serviceType),
			"price": priceMin, "deliveryDays": 3, "revisions": 1,
		},
		{
			"tier": "standard", "name": "Professional",
			"description":  fmt.Sprintf("Complete %s with additional features and faster delivery", serviceType),
			"price": (priceMin + priceMax) / 2, "deliveryDays": 5, "revisions": 3,
		},
		{
			"tier": "premium", "name": "Enterprise",
			"description":  fmt.Sprintf("Full-service %s with priority support and unlimited revisions", serviceType),
			"price": priceMax, "deliveryDays": 7, "revisions": 5,
		},
	}
}

func generateSEOTips(title, category string) []string {
	tips := []string{
		"Include your main keyword in the first 60 characters of your title",
		"Use specific deliverables in your title (e.g., '4K video', '3 concepts')",
		"Add 5-7 relevant tags that match what buyers search for",
		"Mention specific tools you use (e.g., 'Midjourney', 'Sora 2')",
		"Include a clear call-to-action in your description",
		"Use bullet points and formatting for readability",
		"Add FAQ entries to capture long-tail search queries",
	}
	if category != "" {
		tips = append(tips, fmt.Sprintf("Research top gigs in '%s' for keyword inspiration", category))
	}
	return tips
}
