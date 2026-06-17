package database

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"novafield-api/models"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	db   *FileDB
	once sync.Once
)

type FileDB struct {
	Mu            sync.RWMutex
	Users         []models.User           `json:"users"`
	Gigs          []models.Gig            `json:"gigs"`
	Packages      []models.Package        `json:"packages"`
	Orders        []models.Order          `json:"orders"`
	Reviews       []models.Review         `json:"reviews"`
	Messages      []models.Message        `json:"messages"`
	Conversations []models.Conversation   `json:"conversations"`
	Notifications []models.Notification   `json:"notifications"`
	Categories    []models.Category       `json:"categories"`
	Meetings      []models.Meeting        `json:"meetings"`
	Templates     []models.OfficeTemplate `json:"templates"`
	Desks              []models.Desk              `json:"desks"`
	CoworkingSessions  []models.CoworkingSession  `json:"coworkingSessions"`
	Floors             []models.Floor             `json:"floors"`
	Disputes           []models.Dispute           `json:"disputes"`
	DisputeEvidences   []models.DisputeEvidence   `json:"disputeEvidences"`
	ProjectBriefs      []models.ProjectBrief      `json:"projectBriefs"`
	MatchResults       []models.MatchResult       `json:"matchResults"`
	SubscriptionPlans  []models.SubscriptionPlan  `json:"subscriptionPlans"`
	Subscriptions      []models.Subscription      `json:"subscriptions"`
	SubDeliverables    []models.SubscriptionDeliverable `json:"subDeliverables"`
	Invoices           []models.Invoice           `json:"invoices"`
	BillingInfo        []models.BillingInfo       `json:"billingInfo"`
	Recommendations    []models.RecommendationScore `json:"recommendations"`
	UserPreferences    []models.UserPreference    `json:"userPreferences"`
	RecFeedback        []models.RecommendationFeedback `json:"recFeedback"`
	Organizations      []models.Organization      `json:"organizations"`
	OrgMembers         []models.OrgMember         `json:"orgMembers"`
	OrgInvites         []models.OrgInvite         `json:"orgInvites"`
	CaseStudies        []models.CaseStudy         `json:"caseStudies"`
	Testimonials       []models.Testimonial       `json:"testimonials"`
	Languages          []models.Language          `json:"languages"`
	Translations       []models.TranslationFile   `json:"translations"`
	Verifications      []models.Verification      `json:"verifications"`
	Milestones         []models.Milestone         `json:"milestones"`
	ClientBriefs       []models.ClientBrief       `json:"clientBriefs"`
	Proposals          []models.Proposal          `json:"proposals"`
	Referrals          []models.Referral          `json:"referrals"`
	ReferralEarnings   []models.ReferralEarning   `json:"referralEarnings"`
	Assessments        []models.Assessment        `json:"assessments"`
	AssessmentResults  []models.AssessmentResult  `json:"assessmentResults"`
	Badges             []models.Badge             `json:"badges"`
	SavedSearches      []models.SavedSearch       `json:"savedSearches"`
	QualityScores      []models.QualityScore      `json:"qualityScores"`
	HelpArticles       []models.HelpArticle       `json:"helpArticles"`
	SupportTickets     []models.SupportTicket     `json:"supportTickets"`
	WorkspaceComments  []models.WorkspaceComment  `json:"workspaceComments"`
	WorkspaceTasks     []models.WorkspaceTask     `json:"workspaceTasks"`
	NotificationPrefs  []models.NotificationPreference `json:"notificationPrefs"`
	NotificationDigests []models.NotificationDigest `json:"notificationDigests"`
	filePath           string
}

func GetDB() *FileDB {
	once.Do(func() {
		db = &FileDB{filePath: "novafield.json"}
		db.load()
	})
	return db
}

func (d *FileDB) load() {
	data, err := os.ReadFile(d.filePath)
	if err != nil {
		log.Println("No existing database, starting fresh")
		return
	}
	if err := json.Unmarshal(data, d); err != nil {
		log.Printf("Error loading database: %v", err)
	}
	log.Printf("Loaded: %d users, %d gigs, %d orders", len(d.Users), len(d.Gigs), len(d.Orders))
}

