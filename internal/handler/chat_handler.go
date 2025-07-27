package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"ai-chat-backend/internal/config"
	"ai-chat-backend/internal/service"
	"ai-chat-backend/internal/utils"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/go-playground/validator/v10"
)

type ChatHandler struct {
	chatService *service.ChatService
	validator   *validator.Validate
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		validator:   validator.New(),
	}
}

type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// GetConversations 获取会话列表
func (h *ChatHandler) GetConversations(ctx context.Context, c *app.RequestContext) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	conversations, total, err := h.chatService.GetConversations(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(consts.StatusOK, PaginationResponse{
		Data:       conversations,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// CreateConversation 创建新会话
func (h *ChatHandler) CreateConversation(ctx context.Context, c *app.RequestContext) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req service.CreateConversationRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	conversation, err := h.chatService.CreateConversation(userID.(uint), &req)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(consts.StatusCreated, SuccessResponse{
		Message: "Conversation created successfully",
		Data:    conversation,
	})
}

// GetConversation 获取会话详情
func (h *ChatHandler) GetConversation(ctx context.Context, c *app.RequestContext) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: "Invalid conversation ID"})
		return
	}

	conversation, err := h.chatService.GetConversation(userID.(uint), uint(conversationID))
	if err != nil {
		c.JSON(consts.StatusNotFound, ErrorResponse{Error: "Conversation not found"})
		return
	}

	c.JSON(consts.StatusOK, SuccessResponse{
		Message: "Conversation retrieved successfully",
		Data:    conversation,
	})
}

// UpdateConversation 更新会话
func (h *ChatHandler) UpdateConversation(ctx context.Context, c *app.RequestContext) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: "Invalid conversation ID"})
		return
	}

	var req struct {
		Title string `json:"title" validate:"required,max=100"`
	}

	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err = h.chatService.UpdateConversation(userID.(uint), uint(conversationID), req.Title)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(consts.StatusOK, SuccessResponse{
		Message: "Conversation updated successfully",
	})
}

// DeleteConversation 删除会话
func (h *ChatHandler) DeleteConversation(ctx context.Context, c *app.RequestContext) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: "Invalid conversation ID"})
		return
	}

	err = h.chatService.DeleteConversation(userID.(uint), uint(conversationID))
	if err != nil {
		c.JSON(consts.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(consts.StatusOK, SuccessResponse{
		Message: "Conversation deleted successfully",
	})
}

// GetMessages 获取消息列表
func (h *ChatHandler) GetMessages(ctx context.Context, c *app.RequestContext) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: "Invalid conversation ID"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	messages, total, err := h.chatService.GetMessages(userID.(uint), uint(conversationID), page, pageSize)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(consts.StatusOK, PaginationResponse{
		Data:       messages,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// SendMessage 发送消息
func (h *ChatHandler) SendMessage(ctx context.Context, c *app.RequestContext) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: "Invalid conversation ID"})
		return
	}

	var req service.SendMessageRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userMessage, assistantMessage, err := h.chatService.SendMessage(ctx, userID.(uint), uint(conversationID), &req)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(consts.StatusOK, SuccessResponse{
		Message: "Message sent successfully",
		Data: map[string]interface{}{
			"user_message":      userMessage,
			"assistant_message": assistantMessage,
		},
	})
}

// StreamChat 流式聊天
func (h *ChatHandler) StreamChat(ctx context.Context, c *app.RequestContext) {
	// 对于SSE，从URL参数获取token（因为EventSource不支持自定义headers）
	token := c.Query("token")
	if token == "" {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "Token is required"})
		return
	}

	// 移除可能的 "Bearer " 前缀
	tokenString := strings.TrimPrefix(token, "Bearer ")

	// 验证JWT token
	cfg := config.Load()
	claims, err := utils.ValidateJWT(tokenString, cfg.JWT.Secret)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "Invalid token"})
		return
	}

	userID := claims.UserID

	conversationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: "Invalid conversation ID"})
		return
	}

	content := c.Query("content")
	if content == "" {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: "Content is required"})
		return
	}

	// 设置SSE头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 发送开始事件
	c.Write([]byte("data: {\"type\": \"start\"}\n\n"))
	c.Flush()

	// 流式处理
	userMessage, err := h.chatService.StreamChat(ctx, userID, uint(conversationID), content, func(chunk string) error {
		// 正确转义JSON字符串
		chunkBytes, _ := json.Marshal(chunk)
		data := fmt.Sprintf("data: {\"type\": \"chunk\", \"content\": %s}\n\n", string(chunkBytes))
		c.Write([]byte(data))
		c.Flush()
		return nil
	})

	if err != nil {
		errorData := fmt.Sprintf("data: {\"type\": \"error\", \"message\": \"%s\"}\n\n", err.Error())
		c.Write([]byte(errorData))
		c.Flush()
		return
	}

	// 发送结束事件
	endData := fmt.Sprintf("data: {\"type\": \"end\", \"user_message_id\": %d}\n\n", userMessage.ID)
	c.Write([]byte(endData))
	c.Flush()
}