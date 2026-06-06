package handler

import (
	"net/http"

	"chatApp/internal/model"
	"chatApp/internal/usecase"

	"github.com/gin-gonic/gin"
)

type CreateChatRequest struct {
	Name           string `json:"name"`
	IsGroup        bool   `json:"is_group"`
	ParticipantIDs []int  `json:"participant_ids"`
}

type AddParticipantRequest struct {
	UserID int `json:"user_id" binding:"required"`
}

type ChatIDRequest struct {
	ChatID int `uri:"chat_id" binding:"required"`
}

type ChatHandler struct {
	usecase usecase.ChatUsecase
}

func NewChatHandler(u usecase.ChatUsecase) *ChatHandler {
	return &ChatHandler{usecase: u}
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	var req CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
		return
	}

	if len(req.ParticipantIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "participant_ids is required"})
		return
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "user not authenticated"})
		return
	}

	userID, ok := userIDRaw.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid user identity"})
		return
	}

	chat := model.Chat{
		Name:      req.Name,
		IsGroup:   req.IsGroup,
		CreatedBy: userID,
	}

	createdChat, err := h.usecase.CreateChat(c.Request.Context(), chat, req.ParticipantIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create chat", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"chat": createdChat})
}

func (h *ChatHandler) GetChats(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "user not authenticated"})
		return
	}

	userID, ok := userIDRaw.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid user identity"})
		return
	}

	chats, err := h.usecase.GetChatsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch chats", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chats": chats})
}

func (h *ChatHandler) GetChat(c *gin.Context) {
	var req ChatIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid chat id", "error": err.Error()})
		return
	}

	chat, err := h.usecase.GetChatByID(c.Request.Context(), req.ChatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch chat", "error": err.Error()})
		return
	}

	if chat == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "chat not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chat": chat})
}

func (h *ChatHandler) AddParticipant(c *gin.Context) {
	var uriReq ChatIDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid chat id", "error": err.Error()})
		return
	}

	var req AddParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
		return
	}

	if err := h.usecase.AddParticipant(c.Request.Context(), uriReq.ChatID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to add participant", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "participant added"})
}

func (h *ChatHandler) GetParticipants(c *gin.Context) {
	var req ChatIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid chat id", "error": err.Error()})
		return
	}

	participants, err := h.usecase.GetParticipantsByChatID(c.Request.Context(), req.ChatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch participants", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"participants": participants})
}