func (d *FileDB) Save() {
	d.Mu.RLock()
	defer d.Mu.RUnlock()

	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		log.Printf("Error marshaling database: %v", err)
		return
	}
	if err := os.WriteFile(d.filePath, data, 0644); err != nil {
		log.Printf("Error writing database: %v", err)
	}
}

func Init() {
	GetDB()
	log.Println("File database initialized")
}

func SeedIfEmpty() {
	d := GetDB()
	d.Mu.Lock()
	defer d.Mu.Unlock()

	if len(d.Users) > 0 {
		return
	}
	log.Println("Seeding database...")
	d.seedCategories()
	d.seedUsers()
	d.seedGigs()
	d.seedTemplates()
	d.seedFloors()
	d.seedDesks()
	d.saveLocked()
	log.Println("Database seeded")
}

func (d *FileDB) saveLocked() {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(d.filePath, data, 0644)
}

func (d *FileDB) seedCategories() {
	d.Categories = []models.Category{
		{ID: "cat-1", Name: "AI Video Generation", Slug: "ai-video", Icon: "film", Description: "Create cinematic videos with AI models like Sora, Kling, Veo"},
		{ID: "cat-2", Name: "AI Image Generation", Slug: "ai-image", Icon: "image", Description: "Generate stunning images with DALL-E, Midjourney, Stable Diffusion"},
		{ID: "cat-3", Name: "AI Voice & Audio", Slug: "ai-audio", Icon: "mic", Description: "Voice cloning, music generation, sound effects"},
		{ID: "cat-4", Name: "AI Animation", Slug: "ai-animation", Icon: "play", Description: "Animated characters, motion graphics, 3D renders"},
		{ID: "cat-5", Name: "AI Chatbots & Agents", Slug: "ai-chatbots", Icon: "bot", Description: "Custom chatbots, AI assistants, automation agents"},
		{ID: "cat-6", Name: "AI Web Development", Slug: "ai-webdev", Icon: "code", Description: "AI-powered websites, landing pages, web apps"},
		{ID: "cat-7", Name: "AI Content Writing", Slug: "ai-writing", Icon: "pen", Description: "Blog posts, copywriting, SEO content with AI"},
		{ID: "cat-8", Name: "AI Data Analysis", Slug: "ai-data", Icon: "chart", Description: "Data visualization, ML models, analytics"},
		{ID: "cat-9", Name: "AI Design & Branding", Slug: "ai-design", Icon: "palette", Description: "Logo design, brand kits, UI/UX with AI tools"},
		{ID: "cat-10", Name: "AI Marketing", Slug: "ai-marketing", Icon: "megaphone", Description: "Ad campaigns, social media, AI marketing strategies"},
		{ID: "cat-11", Name: "AI Consulting", Slug: "ai-consulting", Icon: "brain", Description: "AI strategy, implementation, training"},
		{ID: "cat-12", Name: "AI Integration", Slug: "ai-integration", Icon: "plug", Description: "Connect AI to existing systems, APIs, workflows"},
	}
}

func (d *FileDB) seedUsers() {
	type si struct {
		name, bio, loc, email string
		skills                []string
		rate                  float64
		verified              bool
	}
	sellers := []si{
		{"Sarah Chen", "AI video specialist with 5+ years experience", "San Francisco, CA", "sarah.chen@example.com", []string{"Sora 2", "Kling 3.0", "Cinema"}, 85, true},
		{"Marcus Rivera", "Expert in AI image generation and brand design", "New York, NY", "marcus.r@example.com", []string{"DALL-E 3", "Midjourney", "Recraft"}, 65, true},
		{"Yuki Tanaka", "Voice cloning and AI music production specialist", "Tokyo, Japan", "yuki.t@example.com", []string{"ElevenLabs", "Suno AI", "Udio"}, 75, true},
		{"Alex Kim", "Full-stack AI developer, chatbots and automation", "London, UK", "alex.kim@example.com", []string{"GPT-4", "LangChain", "Claude"}, 95, false},
		{"Priya Sharma", "AI content writer and marketing strategist", "Mumbai, India", "priya.s@example.com", []string{"ChatGPT", "Jasper", "Copy.ai"}, 45, false},
		{"James Wilson", "AI data analyst and visualization expert", "Austin, TX", "james.w@example.com", []string{"Python", "TensorFlow", "Tableau"}, 80, false},
	}
	for _, s := range sellers {
		d.Users = append(d.Users, models.User{
			ID: genID(), Email: s.email, PasswordHash: hashPwd("password123"),
			Name: s.name, Role: "freelancer", Bio: s.bio, Skills: s.skills,
			HourlyRate: s.rate, Location: s.loc, JoinedAt: time.Now().UTC().Format(time.RFC3339),
			IsVerified: s.verified,
		})
	}
}

