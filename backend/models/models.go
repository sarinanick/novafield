package models

import (
	"sync"
	"time"
)

type User struct {
	ID           string   `json:"id"`
	Email        string   `json:"email"`
	PasswordHash string   `json:"-"`
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Avatar       string   `json:"avatar"`
	Bio          string   `json:"bio"`
	Skills       []string `json:"skills"`
	HourlyRate   float64  `json:"hourlyRate"`
	Rating       float64  `json:"rating"`
	ReviewsCount int      `json:"reviewsCount"`
	Earnings     float64  `json:"earnings"`
	Spent        float64  `json:"spent"`
	Location     string   `json:"location"`
	Website      string   `json:"website"`
	Language     string   `json:"language"`
	JoinedAt     string   `json:"joinedAt"`
	IsVerified           bool     `json:"isVerified"`
	IsOnline             bool     `json:"isOnline"`
	SpotifyAccessToken   string   `json:"-"`
	SpotifyRefreshToken  string   `json:"-"`
	SpotifyTokenExpiry   int64    `json:"-"`
	SpotifyConnected     bool     `json:"spotifyConnected"`
	SpotifySharing       bool     `json:"spotifySharing"`
	SpotifyTrackName     string   `json:"spotifyTrackName,omitempty"`
	SpotifyTrackArtist   string   `json:"spotifyTrackArtist,omitempty"`
	SpotifyAlbumArt      string   `json:"spotifyAlbumArt,omitempty"`
	SpotifyTrackURL      string   `json:"spotifyTrackUrl,omitempty"`
}

type UserPublic struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Avatar       string   `json:"avatar"`
	Bio          string   `json:"bio"`
	Skills       []string `json:"skills"`
	HourlyRate   float64  `json:"hourlyRate"`
	Rating       float64  `json:"rating"`
	ReviewsCount int      `json:"reviewsCount"`
	Location     string   `json:"location"`
	JoinedAt     string   `json:"joinedAt"`
	IsVerified   bool     `json:"isVerified"`
}

type Gig struct {
	ID           string       `json:"id"`
	FreelancerID string       `json:"freelancerId"`
	Freelancer   *UserPublic  `json:"freelancer,omitempty"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Category     string       `json:"category"`
	Subcategory  string       `json:"subcategory"`
	Tags         []string     `json:"tags"`
	AITools      []string     `json:"aiTools"`
	PriceType    string       `json:"priceType"`
	Price        float64      `json:"price"`
	DeliveryDays int          `json:"deliveryDays"`
	Revisions    int          `json:"revisions"`
	Images       []string     `json:"images"`
	VideoURL     string       `json:"videoUrl"`
	Status       string       `json:"status"`
	OrdersCount  int          `json:"ordersCount"`
	Views        int          `json:"views"`
	Featured     bool         `json:"featured"`
	Rating       float64      `json:"rating"`
	ReviewsCount int          `json:"reviewsCount"`
	CreatedAt    string       `json:"createdAt"`
	UpdatedAt    string       `json:"updatedAt"`
	Packages     []Package    `json:"packages,omitempty"`
}

type Package struct {
	ID           string   `json:"id"`
	GigID        string   `json:"gigId"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Price        float64  `json:"price"`
	DeliveryDays int      `json:"deliveryDays"`
	Revisions    int      `json:"revisions"`
	Features     []string `json:"features"`
}

type Order struct {
	ID               string      `json:"id"`
	GigID            string      `json:"gigId"`
	Gig              *Gig        `json:"gig,omitempty"`
	PackageID        string      `json:"packageId"`
	BuyerID          string      `json:"buyerId"`
	Buyer            *UserPublic `json:"buyer,omitempty"`
	SellerID         string      `json:"sellerId"`
	Seller           *UserPublic `json:"seller,omitempty"`
	Status           string      `json:"status"`
	Requirements     string      `json:"requirements"`
	Price            float64     `json:"price"`
	EscrowStatus     string      `json:"escrowStatus"`
	DeliveryFile     string      `json:"deliveryFile"`
	DeliveryNotes    string      `json:"deliveryNotes"`
	CreatedAt        string      `json:"createdAt"`
	DeliveryDeadline string      `json:"deliveryDeadline"`
	DeliveredAt      string      `json:"deliveredAt,omitempty"`
	CompletedAt      string      `json:"completedAt,omitempty"`
	DisputeID        string      `json:"disputeId,omitempty"`
}

