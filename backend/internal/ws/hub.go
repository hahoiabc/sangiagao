package ws

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub manages all WebSocket connections grouped by conversation.
type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]map[*Client]bool // conversationID -> set of clients
	byUser  map[string]map[*Client]bool // userID -> set of clients
}

func NewHub() *Hub {
	return &Hub{
		rooms:  make(map[string]map[*Client]bool),
		byUser: make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Join(conversationID string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[conversationID] == nil {
		h.rooms[conversationID] = make(map[*Client]bool)
	}
	h.rooms[conversationID][client] = true

	if h.byUser[client.UserID] == nil {
		h.byUser[client.UserID] = make(map[*Client]bool)
	}
	h.byUser[client.UserID][client] = true

	log.Printf("WS: user %s joined conversation %s", client.UserID, conversationID)
}

func (h *Hub) Leave(conversationID string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.rooms[conversationID]; ok {
		delete(room, client)
		if len(room) == 0 {
			delete(h.rooms, conversationID)
		}
	}
	if userClients, ok := h.byUser[client.UserID]; ok {
		delete(userClients, client)
		if len(userClients) == 0 {
			delete(h.byUser, client.UserID)
		}
	}
}

// Broadcast sends a message to all clients in a conversation.
func (h *Hub) Broadcast(conversationID string, event interface{}) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("WS: marshal error: %v", err)
		return
	}

	h.mu.RLock()
	clients := h.rooms[conversationID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.Send <- data:
		default:
			// Client buffer full, close it
			close(client.Send)
			h.Leave(conversationID, client)
		}
	}
}

// BroadcastToUser sends an event to all connections of a specific user.
func (h *Hub) BroadcastToUser(userID string, event interface{}) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients := h.byUser[userID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}