func (d *FileDB) seedGigs() {
	type gd struct {
		sellerEmail, title, desc, cat, tools string
		price                               float64
		days                                int
		pkgs                                []models.Package
	}
	gigs := []gd{
		{"sarah.chen@example.com", "I will create cinematic AI videos with Sora 2", "Professional AI video generation using Sora 2 and Kling 3.0", "ai-video", "Sora 2, Kling 3.0", 150, 3, []models.Package{
			{Name: "Basic", Description: "30s video, 1080p", Price: 50, DeliveryDays: 3, Revisions: 1, Features: []string{"30 second video", "1080p"}},
			{Name: "Standard", Description: "60s video, 4K", Price: 150, DeliveryDays: 5, Revisions: 2, Features: []string{"60 second video", "4K", "Music"}},
			{Name: "Premium", Description: "120s video, full package", Price: 350, DeliveryDays: 7, Revisions: 5, Features: []string{"120s video", "4K", "Music", "License"}},
		}},
		{"marcus.r@example.com", "I will generate stunning AI images for your brand", "High-quality AI image generation with style transfer", "ai-image", "DALL-E 3, Midjourney", 75, 2, []models.Package{
			{Name: "Basic", Description: "5 images, standard", Price: 25, DeliveryDays: 2, Revisions: 1, Features: []string{"5 images", "Standard quality"}},
			{Name: "Standard", Description: "15 images, high quality", Price: 75, DeliveryDays: 3, Revisions: 2, Features: []string{"15 images", "High quality"}},
			{Name: "Premium", Description: "30 images, ultra", Price: 200, DeliveryDays: 5, Revisions: 5, Features: []string{"30 images", "Ultra quality", "Brand guide"}},
		}},
		{"yuki.t@example.com", "I will clone your voice and create AI music", "Professional voice cloning and AI music production", "ai-audio", "ElevenLabs, Suno AI", 120, 3, []models.Package{
			{Name: "Basic", Description: "1 min voice clone", Price: 40, DeliveryDays: 2, Revisions: 1, Features: []string{"1 minute voice clone"}},
			{Name: "Standard", Description: "3 min voice + music", Price: 120, DeliveryDays: 4, Revisions: 2, Features: []string{"3 minute voice", "Custom music"}},
			{Name: "Premium", Description: "Full production", Price: 300, DeliveryDays: 7, Revisions: 5, Features: []string{"Unlimited voice", "Music album"}},
		}},
		{"alex.kim@example.com", "I will build custom AI chatbots and automation agents", "Enterprise-grade AI chatbots and virtual assistants", "ai-chatbots", "GPT-4, LangChain, Claude", 250, 7, []models.Package{
			{Name: "Basic", Description: "Simple FAQ bot", Price: 100, DeliveryDays: 5, Revisions: 2, Features: []string{"FAQ chatbot", "50 entries"}},
			{Name: "Standard", Description: "Advanced AI agent", Price: 250, DeliveryDays: 10, Revisions: 3, Features: []string{"Advanced agent", "500 entries", "Multi-platform"}},
			{Name: "Premium", Description: "Enterprise solution", Price: 500, DeliveryDays: 14, Revisions: 5, Features: []string{"Enterprise agent", "Unlimited", "API access"}},
		}},
		{"priya.s@example.com", "I will write SEO-optimized content using AI", "Professional AI-powered content writing", "ai-writing", "ChatGPT, Jasper", 60, 2, []models.Package{
			{Name: "Basic", Description: "1 blog post (500 words)", Price: 20, DeliveryDays: 1, Revisions: 1, Features: []string{"500 word post", "SEO optimized"}},
			{Name: "Standard", Description: "3 blog posts", Price: 60, DeliveryDays: 3, Revisions: 2, Features: []string{"3x 1500 words", "Keyword research"}},
			{Name: "Premium", Description: "10 posts + strategy", Price: 150, DeliveryDays: 7, Revisions: 3, Features: []string{"10x 2000 words", "SEO strategy"}},
		}},
		{"james.w@example.com", "I will analyze your data with AI-powered insights", "Transform raw data into actionable insights", "ai-data", "Python, TensorFlow", 180, 5, []models.Package{
			{Name: "Basic", Description: "Basic data report", Price: 80, DeliveryDays: 3, Revisions: 1, Features: []string{"Data report", "5 insights"}},
			{Name: "Standard", Description: "Interactive dashboard", Price: 180, DeliveryDays: 5, Revisions: 2, Features: []string{"Dashboard", "10 metrics"}},
			{Name: "Premium", Description: "Full ML pipeline", Price: 400, DeliveryDays: 10, Revisions: 5, Features: []string{"Custom ML model", "Full pipeline"}},
		}},
	}
	for _, g := range gigs {
		seller := d.findUserByEmail(g.sellerEmail)
		if seller == nil {
			continue
		}
		gigID := genID()
		d.Gigs = append(d.Gigs, models.Gig{
			ID: gigID, FreelancerID: seller.ID, Title: g.title, Description: g.desc,
			Category: g.cat, Tags: []string{g.tools}, AITools: splitTools(g.tools),
			PriceType: "fixed", Price: g.price, DeliveryDays: g.days, Revisions: 3,
			Status: "active", CreatedAt: time.Now().UTC().Format(time.RFC3339),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		})
		for _, pkg := range g.pkgs {
			pkg.ID = genID()
			pkg.GigID = gigID
			d.Packages = append(d.Packages, pkg)
		}
		for i := range d.Categories {
			if d.Categories[i].Slug == g.cat {
				d.Categories[i].GigsCount++
				break
			}
		}
	}
}

