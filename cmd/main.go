package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"chatApp/internal/config"
	"chatApp/internal/db"
	"chatApp/internal/handler"
	"chatApp/internal/middleware"
	"chatApp/internal/repository"
	"chatApp/internal/usecase"
	"chatApp/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("config error:", err)
	}

	database, err := db.NewPostgres(cfg.DBURL)
	if err != nil {
		log.Fatal("db connection error:", err)
	}
	defer database.Close()

	userRepo := repository.NewUserRepository(database)
	userUsecase := usecase.NewUserUsecase(userRepo, cfg.JWTSecret, cfg.TokenExpiryHours)
	userHandler := handler.NewUserHandler(userUsecase)

	// Create WebSocket hub
	hub := websocket.NewHub()

	// Create repositories
	messageRepo := repository.NewMessageRepository(database)
	chatRepo := repository.NewChatRepository(database)

	// Set message repository for the hub to persist messages
	hub.SetMessageRepository(messageRepo)

	go hub.Run()

	// Create WebSocket handler
	wsHandler := websocket.NewWebSocketHandler(hub, messageRepo)

	// Create usecases and handlers
	messageUseCase := usecase.NewMessageUsecase(messageRepo, hub)
	chatUsecase := usecase.NewChatUsecase(chatRepo)
	messageHandler := handler.NewMessageHandler(messageUseCase)
	chatHandler := handler.NewChatHandler(chatUsecase)

	r := gin.Default()	

	// Add middleware
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.CORSMiddleware())

	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)

	// WebSocket route
	r.GET("/ws/:chat_id", wsHandler.HandleConnection)

	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	auth.POST("/messages", messageHandler.SendMessage)
	auth.GET("/messages/:chat_id", messageHandler.GetMessages)
	auth.DELETE("/messages/:message_id", messageHandler.DeleteMessage)

	auth.POST("/chats", chatHandler.CreateChat)
	auth.GET("/chats", chatHandler.GetChats)
	auth.GET("/chats/:chat_id", chatHandler.GetChat)
	auth.POST("/chats/:chat_id/participants", chatHandler.AddParticipant)
	auth.GET("/chats/:chat_id/participants", chatHandler.GetParticipants)

	// WebSocket route
	// auth.GET("/ws/:chat_id", wsHandler.HandleConnection)
	// Graceful shutdown
	go func() {
		log.Printf("server running on :%s", cfg.Port)
		if err := r.Run(":" + cfg.Port); err != nil {
			log.Fatal("server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Close database connection
	if err := database.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server exited")
}
