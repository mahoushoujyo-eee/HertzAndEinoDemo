package service

import (
	"context"
	"fmt"

	"ai-chat-backend/internal/model"

	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

type ChatService struct {
	db        *gorm.DB
	aiService *AIService
}

func NewChatService(db *gorm.DB, aiService *AIService) *ChatService {
	return &ChatService{
		db:        db,
		aiService: aiService,
	}
}

type CreateConversationRequest struct {
	Title string `json:"title" validate:"required,max=100"`
}

type SendMessageRequest struct {
	Content string `json:"content" validate:"required,max=4000"`
}

// GetConversations 获取用户的会话列表
func (s *ChatService) GetConversations(userID uint, page, pageSize int) ([]model.Conversation, int64, error) {
	var conversations []model.Conversation
	var total int64

	query := s.db.Where("user_id = ?", userID)

	// 获取总数
	if err := query.Model(&model.Conversation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&conversations).Error; err != nil {
		return nil, 0, err
	}

	return conversations, total, nil
}

// CreateConversation 创建新会话
func (s *ChatService) CreateConversation(userID uint, req *CreateConversationRequest) (*model.Conversation, error) {
	conversation := model.Conversation{
		UserID: userID,
		Title:  req.Title,
	}

	if err := s.db.Create(&conversation).Error; err != nil {
		return nil, err
	}

	return &conversation, nil
}

// GetConversation 获取会话详情
func (s *ChatService) GetConversation(userID, conversationID uint) (*model.Conversation, error) {
	var conversation model.Conversation
	if err := s.db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}

// UpdateConversation 更新会话
func (s *ChatService) UpdateConversation(userID, conversationID uint, title string) error {
	return s.db.Model(&model.Conversation{}).Where("id = ? AND user_id = ?", conversationID, userID).Update("title", title).Error
}

// DeleteConversation 删除会话
func (s *ChatService) DeleteConversation(userID, conversationID uint) error {
	// 删除会话及其所有消息
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除消息
	if err := tx.Where("conversation_id = ?", conversationID).Delete(&model.Message{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除会话
	if err := tx.Where("id = ? AND user_id = ?", conversationID, userID).Delete(&model.Conversation{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetMessages 获取会话消息
func (s *ChatService) GetMessages(userID, conversationID uint, page, pageSize int) ([]model.Message, int64, error) {
	// 验证会话是否属于用户
	var conversation model.Conversation
	if err := s.db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		return nil, 0, err
	}

	var messages []model.Message
	var total int64

	query := s.db.Where("conversation_id = ?", conversationID)

	// 获取总数
	if err := query.Model(&model.Message{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// SendMessage 发送消息并获取AI回复
func (s *ChatService) SendMessage(ctx context.Context, userID, conversationID uint, req *SendMessageRequest) (*model.Message, *model.Message, error) {
	// 验证会话是否属于用户
	var conversation model.Conversation
	if err := s.db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		return nil, nil, err
	}

	// 保存用户消息
	userMessage := model.Message{
		ConversationID: conversationID,
		Role:           "user",
		Content:        req.Content,
	}
	if err := s.db.Create(&userMessage).Error; err != nil {
		return nil, nil, err
	}

	// 获取历史消息用于AI上下文
	var historyMessages []model.Message
	if err := s.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Limit(20).Find(&historyMessages).Error; err != nil {
		return nil, nil, err
	}

	// 转换为AI模型格式
	aiMessages := make([]*schema.Message, len(historyMessages))
	for i, msg := range historyMessages {
		var role schema.RoleType
		switch msg.Role {
		case "user":
			role = schema.User
		case "assistant":
			role = schema.Assistant
		case "system":
			role = schema.System
		default:
			role = schema.User
		}
		aiMessages[i] = &schema.Message{
			Role:    role,
			Content: msg.Content,
		}
	}

	// 获取AI回复
	aiResponse, err := s.aiService.GenerateResponse(ctx, aiMessages)
	if err != nil {
		return &userMessage, nil, err
	}

	// 保存AI回复
	assistantMessage := model.Message{
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        aiResponse,
	}
	if err := s.db.Create(&assistantMessage).Error; err != nil {
		return &userMessage, nil, err
	}

	// 更新会话的更新时间
	s.db.Model(&conversation).Update("updated_at", assistantMessage.CreatedAt)

	return &userMessage, &assistantMessage, nil
}

// StreamChat 流式聊天
func (s *ChatService) StreamChat(ctx context.Context, userID, conversationID uint, content string, callback func(string) error) (*model.Message, error) {
	// 验证会话是否属于用户
	var conversation model.Conversation
	if err := s.db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		return nil, err
	}

	// 保存用户消息
	userMessage := model.Message{
		ConversationID: conversationID,
		Role:           "user",
		Content:        content,
	}
	if err := s.db.Create(&userMessage).Error; err != nil {
		return nil, err
	}

	// 获取历史消息
	var historyMessages []model.Message
	if err := s.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Limit(20).Find(&historyMessages).Error; err != nil {
		return nil, err
	}

	// 转换为AI模型格式
	aiMessages := make([]*schema.Message, len(historyMessages))
	for i, msg := range historyMessages {
		var role schema.RoleType
		switch msg.Role {
		case "user":
			role = schema.User
		case "assistant":
			role = schema.Assistant
		case "system":
			role = schema.System
		default:
			role = schema.User
		}
		aiMessages[i] = &schema.Message{
			Role:    role,
			Content: msg.Content,
		}
	}

	// 流式获取AI回复
	respChan, errorChan := s.aiService.StreamResponse(ctx, aiMessages)
	var fullResponse string

	for {
		select {
		case chunk, ok := <-respChan:
			if !ok {
				// 通道关闭，流式响应结束
				goto StreamEnd
			}
			fullResponse += chunk
			if err := callback(chunk); err != nil {
				return &userMessage, err
			}
		case err := <-errorChan:
			if err != nil {
				return &userMessage, err
			}
		}
	}

StreamEnd:

	// 保存完整的AI回复
	assistantMessage := model.Message{
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        fullResponse,
	}
	if err := s.db.Create(&assistantMessage).Error; err != nil {
		return &userMessage, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// 更新会话的更新时间
	s.db.Model(&conversation).Update("updated_at", assistantMessage.CreatedAt)

	return &userMessage, nil
}