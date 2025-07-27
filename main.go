package main

import (
	"context"
	"log"
	"time"

	"ai-chat-backend/internal/config"
	"ai-chat-backend/internal/database"
	"ai-chat-backend/internal/handler"
	"ai-chat-backend/internal/middleware"
	"ai-chat-backend/internal/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
	// 初始化配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.Init(cfg.Database.DSN)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 初始化AI服务
	aiService, err := service.NewAIService(cfg)
	if err != nil {
		log.Fatal("Failed to initialize AI service:", err)
	}

	// 初始化服务层
	userService := service.NewUserService(db)
	chatService := service.NewChatService(db, aiService)

	// 初始化处理器
	userHandler := handler.NewUserHandler(userService)
	chatHandler := handler.NewChatHandler(chatService)

	// 创建Hertz服务器
	h := server.Default(
		server.WithHostPorts(cfg.Server.Address),
		server.WithReadTimeout(30*time.Second),
		server.WithWriteTimeout(30*time.Second),
	)

	// 中间件
	h.Use(middleware.CORS())
	h.Use(middleware.Logger())

	// API路由
	api := h.Group("/api/v1")
	{
		// 用户相关路由
		user := api.Group("/user")
		{
			user.POST("/register", userHandler.Register)
			user.POST("/login", userHandler.Login)
			user.POST("/forgot-password", userHandler.ForgotPassword)
			user.POST("/reset-password", userHandler.ResetPassword)
		}

		// 流式聊天路由（不需要Auth中间件，因为EventSource不支持自定义headers）
		api.GET("/conversations/:id/stream", chatHandler.StreamChat)

		// 需要认证的路由
		auth := api.Group("/", middleware.Auth())
		{
			// 用户信息
			auth.GET("/user/profile", userHandler.GetProfile)
			auth.PUT("/user/profile", userHandler.UpdateProfile)
			auth.PUT("/user/password", userHandler.ChangePassword)

			// 聊天相关
			auth.GET("/conversations", chatHandler.GetConversations)
			auth.POST("/conversations", chatHandler.CreateConversation)
			auth.GET("/conversations/:id", chatHandler.GetConversation)
			auth.PUT("/conversations/:id", chatHandler.UpdateConversation)
			auth.DELETE("/conversations/:id", chatHandler.DeleteConversation)
			auth.GET("/conversations/:id/messages", chatHandler.GetMessages)
			auth.POST("/conversations/:id/messages", chatHandler.SendMessage)
		}
	}

	// 健康检查
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]string{"status": "ok"})
	})

	hlog.Info("Server starting on", cfg.Server.Address)
	h.Spin()
}