type Review struct {
	ID         string      `json:"id"`
	OrderID    string      `json:"orderId"`
	GigID      string      `json:"gigId"`
	ReviewerID string      `json:"reviewerId"`
	Reviewer   *UserPublic `json:"reviewer,omitempty"`
	RevieweeID string      `json:"revieweeId"`
	Rating     int         `json:"rating"`
	Comment    string      `json:"comment"`
	CreatedAt  string      `json:"createdAt"`
}

type Message struct {
	ID         string `json:"id"`
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId"`
	OrderID    string `json:"orderId,omitempty"`
	Content    string `json:"content"`
	FileURL    string `json:"fileUrl,omitempty"`
	IsRead     bool   `json:"isRead"`
	CreatedAt  string `json:"createdAt"`
}

type Conversation struct {
	ID            string      `json:"id"`
	User1ID       string      `json:"user1Id"`
	User2ID       string      `json:"user2Id"`
	OtherUser     *UserPublic `json:"otherUser,omitempty"`
	LastMessage   string      `json:"lastMessage"`
	LastMessageAt string      `json:"lastMessageAt"`
	UnreadCount   int         `json:"unreadCount"`
}

type Notification struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Link      string `json:"link"`
	IsRead    bool   `json:"isRead"`
	CreatedAt string `json:"createdAt"`
}

type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	GigsCount   int    `json:"gigsCount"`
}

type DashboardStats struct {
	TotalOrders     int         `json:"totalOrders"`
	ActiveOrders    int         `json:"activeOrders"`
	CompletedOrders int         `json:"completedOrders"`
	TotalEarnings   float64     `json:"totalEarnings"`
	TotalSpent      float64     `json:"totalSpent"`
	PendingPayouts  float64     `json:"pendingPayouts"`
	AvgRating       float64     `json:"avgRating"`
	TotalReviews    int         `json:"totalReviews"`
	TotalGigs       int         `json:"totalGigs"`
	TotalViews      int         `json:"totalViews"`
	ConversionRate  float64     `json:"conversionRate"`
	RecentOrders    []Order     `json:"recentOrders"`
	MonthlyEarnings []MonthData `json:"monthlyEarnings,omitempty"`
}

type MonthData struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

type PaginatedGigs struct {
	Gigs       []Gig `json:"gigs"`
	Total      int   `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

type Meeting struct {
	ID             string      `json:"id"`
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	OrganizerID    string      `json:"organizerId"`
	Organizer      *UserPublic `json:"organizer,omitempty"`
	RoomID         string      `json:"roomId"`
	StartTime      string      `json:"startTime"`
	EndTime        string      `json:"endTime"`
	Recurring      bool        `json:"recurring"`
	RecurrenceRule string      `json:"recurrenceRule,omitempty"`
	AttendeeIDs    []string    `json:"attendeeIds"`
	Status         string      `json:"status"`
	CreatedAt      string      `json:"createdAt"`
}

type TokenEntry struct {
	UserID    string
	ExpiresAt time.Time
}

type Floor struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Level     int     `json:"level"` // 0 = ground floor
	Zones     []Zone  `json:"zones"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	IsDefault bool    `json:"isDefault"`
	CreatedAt string  `json:"createdAt"`
}

type Zone struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"` // work, meeting, social, lounge, hallway
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	W        float64 `json:"w"`
	H        float64 `json:"h"`
	Color    string  `json:"color,omitempty"`
	Capacity int     `json:"capacity,omitempty"`
}

type OfficeTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Size        string `json:"size"` // small, medium, large
	Layout      string `json:"layout"` // open, private, hybrid
	Zones       []Zone `json:"zones"`
	Thumbnail   string `json:"thumbnail"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type DeskObject struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Scale    float64  `json:"scale"`
	Rotation float64  `json:"rotation"`
}

type Desk struct {
	ID        string      `json:"id"`
	ZoneID    string      `json:"zoneId"`
	Position  Position    `json:"position"`
	OwnerID   string      `json:"ownerId,omitempty"`
	Owner     *UserPublic `json:"owner,omitempty" gorm:"-"`
	IsLocked  bool        `json:"isLocked"`
	Objects   []DeskObject `json:"objects"`
	Color     string      `json:"color"`
	CreatedAt string      `json:"createdAt"`
}

type CoworkingSession struct {
	ID              string       `json:"id"`
	HostID          string       `json:"hostId"`
	Host            *UserPublic  `json:"host,omitempty" gorm:"-"`
	Type            string       `json:"type"` // focused, pomodoro, casual
	Title           string       `json:"title"`
	ZoneID          string       `json:"zoneId"`
	StartTime       string       `json:"startTime"`
	Duration        int          `json:"duration"` // minutes
	ParticipantIDs  []string     `json:"participantIds"`
	MaxParticipants int          `json:"maxParticipants"`
	Status          string       `json:"status"` // active, completed, cancelled
	TimerState      *TimerState  `json:"timerState,omitempty"`
	CreatedAt       string       `json:"createdAt"`
}

type TimerState struct {
	Remaining    int    `json:"remaining"` // seconds
	IsPaused     bool   `json:"isPaused"`
	Phase        string `json:"phase"` // work, break (for pomodoro)
	LastResumed  string `json:"lastResumed,omitempty"`
}

type DB struct {
	Mu        sync.RWMutex
	Favorites map[string]map[string]bool // userID -> gigID -> true
	Tokens    map[string]TokenEntry      // token -> TokenEntry
}

type Dispute struct {
	ID          string       `json:"id"`
	OrderID     string       `json:"orderId"`
	OpenedBy    string       `json:"openedBy"`
	OpenedRole  string       `json:"openedRole"` // "client" or "freelancer"
	Reason      string       `json:"reason"`     // "quality", "scope", "communication", "deadline", "other"
	Description string       `json:"description"`
	Status      string       `json:"status"` // "open", "evidence_pending", "under_review", "resolved", "escalated"
	Resolution  *Resolution  `json:"resolution,omitempty"`
	CreatedAt   string       `json:"createdAt"`
	ResolvedAt  string       `json:"resolvedAt,omitempty"`
}

type DisputeEvidence struct {
	ID        string `json:"id"`
	DisputeID string `json:"disputeId"`
	UserID    string `json:"userId"`
	Type      string `json:"type"`    // "text", "file"
	Content   string `json:"content"` // text or file path
	CreatedAt string `json:"createdAt"`
}

type Resolution struct {
	Ruling      string  `json:"ruling"`       // "full_refund", "full_release", "split"
	SplitClient float64 `json:"splitClient"`  // percentage to client (0-100)
	AdminNote   string  `json:"adminNote"`
	ResolvedBy  string  `json:"resolvedBy"`
}

type ProjectBrief struct {
	ID           string   `json:"id"`
	ClientID     string   `json:"clientId"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Skills       []string `json:"skills"`
	BudgetMin    float64  `json:"budgetMin"`
	BudgetMax    float64  `json:"budgetMax"`
	Timeline     string   `json:"timeline"` // "urgent", "1_week", "1_month", "flexible"
	Status       string   `json:"status"`   // "active", "matched", "expired"
	CreatedAt    string   `json:"createdAt"`
}

type MatchResult struct {
	ID           string        `json:"id"`
	BriefID      string        `json:"briefId"`
	FreelancerID string        `json:"freelancerId"`
	Score        float64       `json:"score"` // 0.0 - 1.0
	Reasons      []MatchReason `json:"reasons"`
	Rank         int           `json:"rank"`
	Status       string        `json:"status"` // "suggested", "viewed", "contacted", "hired", "dismissed"
	CreatedAt    string        `json:"createdAt"`
}

type MatchReason struct {
	Factor string  `json:"factor"`
	Weight float64 `json:"weight"`
	Detail string  `json:"detail"`
}

type SubscriptionPlan struct {
	ID           string   `json:"id"`
	GigID        string   `json:"gigId"`
	Name         string   `json:"name"`
	Interval     string   `json:"interval"` // "monthly", "quarterly"
	Price        float64  `json:"price"`
	Deliverables string   `json:"deliverables"`
	MaxRevisions int      `json:"maxRevisions"`
	IsActive     bool     `json:"isActive"`
	CreatedAt    string   `json:"createdAt"`
}

