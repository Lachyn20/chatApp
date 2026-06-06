package websocket

import (
	"context"
	"log"
	"sync"
	"time"

	"chatApp/internal/model"
	"chatApp/internal/repository"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients in a specific chat.
type Hub struct {
	// Registered clients by chatID
	clients map[int]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan *Message

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread safety
	mutex sync.RWMutex

	// Message repository for persisting messages
	messageRepo repository.MessageRepository
}

// Message represents a message to be broadcasted
type Message struct {
	ID        int    `json:"id"`
	ChatID    int    `json:"chat_id"`
	SenderID  int    `json:"sender_id"`
	Content   string `json:"content"`
	Type      string `json:"type"` // "message", "user_joined", "user_left", etc.
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	ID     int
	ChatID int
	Conn   *websocket.Conn
	Hub    *Hub
	Send   chan *Message
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int]map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// SetMessageRepository sets the message repository for the hub
func (h *Hub) SetMessageRepository(repo repository.MessageRepository) {
	h.messageRepo = repo
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastToChat(message)
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.clients[client.ChatID] == nil {
		h.clients[client.ChatID] = make(map[*Client]bool)
	}
	h.clients[client.ChatID][client] = true

	log.Printf("Client %d joined chat %d", client.ID, client.ChatID)
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if clients, ok := h.clients[client.ChatID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)
			log.Printf("Client %d left chat %d", client.ID, client.ChatID)

			// Clean up empty chat rooms
			if len(clients) == 0 {
				delete(h.clients, client.ChatID)
			}
		}
	}
}

// broadcastToChat sends a message to all clients in a specific chat
func (h *Hub) broadcastToChat(message *Message) {
	// Save message to database if it's a regular message (not history)
	if message.Type == "message" && h.messageRepo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		dbMessage := model.Message{
			ChatID:    message.ChatID,
			SenderID:  message.SenderID,
			Content:   message.Content,
			IsRead:    false,
			CreatedAt: time.Now(),
		}
		if err := h.messageRepo.SendMessage(ctx, dbMessage); err != nil {
			log.Printf("Error saving message to database: %v", err)
		}
		cancel()
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clients, ok := h.clients[message.ChatID]; ok {
		for client := range clients {
			select {
			case client.Send <- message:
			default:
				h.unregisterClient(client)
			}
		}
	}
}

// BroadcastMessage broadcasts a message to a specific chat
func (h *Hub) BroadcastMessage(chatID int, content string, messageType string) {
	message := &Message{
		ChatID:  chatID,
		Content: content,
		Type:    messageType,
	}
	h.broadcast <- message
}