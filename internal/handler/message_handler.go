package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"chatApp/internal/model"
	"chatApp/internal/usecase"

	"github.com/gin-gonic/gin"
)

type SendMessageRequest struct {
	ChatID  int    `json:"chat_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type GetMessagesRequest struct {
	ChatID int `uri:"chat_id" binding:"required"`
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

type DeleteMessageRequest struct {
	MessageID int `uri:"message_id" binding:"required"`
}

type MessageHandler struct {
	usecase usecase.MessageUsecase
}

func NewMessageHandler(u usecase.MessageUsecase) *MessageHandler {
	return &MessageHandler{usecase: u}
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
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

	message := model.Message{
		ChatID:   req.ChatID,
		SenderID: userID,
		Content:  req.Content,
	}

	if err := h.usecase.SendMessage(c.Request.Context(), message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to send message", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "message sent"})
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	var req GetMessagesRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
		return
	}

	// Parse query parameters for pagination
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid query parameters", "error": err.Error()})
		return
	}

	messages, err := h.usecase.GetMessagesByChatID(c.Request.Context(), req.ChatID, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch messages", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	var req DeleteMessageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
		return
	}

	if err := h.usecase.DeleteMessage(c.Request.Context(), req.MessageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"message": "message not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete message", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message deleted"})
}
