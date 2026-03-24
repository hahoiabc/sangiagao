package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096

	// Rate limiting: max 10 messages per second per client
	rateLimitBurst    = 10
	rateLimitInterval = time.Second
)

// Client represents a single WebSocket connection.
type Client struct {
	UserID         string
	ConversationID string
	Conn           *websocket.Conn
	Send           chan []byte
	Hub            *Hub
	OnMessage      func(client *Client, msg []byte) // callback for incoming messages

	// Rate limiting
	msgCount  int
	windowEnd time.Time
}

// WSMessage is the generic envelope for WebSocket events.
type WSMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// NewMessageEvent is sent when a new message arrives.
type NewMessageEvent struct {
	Event   string      `json:"event"`
	Message interface{} `json:"message"`
}

// ReadReceiptEvent is sent when messages are marked as read.
type ReadReceiptEvent struct {
	Event          string `json:"event"`
	ConversationID string `json:"conversation_id"`
	ReaderID       string `json:"reader_id"`
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Leave(c.ConversationID, c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WS read error: %v", err)
			}
			break
		}
		if c.OnMessage != nil {
			now := time.Now()
			if now.After(c.windowEnd) {
				c.msgCount = 0
				c.windowEnd = now.Add(rateLimitInterval)
			}
			c.msgCount++
			if c.msgCount > rateLimitBurst {
				log.Printf("WS: rate limit exceeded for user %s", c.UserID)
				continue
			}
			c.OnMessage(c, message)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
