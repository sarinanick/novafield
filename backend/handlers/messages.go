package handlers

import (
	"encoding/json"
	"net/http"
	"novafield-api/database"
	"novafield-api/models"
	"novafield-api/store"
	"sort"
	"strings"
)

func GetConversationsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	d := database.GetDB()
	d.Mu.RLock()
	var convos []models.Conversation
	for _, c := range d.Conversations {
		if c.User1ID == user.ID || c.User2ID == user.ID {
			conv := c
			if conv.User2ID == user.ID {
				conv.UnreadCount = 0
			}
			convos = append(convos, conv)
		}
	}
	d.Mu.RUnlock()

	for i := range convos {
		otherID := convos[i].User2ID
		if convos[i].User2ID == user.ID {
			otherID = convos[i].User1ID
		}
		other := store.FindUserByID(otherID)
		if other != nil {
			pub := store.ToPublic(*other)
			convos[i].OtherUser = &pub
		}
	}
	if convos == nil {
		convos = []models.Conversation{}
	}
	JSON(w, 200, convos)
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var req struct {
		ReceiverID string `json:"receiverId"`
		Content    string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		Error(w, 400, "Receiver and content required")
		return
	}

	msgID := store.NewID()
	msg := models.Message{
		ID: msgID, SenderID: user.ID, ReceiverID: req.ReceiverID,
		Content: req.Content, CreatedAt: store.Now(),
	}

	d := database.GetDB()
	d.Mu.Lock()

	d.Messages = append(d.Messages, msg)

	var convIdx = -1
	for i := range d.Conversations {
		if (d.Conversations[i].User1ID == user.ID && d.Conversations[i].User2ID == req.ReceiverID) ||
			(d.Conversations[i].User1ID == req.ReceiverID && d.Conversations[i].User2ID == user.ID) {
			convIdx = i
			break
		}
	}

	if convIdx >= 0 {
		if d.Conversations[convIdx].User2ID == user.ID {
			d.Conversations[convIdx].UnreadCount++
		}
		d.Conversations[convIdx].LastMessage = req.Content
		d.Conversations[convIdx].LastMessageAt = store.Now()
		convID := d.Conversations[convIdx].ID
		d.Mu.Unlock()
		d.Save()
		BroadcastMessage(user.ID, req.ReceiverID, msg)
		JSON(w, 201, H{"id": msgID, "conversationId": convID})
	} else {
		convID := store.NewID()
		d.Conversations = append(d.Conversations, models.Conversation{
			ID: convID, User1ID: user.ID, User2ID: req.ReceiverID,
			LastMessage: req.Content, LastMessageAt: store.Now(),
		})
		d.Mu.Unlock()
		d.Save()
		BroadcastMessage(user.ID, req.ReceiverID, msg)
		JSON(w, 201, H{"id": msgID, "conversationId": convID})
	}
}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}
	otherID := strings.TrimPrefix(r.URL.Path, "/api/v1/messages/")
	otherID = strings.TrimSuffix(otherID, "")

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Conversations {
		if (d.Conversations[i].User1ID == user.ID && d.Conversations[i].User2ID == otherID) ||
			(d.Conversations[i].User1ID == otherID && d.Conversations[i].User2ID == user.ID) {
			d.Conversations[i].UnreadCount = 0
			break
		}
	}

	for i := range d.Messages {
		if d.Messages[i].SenderID == otherID && d.Messages[i].ReceiverID == user.ID {
			d.Messages[i].IsRead = true
		}
	}

	var messages []models.Message
	for _, m := range d.Messages {
		if (m.SenderID == user.ID && m.ReceiverID == otherID) ||
			(m.SenderID == otherID && m.ReceiverID == user.ID) {
			messages = append(messages, m)
		}
	}

	d.Mu.Unlock()
	d.Save()

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt < messages[j].CreatedAt
	})

	if messages == nil {
		messages = []models.Message{}
	}
	JSON(w, 200, messages)
}

func GetUnreadCountHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}

	var count int
	d := database.GetDB()
	d.Mu.RLock()
	for _, m := range d.Messages {
		if m.ReceiverID == user.ID && !m.IsRead {
			count++
		}
	}
	d.Mu.RUnlock()
	JSON(w, 200, H{"count": count})
}

func MarkConversationReadHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		Error(w, 401, "Unauthorized")
		return
	}
	otherID := strings.TrimPrefix(r.URL.Path, "/api/v1/conversations/")
	otherID = strings.TrimSuffix(otherID, "/read")

	d := database.GetDB()
	d.Mu.Lock()

	for i := range d.Conversations {
		if (d.Conversations[i].User1ID == user.ID && d.Conversations[i].User2ID == otherID) ||
			(d.Conversations[i].User1ID == otherID && d.Conversations[i].User2ID == user.ID) {
			d.Conversations[i].UnreadCount = 0
			break
		}
	}

	for i := range d.Messages {
		if d.Messages[i].SenderID == otherID && d.Messages[i].ReceiverID == user.ID {
			d.Messages[i].IsRead = true
		}
	}

	d.Mu.Unlock()
	d.Save()

	JSON(w, 200, H{"message": "Conversation marked as read"})
}
