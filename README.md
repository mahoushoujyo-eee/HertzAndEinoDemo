# AI Chat Backend

一个基于 Go 语言开发的 AI 聊天应用后端服务，使用 Hertz 框架构建，支持用户管理、实时聊天和 AI 对话功能。

## 🚀 功能特性

- **用户管理**：用户注册、登录、密码重置、个人资料管理
- **JWT 认证**：基于 JWT 的用户身份验证和授权
- **AI 聊天**：集成 OpenAI API，支持流式对话
- **会话管理**：创建、查看、更新和删除聊天会话
- **消息历史**：完整的聊天记录存储和检索
- **CORS 支持**：跨域资源共享配置
- **健康检查**：服务状态监控端点

## 🛠 技术栈

- **框架**：[CloudWeGo Hertz](https://github.com/cloudwego/hertz) - 高性能 HTTP 框架
- **数据库**：MySQL + [GORM](https://gorm.io/) ORM
- **AI 服务**：[CloudWeGo Eino](https://github.com/cloudwego/eino) + OpenAI API
- **认证**：JWT (JSON Web Tokens)
- **密码加密**：bcrypt
- **参数验证**：go-playground/validator

## 📁 项目结构

```
backend/
├── main.go                 # 应用入口
├── go.mod                  # Go 模块依赖
├── go.sum                  # 依赖校验文件
├── .gitignore             # Git 忽略文件
└── internal/              # 内部包
    ├── config/            # 配置管理
    │   └── config.go
    ├── database/          # 数据库连接
    │   └── database.go
    ├── handler/           # HTTP 处理器
    │   ├── chat_handler.go
    │   └── user_handler.go
    ├── middleware/        # 中间件
    │   └── middleware.go
    ├── model/            # 数据模型
    │   └── user.go
    ├── service/          # 业务逻辑层
    │   ├── ai_service.go
    │   ├── chat_service.go
    │   └── user_service.go
    └── utils/            # 工具函数
        ├── jwt.go
        └── password.go
```

## 🚦 快速开始

### 环境要求

- Go 1.23.0+
- MySQL 5.7+
- OpenAI API Key (或兼容的 API 服务)

### 安装依赖

```bash
go mod download
```

### 环境变量配置

创建 `.env` 文件或设置以下环境变量：

```bash
# 服务器配置
SERVER_ADDRESS=:8080

# 数据库配置
DATABASE_DSN=username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local

# AI 服务配置
AI_BASE_URL=https://openai.qiniu.com/v1
AI_API_KEY=your-api-key
AI_MODEL=deepseek-v3-0324

# JWT 配置
JWT_SECRET=your-secret-key-change-in-production
```

### 运行应用

```bash
# 开发模式
go run main.go

# 编译运行
go build -o ai-chat-backend
./ai-chat-backend
```

服务将在 `http://localhost:8080` 启动。

## 📚 API 文档

### 用户相关 API

#### 用户注册
```http
POST /api/v1/user/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "nickname": "用户昵称"
}
```

#### 用户登录
```http
POST /api/v1/user/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

#### 忘记密码
```http
POST /api/v1/user/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

#### 重置密码
```http
POST /api/v1/user/reset-password
Content-Type: application/json

{
  "token": "reset-token",
  "password": "newpassword123"
}
```

### 认证相关 API (需要 Authorization Header)

#### 获取用户信息
```http
GET /api/v1/user/profile
Authorization: Bearer <jwt-token>
```

#### 更新用户信息
```http
PUT /api/v1/user/profile
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "nickname": "新昵称",
  "avatar": "头像URL"
}
```

#### 修改密码
```http
PUT /api/v1/user/password
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "old_password": "oldpassword",
  "new_password": "newpassword123"
}
```

### 聊天相关 API

#### 获取会话列表
```http
GET /api/v1/conversations
Authorization: Bearer <jwt-token>
```

#### 创建新会话
```http
POST /api/v1/conversations
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "title": "会话标题"
}
```

#### 获取会话详情
```http
GET /api/v1/conversations/{id}
Authorization: Bearer <jwt-token>
```

#### 获取会话消息
```http
GET /api/v1/conversations/{id}/messages
Authorization: Bearer <jwt-token>
```

#### 发送消息
```http
POST /api/v1/conversations/{id}/messages
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "content": "用户消息内容"
}
```

#### 流式聊天 (Server-Sent Events)
```http
GET /api/v1/conversations/{id}/stream?token=<jwt-token>&message=<message>
```

### 健康检查
```http
GET /health
```

## 🗄️ 数据库模型

### User (用户表)
- `id`: 主键
- `email`: 邮箱 (唯一)
- `password`: 加密密码
- `nickname`: 昵称
- `avatar`: 头像URL
- `is_active`: 是否激活
- `created_at`: 创建时间
- `updated_at`: 更新时间

### Conversation (会话表)
- `id`: 主键
- `user_id`: 用户ID (外键)
- `title`: 会话标题
- `created_at`: 创建时间
- `updated_at`: 更新时间

### Message (消息表)
- `id`: 主键
- `conversation_id`: 会话ID (外键)
- `role`: 角色 (user/assistant)
- `content`: 消息内容
- `created_at`: 创建时间
- `updated_at`: 更新时间

## 🔧 配置说明

应用支持通过环境变量进行配置，如果未设置环境变量，将使用默认值：

- `SERVER_ADDRESS`: 服务器监听地址 (默认: `:8080`)
- `DATABASE_DSN`: MySQL 数据库连接字符串
- `AI_BASE_URL`: AI 服务基础URL (默认: `https://openai.qiniu.com/v1`)
- `AI_API_KEY`: AI 服务 API 密钥
- `AI_MODEL`: AI 模型名称 (默认: `deepseek-v3-0324`)
- `JWT_SECRET`: JWT 签名密钥 (生产环境必须修改)

## 🛡️ 安全特性

- **密码加密**：使用 bcrypt 算法加密存储用户密码
- **JWT 认证**：基于 JWT 的无状态身份验证
- **CORS 配置**：支持跨域请求配置
- **参数验证**：严格的输入参数验证
- **软删除**：数据库记录软删除，保护数据安全

## 📝 开发说明

### 添加新的 API 端点

1. 在 `internal/handler/` 中添加处理器函数
2. 在 `internal/service/` 中添加业务逻辑
3. 在 `main.go` 中注册路由
4. 更新 API 文档

### 数据库迁移

应用启动时会自动执行数据库迁移，创建或更新表结构。如需手动控制迁移，可以修改 `internal/database/database.go` 文件。

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关链接

- [CloudWeGo Hertz](https://github.com/cloudwego/hertz)
- [CloudWeGo Eino](https://github.com/cloudwego/eino)
- [GORM](https://gorm.io/)
- [JWT](https://jwt.io/)