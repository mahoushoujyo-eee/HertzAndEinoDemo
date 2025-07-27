package service

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"ai-chat-backend/internal/config"
)

type AIService struct {
	model *openai.ChatModel
}

func NewAIService(cfg *config.Config) (*AIService, error) {
	ctx := context.Background()
	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.AI.BaseURL,
		APIKey:  cfg.AI.APIKey,
		Timeout: cfg.AI.Timeout,
		Model:   cfg.AI.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI model: %w", err)
	}

	return &AIService{
		model: model,
	}, nil
}

// GenerateResponse 生成AI回复
func (s *AIService) GenerateResponse(ctx context.Context, messages []*schema.Message) (string, error) {
	resp, err := s.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	if resp == nil || resp.Content == "" {
		return "", fmt.Errorf("no response generated")
	}

	return resp.Content, nil
}

// StreamResponse 流式生成AI回复
func (s *AIService) StreamResponse(ctx context.Context, messages []*schema.Message) (<-chan string, <-chan error) {
	respChan := make(chan string, 10) // 减小缓冲区以确保实时性
	errorChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errorChan)

		log.Printf("Starting stream for %d messages", len(messages))
		stream, err := s.model.Stream(ctx, messages)
		if err != nil {
			log.Printf("Failed to create stream: %v", err)
			errorChan <- fmt.Errorf("failed to create stream: %w", err)
			return
		}
		defer stream.Close()

		log.Printf("Stream created successfully, starting to receive chunks")
		for {
			chunk, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					log.Printf("Stream ended normally")
				} else {
					log.Printf("Stream error: %v", err)
					errorChan <- err
				}
				break
			}

			if chunk != nil && chunk.Content != "" {
				log.Printf("Received chunk: %q", chunk.Content)
				select {
				case respChan <- chunk.Content:
					// 成功发送
				case <-ctx.Done():
					log.Printf("Context cancelled")
					return
				}
			}
		}
		log.Printf("Stream processing completed")
	}()

	return respChan, errorChan
}