func (d *FileDB) seedTemplates() {
	d.Templates = []models.OfficeTemplate{
		{
			ID:          "tpl-startup",
			Name:        "Startup",
			Description: "Compact open plan for small teams. Promotes collaboration and quick communication.",
			Size:        "small",
			Layout:      "open",
			Thumbnail:   "/templates/startup.svg",
			Zones: []models.Zone{
				{ID: "z1", Name: "Work Area", Type: "work", X: 50, Y: 50, W: 300, H: 200, Color: "#3b82f6", Capacity: 10},
				{ID: "z2", Name: "Meeting Room", Type: "meeting", X: 400, Y: 50, W: 200, H: 150, Color: "#10b981", Capacity: 6},
				{ID: "z3", Name: "Social Lounge", Type: "social", X: 50, Y: 300, W: 250, H: 150, Color: "#f59e0b", Capacity: 8},
				{ID: "z4", Name: "Chill Zone", Type: "lounge", X: 350, Y: 300, W: 200, H: 150, Color: "#8b5cf6", Capacity: 6},
			},
		},
		{
			ID:          "tpl-corporate",
			Name:        "Corporate",
			Description: "Hybrid layout balancing private offices with shared collaboration spaces.",
			Size:        "medium",
			Layout:      "hybrid",
			Thumbnail:   "/templates/corporate.svg",
			Zones: []models.Zone{
				{ID: "z1", Name: "Open Workspace", Type: "work", X: 50, Y: 50, W: 250, H: 200, Color: "#3b82f6", Capacity: 20},
				{ID: "z2", Name: "Private Offices", Type: "work", X: 350, Y: 50, W: 200, H: 100, Color: "#6366f1", Capacity: 4},
				{ID: "z3", Name: "Board Room", Type: "meeting", X: 350, Y: 200, W: 200, H: 100, Color: "#10b981", Capacity: 12},
				{ID: "z4", Name: "Break Room", Type: "social", X: 50, Y: 300, W: 200, H: 150, Color: "#f59e0b", Capacity: 10},
				{ID: "z5", Name: "Focus Pods", Type: "lounge", X: 300, Y: 350, W: 250, H: 100, Color: "#8b5cf6", Capacity: 4},
			},
		},
		{
			ID:          "tpl-enterprise",
			Name:        "Enterprise",
			Description: "Large layout with private offices, conference rooms, and recreation areas.",
			Size:        "large",
			Layout:      "private",
			Thumbnail:   "/templates/enterprise.svg",
			Zones: []models.Zone{
				{ID: "z1", Name: "Engineering Wing", Type: "work", X: 50, Y: 50, W: 300, H: 200, Color: "#3b82f6", Capacity: 30},
				{ID: "z2", Name: "Executive Offices", Type: "work", X: 400, Y: 50, W: 150, H: 100, Color: "#6366f1", Capacity: 6},
				{ID: "z3", Name: "Conference Room A", Type: "meeting", X: 400, Y: 200, W: 150, H: 100, Color: "#10b981", Capacity: 16},
				{ID: "z4", Name: "Conference Room B", Type: "meeting", X: 400, Y: 350, W: 150, H: 100, Color: "#10b981", Capacity: 8},
				{ID: "z5", Name: "Cafeteria", Type: "social", X: 50, Y: 300, W: 250, H: 150, Color: "#f59e0b", Capacity: 40},
				{ID: "z6", Name: "Wellness Room", Type: "lounge", X: 300, Y: 300, W: 100, H: 150, Color: "#8b5cf6", Capacity: 4},
			},
		},
		{
			ID:          "tpl-creative",
			Name:        "Creative Studio",
			Description: "Open layout designed for creative teams. Features collaborative spaces and inspiration zones.",
			Size:        "medium",
			Layout:      "open",
			Thumbnail:   "/templates/creative.svg",
			Zones: []models.Zone{
				{ID: "z1", Name: "Design Lab", Type: "work", X: 50, Y: 50, W: 300, H: 250, Color: "#ec4899", Capacity: 15},
				{ID: "z2", Name: "Brainstorm Room", Type: "meeting", X: 400, Y: 50, W: 150, H: 150, Color: "#10b981", Capacity: 8},
				{ID: "z3", Name: "Gallery Wall", Type: "social", X: 50, Y: 350, W: 200, H: 100, Color: "#f59e0b", Capacity: 10},
				{ID: "z4", Name: "Lounge", Type: "lounge", X: 300, Y: 350, W: 250, H: 100, Color: "#8b5cf6", Capacity: 8},
			},
		},
		{
			ID:          "tpl-remote",
			Name:        "Remote Team",
			Description: "Hybrid layout for distributed teams. Emphasizes meeting spaces for virtual collaboration.",
			Size:        "medium",
			Layout:      "hybrid",
			Thumbnail:   "/templates/remote.svg",
			Zones: []models.Zone{
				{ID: "z1", Name: "Hot Desks", Type: "work", X: 50, Y: 50, W: 250, H: 150, Color: "#3b82f6", Capacity: 8},
				{ID: "z2", Name: "Video Call Studio", Type: "meeting", X: 350, Y: 50, W: 200, H: 150, Color: "#10b981", Capacity: 4},
				{ID: "z3", Name: "All-Hands Area", Type: "meeting", X: 50, Y: 250, W: 300, H: 200, Color: "#06b6d4", Capacity: 30},
				{ID: "z4", Name: "Social Hub", Type: "social", X: 400, Y: 300, W: 150, H: 150, Color: "#f59e0b", Capacity: 10},
			},
		},
		{
			ID:          "tpl-event",
			Name:        "Event Space",
			Description: "Large open layout for events, workshops, and presentations. Flexible furniture arrangement.",
			Size:        "large",
			Layout:      "open",
			Thumbnail:   "/templates/event.svg",
			Zones: []models.Zone{
				{ID: "z1", Name: "Main Stage", Type: "work", X: 50, Y: 50, W: 350, H: 200, Color: "#ef4444", Capacity: 50},
				{ID: "z2", Name: "Workshop Area", Type: "meeting", X: 450, Y: 50, W: 100, H: 200, Color: "#10b981", Capacity: 20},
				{ID: "z3", Name: "Networking Lounge", Type: "social", X: 50, Y: 300, W: 300, H: 150, Color: "#f59e0b", Capacity: 30},
				{ID: "z4", Name: "Registration", Type: "hallway", X: 400, Y: 300, W: 150, H: 150, Color: "#6b7280", Capacity: 10},
			},
		},
	}
}

