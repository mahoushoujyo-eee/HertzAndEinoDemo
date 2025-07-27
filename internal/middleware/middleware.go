package middleware

import (
	"context"
	"strings"
	"time"

	"ai-chat-backend/internal/config"
	"ai-chat-backend/internal/utils"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// CORS 中间件
func CORS() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}

		c.Next(ctx)
	}
}

// Logger 中间件
func Logger() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Path())
		method := string(c.Method())

		c.Next(ctx)

		latency := time.Since(start)
		status := c.Response.StatusCode()

		hlog.Infof("%s %s %d %v", method, path, status, latency)
	}
}

// Auth 认证中间件
func Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := c.GetHeader("Authorization")
		if len(token) == 0 {
			c.JSON(consts.StatusUnauthorized, map[string]string{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		tokenString := strings.TrimPrefix(string(token), "Bearer ")
		if tokenString == string(token) {
			c.JSON(consts.StatusUnauthorized, map[string]string{
				"error": "Invalid token format",
			})
			c.Abort()
			return
		}

		// 验证JWT token
		cfg := config.Load()
		claims, err := utils.ValidateJWT(tokenString, cfg.JWT.Secret)
		if err != nil {
			c.JSON(consts.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// 将用户ID存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Next(ctx)
	}
}