package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"novafield-api/store"
	"sync"

	"github.com/gorilla/websocket"
)

type RealtimeConn struct {
	conn   *websocket.Conn
	userID string
	mu     sync.Mutex
}

type RealtimeHub struct {
	mu      sync.RWMutex
	clients map[string]*RealtimeConn
}

var realtimeHub = &RealtimeHub{
	clients: make(map[string]*RealtimeConn),
}

type RealtimeMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (h *RealtimeHub) Register(userID string, conn *RealtimeConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.clients[userID]; ok {
		old.conn.Close()
	}
	h.clients[userID] = conn
	log.Printf("Realtime: user %s connected", userID)

	for id, c := range h.clients {
		if id != userID {
			c.writeJSON(RealtimeMessage{
				Type:    "online-status",
				Payload: map[string]interface{}{"userId": userID, "online": true},
			})
		}
	}
}

func (h *RealtimeHub) Unregister(userID string) {
	h.mu.Lock()
	delete(h.clients, userID)
	h.mu.Unlock()
	log.Printf("Realtime: user %s disconnected", userID)

	h.mu.RLock()
	conns := make(map[string]*RealtimeConn)
	for id, c := range h.clients {
		conns[id] = c
	}
	h.mu.RUnlock()

	for id, c := range conns {
		if id != userID {
			c.writeJSON(RealtimeMessage{
				Type:    "online-status",
				Payload: map[string]interface{}{"userId": userID, "online": false},
			})
		}
	}
}

func (h *RealtimeHub) SendToUser(userID string, msg RealtimeMessage) {
	h.mu.RLock()
	conn, ok := h.clients[userID]
	h.mu.RUnlock()
	if ok {
		conn.writeJSON(msg)
	}
}

func (h *RealtimeHub) Broadcast(msg RealtimeMessage, excludeID string) {
	h.mu.RLock()
	conns := make(map[string]*RealtimeConn)
	for id, c := range h.clients {
		conns[id] = c
	}
	h.mu.RUnlock()

	for id, c := range conns {
		if id != excludeID {
			c.writeJSON(msg)
		}
	}
}

func (c *RealtimeConn) writeJSON(msg RealtimeMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.WriteJSON(msg)
}

func BroadcastMessage(senderID, receiverID string, message interface{}) {
	realtimeHub.SendToUser(receiverID, RealtimeMessage{
		Type:    "chat-message",
		Payload: message,
	})
}

func BroadcastNotification(userID string, notification interface{}) {
	realtimeHub.SendToUser(userID, RealtimeMessage{
		Type:    "notification",
		Payload: notification,
	})
}

func (h *RealtimeHub) GetOnlineUserIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var ids []string
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}

func HandleRealtimeWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		Error(w, 401, "Token required")
		return
	}

	user := store.GetUserByToken(token)
	if user == nil {
		Error(w, 401, "Invalid token")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Realtime WebSocket upgrade error: %v", err)
		return
	}

	rtConn := &RealtimeConn{conn: conn, userID: user.ID}
	realtimeHub.Register(user.ID, rtConn)
	defer func() {
		realtimeHub.Unregister(user.ID)
		conn.Close()
	}()

	rtConn.writeJSON(RealtimeMessage{
		Type:    "connected",
		Payload: map[string]interface{}{"userId": user.ID, "name": user.Name, "onlineUsers": realtimeHub.GetOnlineUserIDs()},
	})

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Realtime read error: %v", err)
			break
		}

		var msg RealtimeMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "typing":
			if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
				if targetID, ok := payloadMap["receiverId"].(string); ok {
					isTyping, _ := payloadMap["typing"].(bool)
					realtimeHub.SendToUser(targetID, RealtimeMessage{
						Type: "typing",
						Payload: map[string]interface{}{
							"userId":  user.ID,
							"name":    user.Name,
							"typing":  isTyping,
						},
					})
				}
			}
		case "mark-read":
			if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
				if senderID, ok := payloadMap["senderId"].(string); ok {
					realtimeHub.SendToUser(senderID, RealtimeMessage{
						Type:    "message-read",
						Payload: map[string]interface{}{"by": user.ID},
					})
				}
			}
		}
	}
}
