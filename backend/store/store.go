package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"novafield-api/database"
	"novafield-api/models"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var tokenMu sync.RWMutex

var DB = &models.DB{
	Favorites: make(map[string]map[string]bool),
	Tokens:    make(map[string]models.TokenEntry),
}

var tokenExpiry = 72 * time.Hour

func NewID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func HashPassword(p string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("bcrypt: %v", err))
	}
	return string(hash)
}

func CheckPassword(hash, p string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) == nil
}

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func GenerateToken(userID, email, role string) string {
	token := NewID()
	tokenMu.Lock()
	DB.Tokens[token] = models.TokenEntry{
		UserID:    userID,
		ExpiresAt: time.Now().Add(tokenExpiry),
	}
	tokenMu.Unlock()
	return token
}

func GetUserByToken(token string) *models.User {
	tokenMu.RLock()
	entry, ok := DB.Tokens[token]
	tokenMu.RUnlock()
	if !ok || time.Now().After(entry.ExpiresAt) {
		if ok {
			tokenMu.Lock()
			delete(DB.Tokens, token)
			tokenMu.Unlock()
		}
		return nil
	}
	d := database.GetDB()
	d.Mu.RLock()
	defer d.Mu.RUnlock()
	for i := range d.Users {
		if d.Users[i].ID == entry.UserID {
			u := d.Users[i]
			return &u
		}
	}
	return nil
}

func ToPublic(u models.User) models.UserPublic {
	return models.UserPublic{
		ID: u.ID, Name: u.Name, Role: u.Role, Avatar: u.Avatar,
		Bio: u.Bio, Skills: u.Skills, HourlyRate: u.HourlyRate,
		Rating: u.Rating, ReviewsCount: u.ReviewsCount,
		Location: u.Location, JoinedAt: u.JoinedAt, IsVerified: u.IsVerified,
	}
}

func FindUserByEmail(email string) *models.User {
	d := database.GetDB()
	d.Mu.RLock()
	defer d.Mu.RUnlock()
	for i := range d.Users {
		if d.Users[i].Email == email {
			u := d.Users[i]
			return &u
		}
	}
	return nil
}

func FindUserByID(id string) *models.User {
	d := database.GetDB()
	d.Mu.RLock()
	defer d.Mu.RUnlock()
	for i := range d.Users {
		if d.Users[i].ID == id {
			u := d.Users[i]
			return &u
		}
	}
	return nil
}

func FindGigByID(id string) *models.Gig {
	d := database.GetDB()
	d.Mu.RLock()
	defer d.Mu.RUnlock()
	for i := range d.Gigs {
		if d.Gigs[i].ID == id {
			g := d.Gigs[i]
			return &g
		}
	}
	return nil
}

func FindPackageByID(id string) *models.Package {
	d := database.GetDB()
	d.Mu.RLock()
	defer d.Mu.RUnlock()
	for i := range d.Packages {
		if d.Packages[i].ID == id {
			p := d.Packages[i]
			return &p
		}
	}
	return nil
}

func FindOrderByID(id string) *models.Order {
	d := database.GetDB()
	d.Mu.RLock()
	defer d.Mu.RUnlock()
	for i := range d.Orders {
		if d.Orders[i].ID == id {
			o := d.Orders[i]
			return &o
		}
	}
	return nil
}

func CalcGigRating(gigID string) (float64, int) {
	d := database.GetDB()
	d.Mu.RLock()
	defer d.Mu.RUnlock()
	var sum float64
	var count int
	for _, r := range d.Reviews {
		if r.GigID == gigID {
			sum += float64(r.Rating)
			count++
		}
	}
	if count == 0 {
		return 0, 0
	}
	return math.Round(sum/float64(count)*10) / 10, count
}

func SearchGigs(q, category, minPriceS, maxPriceS, minRatingS, deliveryS, sortBy string, page, limit int) models.PaginatedGigs {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 12
	}

	minPrice, _ := strconv.ParseFloat(minPriceS, 64)
	maxPrice, _ := strconv.ParseFloat(maxPriceS, 64)
	minRating, _ := strconv.ParseFloat(minRatingS, 64)
	delivery, _ := strconv.Atoi(deliveryS)
	if maxPrice == 0 {
		maxPrice = 999999
	}

	d := database.GetDB()
	d.Mu.RLock()
	defer d.Mu.RUnlock()

	var filtered []models.Gig
	for _, g := range d.Gigs {
		if g.Status != "active" {
			continue
		}
		if q != "" {
			ql := strings.ToLower(q)
			if !strings.Contains(strings.ToLower(g.Title), ql) &&
				!strings.Contains(strings.ToLower(g.Description), ql) {
				match := false
				for _, t := range g.Tags {
					if strings.Contains(strings.ToLower(t), ql) {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}
		}
		if category != "" && g.Category != category {
			continue
		}
		if minPrice > 0 && g.Price < minPrice {
			continue
		}
		if maxPrice < 999999 && g.Price > maxPrice {
			continue
		}
		if minRating > 0 && g.Rating < minRating {
			continue
		}
		if delivery > 0 && g.DeliveryDays > delivery {
			continue
		}
		filtered = append(filtered, g)
	}

	switch sortBy {
	case "popular":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].OrdersCount > filtered[j].OrdersCount })
	case "rating":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Rating > filtered[j].Rating })
	case "price-low":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Price < filtered[j].Price })
	case "price-high":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Price > filtered[j].Price })
	default:
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt > filtered[j].CreatedAt })
	}

	total := len(filtered)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	ratingSums := make(map[string]float64)
	ratingCounts := make(map[string]int)
	for _, r := range d.Reviews {
		ratingSums[r.GigID] += float64(r.Rating)
		ratingCounts[r.GigID]++
	}

	result := make([]models.Gig, 0)
	for _, g := range filtered[start:end] {
		gg := g
		for i := range d.Users {
			if d.Users[i].ID == g.FreelancerID {
				pub := ToPublic(d.Users[i])
				gg.Freelancer = &pub
				break
			}
		}
		if count := ratingCounts[g.ID]; count > 0 {
			gg.Rating = math.Round(ratingSums[g.ID]/float64(count)*10) / 10
			gg.ReviewsCount = count
		}
		for i := range d.Packages {
			if d.Packages[i].GigID == g.ID {
				gg.Packages = append(gg.Packages, d.Packages[i])
			}
		}
		result = append(result, gg)
	}

	return models.PaginatedGigs{
		Gigs:       result,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
}
