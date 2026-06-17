package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"novafield-api/store"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

func getAllowedOrigins() map[string]bool {
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
	return origins
}

var allowedOrigins = getAllowedOrigins()

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return allowedOrigins[r.Header.Get("Origin")]
	},
}

type WorldUser struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Avatar   string  `json:"avatar"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Zone     string  `json:"zone"`
	FloorID  string  `json:"floorId"`
	IsMoving bool    `json:"isMoving"`
	Emoji    string  `json:"emoji,omitempty"`
}

type WorldMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Conn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type WorldState struct {
	mu          sync.RWMutex
	users       map[string]*WorldUser
	conns       map[string]*Conn
	lockedZones map[string]bool
	floorZones  map[string]map[string]struct{ X, Y, W, H float64; Name string }
}

var world = &WorldState{
	users:       make(map[string]*WorldUser),
	conns:       make(map[string]*Conn),
	lockedZones: make(map[string]bool),
	floorZones: make(map[string]map[string]struct{ X, Y, W, H float64; Name string }),
}

func init() {
	world.floorZones["floor-ground"] = map[string]struct {
		X, Y, W, H float64
		Name        string
	}{
		"work":    {50, 50, 300, 200, "Work Area"},
		"social":  {400, 50, 200, 200, "Social Lounge"},
		"meeting": {50, 300, 300, 150, "Meeting Room"},
		"lounge":  {400, 300, 200, 150, "Chill Zone"},
	}
	world.floorZones["floor-first"] = map[string]struct {
		X, Y, W, H float64
		Name        string
	}{
		"work":    {50, 50, 350, 250, "Engineering"},
		"meeting": {450, 50, 200, 150, "Conference Room"},
		"social":  {50, 350, 250, 150, "Break Room"},
		"lounge":  {350, 350, 200, 150, "Focus Pods"},
	}
	world.floorZones["floor-rooftop"] = map[string]struct {
		X, Y, W, H float64
		Name        string
	}{
		"work":   {50, 50, 400, 200, "Open Workspace"},
		"lounge": {500, 50, 200, 200, "Garden Lounge"},
		"social": {50, 300, 350, 200, "Event Area"},
	}
}

var (
	WorldWidth  float64 = 800
	WorldHeight float64 = 600
)

func GetZoneName(x, y float64) string {
	for id, z := range world.floorZones["floor-ground"] {
		if x >= z.X && x <= z.X+z.W && y >= z.Y && y <= z.Y+z.H {
			return id
		}
	}
	return "hallway"
}

func HandleWorldWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		Error(w, 401, "Token required")
		return
	}

	authedUser := store.GetUserByToken(token)
	if authedUser == nil {
		Error(w, 401, "Invalid token")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	userID := authedUser.ID
	userName := authedUser.Name
	userAvatar := authedUser.Avatar

	user := &WorldUser{
		ID:     userID,
		Name:   userName,
		Avatar: userAvatar,
		X:      100,
		Y:      100,
		Zone:   "work",
		FloorID: "floor-ground",
	}

	wrappedConn := &Conn{conn: conn}

	world.mu.Lock()
	world.users[userID] = user
	world.conns[userID] = wrappedConn
	world.mu.Unlock()

	log.Printf("User joined: %s (%s)", userName, userID)

	// Send init
	world.mu.RLock()
	allUsers := make([]*WorldUser, 0, len(world.users))
	for _, u := range world.users {
		allUsers = append(allUsers, u)
	}
	initLockedZones := make(map[string]bool)
	for z, l := range world.lockedZones {
		initLockedZones[z] = l
	}
	world.mu.RUnlock()

	wrappedConn.writeJSON(WorldMessage{
		Type:    "init",
		Payload: map[string]interface{}{"user": user, "users": allUsers, "lockedZones": initLockedZones, "floorZones": world.floorZones},
	})

	// Broadcast join
	broadcast(WorldMessage{
		Type:    "user-joined",
		Payload: user,
	}, userID)

	// Read loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		var msg WorldMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "move":
			var pos struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			}
			if payloadStr, ok := msg.Payload.(string); ok {
				json.Unmarshal([]byte(payloadStr), &pos)
			} else if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
				if x, ok := payloadMap["x"].(float64); ok {
					pos.X = x
				}
				if y, ok := payloadMap["y"].(float64); ok {
					pos.Y = y
				}
			}

			world.mu.Lock()
			if u, ok := world.users[userID]; ok {
				u.X = pos.X
				u.Y = pos.Y
				u.Zone = GetZoneName(pos.X, pos.Y)
			}
			world.mu.Unlock()

			broadcast(WorldMessage{
				Type:    "user-moved",
				Payload: map[string]interface{}{"id": userID, "x": pos.X, "y": pos.Y},
			}, "")

		case "emoji":
			world.mu.Lock()
			if u, ok := world.users[userID]; ok {
				if emoji, ok := msg.Payload.(string); ok {
					u.Emoji = emoji
				}
			}
			world.mu.Unlock()

			broadcast(WorldMessage{
				Type:    "user-emoji",
				Payload: map[string]interface{}{"id": userID, "emoji": msg.Payload},
			}, "")

	case "wave":
			broadcast(WorldMessage{
				Type:    "user-waved",
				Payload: map[string]interface{}{"id": userID, "name": userName},
			}, "")

		case "lock-conversation":
			if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
				if zoneID, ok := payloadMap["zone"].(string); ok {
					world.mu.Lock()
					world.lockedZones[zoneID] = true
					world.mu.Unlock()
					broadcast(WorldMessage{
						Type:    "conversation-locked",
						Payload: map[string]interface{}{"zone": zoneID, "lockedBy": userID},
					}, "")
				}
			}

		case "unlock-conversation":
			if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
				if zoneID, ok := payloadMap["zone"].(string); ok {
					world.mu.Lock()
					delete(world.lockedZones, zoneID)
					world.mu.Unlock()
					broadcast(WorldMessage{
						Type:    "conversation-unlocked",
						Payload: map[string]interface{}{"zone": zoneID, "unlockedBy": userID},
					}, "")
				}
			}

		case "voice-offer", "voice-answer", "voice-candidate", "voice-hangup":
			if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
				if targetID, ok := payloadMap["to"].(string); ok {
					world.mu.RLock()
					targetConn, exists := world.conns[targetID]
					world.mu.RUnlock()
					if exists {
						targetConn.writeJSON(WorldMessage{
							Type: msg.Type,
							Payload: map[string]interface{}{
								"from": userID,
								"data": payloadMap["data"],
							},
						})
					}
				}
			}

		case "spotify-track":
			broadcast(WorldMessage{
				Type: "spotify-track",
				Payload: map[string]interface{}{
					"userId":    userID,
					"userName":  userName,
					"trackName": msg.Payload,
				},
			}, "")

		case "switch-floor":
			if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
				if floorID, ok := payloadMap["floorId"].(string); ok {
					world.mu.Lock()
					if u, ok := world.users[userID]; ok {
						u.FloorID = floorID
						u.X = 100
						u.Y = 100
						u.Zone = "work"
					}
					world.mu.Unlock()
					broadcast(WorldMessage{
						Type:    "user-switched-floor",
						Payload: map[string]interface{}{"id": userID, "floorId": floorID, "x": 100, "y": 100},
					}, "")
				}
			}
		}
	}

	// Cleanup
	world.mu.Lock()
	delete(world.users, userID)
	delete(world.conns, userID)
	world.mu.Unlock()

	broadcast(WorldMessage{
		Type:    "user-left",
		Payload: map[string]interface{}{"id": userID, "name": userName},
	}, userID)

	log.Printf("User left: %s (%s)", userName, userID)
}

func (c *Conn) writeJSON(msg WorldMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(msg)
}

func broadcast(msg WorldMessage, excludeID string) {
	world.mu.RLock()
	conns := make(map[string]*Conn)
	for id, conn := range world.conns {
		conns[id] = conn
	}
	world.mu.RUnlock()

	for id, conn := range conns {
		if id != excludeID {
			conn.writeJSON(msg)
		}
	}
}

func GetWorldStateHandler(w http.ResponseWriter, r *http.Request) {
	world.mu.RLock()
	defer world.mu.RUnlock()

	users := make([]*WorldUser, 0, len(world.users))
	for _, u := range world.users {
		users = append(users, u)
	}

	lockedZones := make(map[string]bool)
	for z, l := range world.lockedZones {
		lockedZones[z] = l
	}

	JSON(w, 200, H{
		"users":       users,
		"lockedZones": lockedZones,
		"world":       H{"width": WorldWidth, "height": WorldHeight},
		"floorZones":  world.floorZones,
	})
}