func (d *FileDB) seedFloors() {
	d.Floors = []models.Floor{
		{
			ID:    "floor-ground",
			Name:  "Ground Floor",
			Level: 0,
			Zones: []models.Zone{
				{ID: "z1", Name: "Work Area", Type: "work", X: 50, Y: 50, W: 300, H: 200, Color: "#3b82f6", Capacity: 20},
				{ID: "z2", Name: "Social Lounge", Type: "social", X: 400, Y: 50, W: 200, H: 200, Color: "#f59e0b", Capacity: 10},
				{ID: "z3", Name: "Meeting Room", Type: "meeting", X: 50, Y: 300, W: 300, H: 150, Color: "#10b981", Capacity: 8},
				{ID: "z4", Name: "Chill Zone", Type: "lounge", X: 400, Y: 300, W: 200, H: 150, Color: "#8b5cf6", Capacity: 6},
			},
			Width:     800,
			Height:    600,
			IsDefault: true,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
		{
			ID:    "floor-first",
			Name:  "First Floor",
			Level: 1,
			Zones: []models.Zone{
				{ID: "z1", Name: "Engineering", Type: "work", X: 50, Y: 50, W: 350, H: 250, Color: "#3b82f6", Capacity: 30},
				{ID: "z2", Name: "Conference Room", Type: "meeting", X: 450, Y: 50, W: 200, H: 150, Color: "#10b981", Capacity: 12},
				{ID: "z3", Name: "Break Room", Type: "social", X: 50, Y: 350, W: 250, H: 150, Color: "#f59e0b", Capacity: 10},
				{ID: "z4", Name: "Focus Pods", Type: "lounge", X: 350, Y: 350, W: 200, H: 150, Color: "#8b5cf6", Capacity: 4},
			},
			Width:     800,
			Height:    600,
			IsDefault: false,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
		{
			ID:    "floor-rooftop",
			Name:  "Rooftop",
			Level: 2,
			Zones: []models.Zone{
				{ID: "z1", Name: "Open Workspace", Type: "work", X: 50, Y: 50, W: 400, H: 200, Color: "#06b6d4", Capacity: 15},
				{ID: "z2", Name: "Garden Lounge", Type: "lounge", X: 500, Y: 50, W: 200, H: 200, Color: "#10b981", Capacity: 8},
				{ID: "z3", Name: "Event Area", Type: "social", X: 50, Y: 300, W: 350, H: 200, Color: "#f59e0b", Capacity: 25},
			},
			Width:     800,
			Height:    600,
			IsDefault: false,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func (d *FileDB) seedDesks() {
	workZonePositions := []models.Position{
		{X: 80, Y: 80}, {X: 140, Y: 80}, {X: 200, Y: 80},
		{X: 80, Y: 140}, {X: 140, Y: 140}, {X: 200, Y: 140},
		{X: 80, Y: 190}, {X: 140, Y: 190}, {X: 200, Y: 190},
		{X: 260, Y: 80}, {X: 260, Y: 140}, {X: 260, Y: 190},
	}
	for _, pos := range workZonePositions {
		d.Desks = append(d.Desks, models.Desk{
			ID:        genID(),
			ZoneID:    "work",
			Position:  pos,
			Color:     "#6366f1",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		})
	}

	meetingZonePositions := []models.Position{
		{X: 100, Y: 350}, {X: 180, Y: 350}, {X: 260, Y: 350},
		{X: 100, Y: 400}, {X: 180, Y: 400},
	}
	for _, pos := range meetingZonePositions {
		d.Desks = append(d.Desks, models.Desk{
			ID:        genID(),
			ZoneID:    "meeting",
			Position:  pos,
			Color:     "#10b981",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func (d *FileDB) findUserByEmail(email string) *models.User {
	for i := range d.Users {
		if d.Users[i].Email == email {
			return &d.Users[i]
		}
	}
	return nil
}

func genID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func splitTools(s string) []string {
	parts := strings.Split(s, ", ")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func hashPwd(p string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(hash)
}
