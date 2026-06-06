package websocket

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"chatApp/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should check the origin
		return true
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub             *Hub
	messageRepo     repository.MessageRepository
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *Hub, messageRepo repository.MessageRepository) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, messageRepo: messageRepo}
}

// HandleConnection handles WebSocket connections for a specific chat
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	// Get chat ID from URL parameter
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	// // Get user ID (in a real app, this would come from authentication middleware)
	// userIDStr := c.Query("user_id")
	// userID, err := strconv.Atoi(userIDStr)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		// Get authenticated user ID from the middleware
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, ok := userIDRaw.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create client
	client := &Client{
		ID:     userID,
		ChatID: chatID,
		Conn:   conn,
		Hub:    h.hub,
		Send:   make(chan *Message, 256),
	}

	// Register client with hub
	client.Hub.register <- client

	// Fetch historical messages from database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	messages, err := h.messageRepo.GetMessagesByChatID(ctx, chatID, 50, 0)
	if err != nil {
		log.Printf("Error fetching historical messages: %v", err)
	} else {
		// Send historical messages to the newly connected client
		// Reverse the order so oldest messages come first
		for i := len(messages) - 1; i >= 0; i-- {
			msg := messages[i]
			client.Send <- &Message{
				ID:       msg.ID,
				ChatID:   msg.ChatID,
				SenderID: msg.SenderID,
				Content:  msg.Content,
				Type:     "history",
			}
		}
	}

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.Conn.WriteJSON(message); err != nil {
			log.Printf("Error writing message: %v", err)
			return
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		content := string(messageBytes)
		log.Printf("Received message from client %d in chat %d: %s", c.ID, c.ChatID, content)

		// Create message object
		msg := &Message{
			ChatID:   c.ChatID,
			SenderID: c.ID,
			Content:  content,
			Type:     "message",
		}

		// Broadcast the received message to all members of the same chat.
		c.Hub.broadcast <- msg
	}
}

// Time allowed to write a message to the peer
const writeWait = 10 * time.Second

// Time allowed to read the next pong message from the peer
const pongWait = 60 * time.Second

// Send pings to peer with this period. Must be less than pongWait
const pingPeriod = (pongWait * 9) / 10

// Maximum message size allowed from peer
const maxMessageSize = 512