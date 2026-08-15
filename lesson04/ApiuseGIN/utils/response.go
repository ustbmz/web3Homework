// utils/response.go 定义统一的 JSON 响应格式，所有 API 接口共用。
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是 API 统一响应结构：code=0 表示成功，非 0 表示业务错误码。
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Success 返回成功响应，msg 固定为 "ok"，HTTP 状态码 200。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "ok", Data: data})
}

// SuccessMsg 返回成功响应，可自定义 msg 消息（如 "register success"）。
func SuccessMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: msg, Data: data})
}

// Fail 返回错误响应，httpStatus 为 HTTP 状态码，code 为业务错误码，msg 为错误描述。
func Fail(c *gin.Context, httpStatus, code int, msg string) {
	c.JSON(httpStatus, Response{Code: code, Msg: msg})
}

// AbortFail 与 Fail 类似，但会调用 c.Abort() 终止后续中间件和 handler 执行。
// 用于认证中间件等需要在失败时立即中断请求链路的场景。
func AbortFail(c *gin.Context, httpStatus, code int, msg string) {
	c.AbortWithStatusJSON(httpStatus, Response{Code: code, Msg: msg})
}