type Subscription struct {
	ID                 string `json:"id"`
	PlanID             string `json:"planId"`
	ClientID           string `json:"clientId"`
	FreelancerID       string `json:"freelancerId"`
	Status             string `json:"status"` // "active", "paused", "cancelled", "expired"
	CurrentPeriodStart string `json:"currentPeriodStart"`
	CurrentPeriodEnd   string `json:"currentPeriodEnd"`
	NextBillingDate    string `json:"nextBillingDate"`
	TotalPaid          float64 `json:"totalPaid"`
	CreatedAt          string `json:"createdAt"`
	CancelledAt        string `json:"cancelledAt,omitempty"`
}

type SubscriptionDeliverable struct {
	ID             string `json:"id"`
	SubscriptionID string `json:"subscriptionId"`
	PeriodStart    string `json:"periodStart"`
	Description    string `json:"description"`
	Status         string `json:"status"` // "pending", "delivered", "approved"
	DeliveredAt    string `json:"deliveredAt,omitempty"`
}

type GigGenerationRequest struct {
	ServiceType    string   `json:"serviceType"`
	Skills         []string `json:"skills"`
	Experience     string   `json:"experience"`     // "beginner", "intermediate", "expert"
	TargetAudience string   `json:"targetAudience"`
	Tone           string   `json:"tone"` // "professional", "casual", "technical"
	PriceMin       float64  `json:"priceMin"`
	PriceMax       float64  `json:"priceMax"`
	UniqueSelling  string   `json:"uniqueSelling"`
}

type GigGenerationResult struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	FAQ         []FAQEntry          `json:"faq"`
	Packages    []PackageSuggestion `json:"packages"`
	Tags        []string            `json:"tags"`
	SEOTips     []string            `json:"seoTips"`
}

type FAQEntry struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type PackageSuggestion struct {
	Tier         string `json:"tier"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        float64 `json:"price"`
	DeliveryDays int    `json:"deliveryDays"`
	Revisions    int    `json:"revisions"`
}

type Invoice struct {
	ID          string  `json:"id"`
	InvoiceNo   string  `json:"invoiceNo"`
	OrderID     string  `json:"orderId"`
	IssuedTo    string  `json:"issuedTo"`
	IssuedBy    string  `json:"issuedBy"`
	Amount      float64 `json:"amount"`
	PlatformFee float64 `json:"platformFee"`
	NetAmount   float64 `json:"netAmount"`
	Status      string  `json:"status"`
	IssuedAt    string  `json:"issuedAt"`
	PaidAt      string  `json:"paidAt,omitempty"`
}

type RecommendationScore struct {
	UserID    string   `json:"userId"`
	GigID     string   `json:"gigId"`
	Score     float64  `json:"score"` // 0.0 - 1.0
	Reasons   []string `json:"reasons"`
	UpdatedAt string   `json:"updatedAt"`
}

type UserPreference struct {
	UserID           string   `json:"userId"`
	TopCategories    []string `json:"topCategories"`
	PriceRange       [2]int   `json:"priceRange"` // [min, avg_max]
	PreferredSkills  []string `json:"preferredSkills"`
	InteractionCount int      `json:"interactionCount"`
	UpdatedAt        string   `json:"updatedAt"`
}

type RecommendationFeedback struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	GigID     string `json:"gigId"`
	EventType string `json:"eventType"`
	CreatedAt string `json:"createdAt"`
}

type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	OwnerID     string `json:"ownerId"`
	MemberCount int    `json:"memberCount"`
	CreatedAt   string `json:"createdAt"`
}

type OrgMember struct {
	ID       string `json:"id"`
	OrgID    string `json:"orgId"`
	UserID   string `json:"userId"`
	Role     string `json:"role"` // "owner", "manager", "member"
	JoinedAt string `json:"joinedAt"`
}

type OrgInvite struct {
	ID        string `json:"id"`
	OrgID     string `json:"orgId"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	InvitedBy string `json:"invitedBy"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expiresAt"`
}

