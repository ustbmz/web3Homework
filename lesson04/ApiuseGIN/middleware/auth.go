// middleware/auth.go 提供 JWT 认证中间件，解析 Token 并将用户信息写入 Gin Context。
package middleware

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	"blog-api/utils"
)

// ContextUserIDKey 是 Gin Context 中存储当前用户 ID 的键名。
const ContextUserIDKey = "userID"

// ContextUsernameKey 是 Gin Context 中存储当前用户名的键名。
const ContextUsernameKey = "username"

// Auth 返回 JWT 认证中间件，保护需要登录的路由组。
// 流程：读取 Authorization 头 → 解析 "Bearer <token>" → 验证 JWT → 将 userID/username 写入 Context → c.Next()。
// 任一环节失败则 Abort 并返回 401，不再执行后续 handler。
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.AbortFail(c, 401, 401, "missing authorization header")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.AbortFail(c, 401, 401, "invalid authorization format")
			return
		}
		claims, err := utils.ParseToken(parts[1], jwtSecret)
		if err != nil {
			slog.Warn("jwt parse failed", "error", err)
			utils.AbortFail(c, 401, 401, "invalid or expired token")
			return
		}
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Next()
	}
}

// GetUserID 从 Gin Context 中安全读取当前登录用户的 ID。
// 供各 handler 使用，第二个返回值 ok 为 false 表示未认证或类型断言失败。
func GetUserID(c *gin.Context) (uint, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}
