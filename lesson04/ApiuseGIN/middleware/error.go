// middleware/error.go 提供全局错误处理和请求日志中间件。
package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"blog-api/utils"
)

// ErrorHandler 全局错误捕获中间件。
// 在 c.Next() 之后检查 c.Errors，若有未处理的错误且响应尚未写入，则返回 500 JSON。
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		slog.Error("request error", "path", c.Request.URL.Path, "error", err)
		if !c.Writer.Written() {
			utils.Fail(c, http.StatusInternalServerError, 500, err.Error())
		}
	}
}

// Logger 请求日志中间件。
// 每个请求处理完成后记录 method、path、HTTP status，便于排查问题。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		slog.Info("request", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status())
	}
}
