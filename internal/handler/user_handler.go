package handler

import (
	"context"

	"ai-chat-backend/internal/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService *service.UserService
	validator   *validator.Validate
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator.New(),
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Register 用户注册
func (h *UserHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req service.RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.userService.Register(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(consts.StatusCreated, SuccessResponse{
		Message: "User registered successfully",
		Data:    resp,
	})
}

// Login 用户登录
func (h *UserHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req service.LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.userService.Login(&req)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(consts.StatusOK, SuccessResponse{
		Message: "Login successful",
		Data:    resp,
	})
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(ctx context.Context, c *app.RequestContext) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(consts.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	c.JSON(consts.StatusOK, SuccessResponse{
		Message: "Profile retrieved successfully",
		Data:    user,
	})
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(ctx context.Context, c *app.RequestContext) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}

	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.userService.UpdateProfile(userID.(uint), req.Nickname, req.Avatar)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(consts.StatusOK, SuccessResponse{
		Message: "Profile updated successfully",
	})
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(ctx context.Context, c *app.RequestContext) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(consts.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=6"`
	}

	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.userService.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(consts.StatusOK, SuccessResponse{
		Message: "Password changed successfully",
	})
}

// ForgotPassword 忘记密码（暂时返回成功，实际需要邮件服务）
func (h *UserHandler) ForgotPassword(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// TODO: 实现邮件发送逻辑
	c.JSON(consts.StatusOK, SuccessResponse{
		Message: "Password reset email sent",
	})
}

// ResetPassword 重置密码（暂时返回成功，实际需要验证码）
func (h *UserHandler) ResetPassword(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Email       string `json:"email" validate:"required,email"`
		Code        string `json:"code" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=6"`
	}

	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// TODO: 实现验证码验证和密码重置逻辑
	c.JSON(consts.StatusOK, SuccessResponse{
		Message: "Password reset successfully",
	})
}