type CaseStudy struct {
	ID           string   `json:"id"`
	UserID       string   `json:"userId"`
	Title        string   `json:"title"`
	ClientName   string   `json:"clientName"`
	Category     string   `json:"category"`
	Challenge    string   `json:"challenge"`
	Approach     string   `json:"approach"`
	Results      string   `json:"results"`
	Images       []string `json:"images"`
	Skills       []string `json:"skills"`
	Duration     string   `json:"duration"`
	BudgetRange  string   `json:"budgetRange"`
	LinkedGigIDs []string `json:"linkedGigIds"`
	IsPublic     bool     `json:"isPublic"`
	CreatedAt    string   `json:"createdAt"`
}

type Testimonial struct {
	ID          string `json:"id"`
	CaseStudyID string `json:"caseStudyId"`
	AuthorID    string `json:"authorId"`
	Rating      int    `json:"rating"`
	Text        string `json:"text"`
	IsApproved  bool   `json:"isApproved"`
	CreatedAt   string `json:"createdAt"`
}

type Language struct {
	Code       string `json:"code"` // "en", "es", "pt", "ar", "fr"
	Name       string `json:"name"`
	NativeName string `json:"nativeName"`
	IsRTL      bool   `json:"isRtl"`
	IsActive   bool   `json:"isActive"`
}

type TranslationFile struct {
	Lang      string            `json:"lang"`
	Version   int               `json:"version"`
	Strings   map[string]string `json:"strings"`
	UpdatedAt string            `json:"updatedAt"`
}

type GigTranslation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type BillingInfo struct {
	UserID   string `json:"userId"`
	FullName string `json:"fullName"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Country  string `json:"country"`
	ZipCode  string `json:"zipCode"`
	TaxID    string `json:"taxId,omitempty"`
}

type FinancialSummary struct {
	Period       string  `json:"period"`
	TotalEarned  float64 `json:"totalEarned"`
	TotalSpent   float64 `json:"totalSpent"`
	PlatformFees float64 `json:"platformFees"`
	NetIncome    float64 `json:"netIncome"`
	OrderCount   int     `json:"orderCount"`
}

type Verification struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Type      string `json:"type"` // "email", "phone", "identity"
	Status    string `json:"status"` // "pending", "verified", "rejected"
	Code      string `json:"code,omitempty"`
	DocURL    string `json:"docUrl,omitempty"`
	ReviewedBy string `json:"reviewedBy,omitempty"`
	CreatedAt string `json:"createdAt"`
	VerifiedAt string `json:"verifiedAt,omitempty"`
}

type TrustScore struct {
	UserID           string  `json:"userId"`
	Score            float64 `json:"score"`
	EmailVerified    bool    `json:"emailVerified"`
	PhoneVerified    bool    `json:"phoneVerified"`
	IdentityVerified bool    `json:"identityVerified"`
	SkillsVerified   int     `json:"skillsVerified"`
	CompletionRate   float64 `json:"completionRate"`
	AvgRating        float64 `json:"avgRating"`
	ResponseTime     float64 `json:"responseTime"`
	Level            string  `json:"level"`
}

type Milestone struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"orderId"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"` // "pending", "in_progress", "completed", "approved", "paid"
	DueDate     string  `json:"dueDate,omitempty"`
	CompletedAt string  `json:"completedAt,omitempty"`
	ApprovedAt  string  `json:"approvedAt,omitempty"`
	CreatedAt   string  `json:"createdAt"`
}

type ClientBrief struct {
	ID          string   `json:"id"`
	ClientID    string   `json:"clientId"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Skills      []string `json:"skills"`
	BudgetMin   float64  `json:"budgetMin"`
	BudgetMax   float64  `json:"budgetMax"`
	Timeline    string   `json:"timeline"`
	Status      string   `json:"status"` // "open", "in_progress", "closed"
	CreatedAt   string   `json:"createdAt"`
}

type Proposal struct {
	ID           string  `json:"id"`
	BriefID      string  `json:"briefId"`
	FreelancerID string  `json:"freelancerId"`
	CoverLetter  string  `json:"coverLetter"`
	Price        float64 `json:"price"`
	Timeline     string  `json:"timeline"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"createdAt"`
}

type Referral struct {
	ID         string `json:"id"`
	ReferrerID string `json:"referrerId"`
	ReferredID string `json:"referredId"`
	Code       string `json:"code"`
	Status     string `json:"status"` // "pending", "completed"
	CreatedAt  string `json:"createdAt"`
}

type ReferralEarning struct {
	ID         string  `json:"id"`
	ReferrerID string  `json:"referrerId"`
	OrderID    string  `json:"orderId"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"` // "pending", "paid"
	CreatedAt  string  `json:"createdAt"`
}

type Assessment struct {
	ID          string             `json:"id"`
	Category    string             `json:"category"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Questions   []AssessmentQuestion `json:"questions"`
	PassScore   int                `json:"passScore"`
	CreatedAt   string             `json:"createdAt"`
}

type AssessmentQuestion struct {
	ID       string   `json:"id"`
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Correct  int      `json:"correct"`
}

type AssessmentResult struct {
	ID           string `json:"id"`
	UserID       string `json:"userId"`
	AssessmentID string `json:"assessmentId"`
	Score        int    `json:"score"`
	Passed       bool   `json:"passed"`
	Answers      []int  `json:"answers"`
	CompletedAt  string `json:"completedAt"`
}

type Badge struct {
	ID       string `json:"id"`
	UserID   string `json:"userId"`
	Skill    string `json:"skill"`
	Level    string `json:"level"`
	Score    int    `json:"score"`
	EarnedAt string `json:"earnedAt"`
}

type SavedSearch struct {
	ID         string   `json:"id"`
	UserID     string   `json:"userId"`
	Query      string   `json:"query"`
	Category   string   `json:"category"`
	MinPrice   float64  `json:"minPrice"`
	MaxPrice   float64  `json:"maxPrice"`
	MinRating  float64  `json:"minRating"`
	Skills     []string `json:"skills"`
	AlertEmail bool     `json:"alertEmail"`
	CreatedAt  string   `json:"createdAt"`
}

type QualityScore struct {
	GigID       string   `json:"gigId"`
	Score       float64  `json:"score"`
	TitleScore  float64  `json:"titleScore"`
	DescScore   float64  `json:"descScore"`
	SEOScore    float64  `json:"seoScore"`
	TagsScore   float64  `json:"tagsScore"`
	Suggestions []string `json:"suggestions"`
	ScoredAt    string   `json:"scoredAt"`
}

type HelpArticle struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
	Views     int      `json:"views"`
	CreatedAt string   `json:"createdAt"`
}

type SupportTicket struct {
	ID        string       `json:"id"`
	UserID    string       `json:"userId"`
	Subject   string       `json:"subject"`
	Category  string       `json:"category"`
	Priority  string       `json:"priority"` // "low", "medium", "high", "urgent"
	Status    string       `json:"status"`   // "open", "in_progress", "resolved", "closed"
	CreatedAt string       `json:"createdAt"`
	Replies   []TicketReply `json:"replies,omitempty"`
}

type TicketReply struct {
	ID        string `json:"id"`
	TicketID  string `json:"ticketId"`
	UserID    string `json:"userId"`
	Content   string `json:"content"`
	IsStaff   bool   `json:"isStaff"`
	CreatedAt string `json:"createdAt"`
}

type WorkspaceComment struct {
	ID        string `json:"id"`
	OrderID   string `json:"orderId"`
	UserID    string `json:"userId"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

type WorkspaceTask struct {
	ID          string `json:"id"`
	OrderID     string `json:"orderId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	AssignedTo  string `json:"assignedTo"`
	CreatedAt   string `json:"createdAt"`
	CompletedAt string `json:"completedAt,omitempty"`
}

type NotificationPreference struct {
	UserID       string `json:"userId"`
	EmailOrders  bool   `json:"emailOrders"`
	EmailMessages bool  `json:"emailMessages"`
	EmailMarketing bool `json:"emailMarketing"`
	InAppOrders  bool   `json:"inAppOrders"`
	InAppMessages bool  `json:"inAppMessages"`
	QuietHoursStart string `json:"quietHoursStart"` // "22:00"
	QuietHoursEnd   string `json:"quietHoursEnd"`   // "08:00"
	DigestFrequency string `json:"digestFrequency"` // "daily", "weekly", "none"
}

type NotificationDigest struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Type      string `json:"type"` // "daily", "weekly"`
	Content   string `json:"content"`
	SentAt    string `json:"sentAt"`
}
