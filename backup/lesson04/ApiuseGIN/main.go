// main 包是程序入口，负责初始化配置、数据库、路由，并启动 HTTP 服务。
package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"blog-api/config"
	"blog-api/handlers"
	"blog-api/middleware"
)

// main 启动博客 API 服务，整体流程：
//  1. 配置结构化日志
//  2. 加载 .env 配置并连接 SQLite 数据库
//  3. 创建 Gin 引擎，挂载全局中间件（Recovery / 日志 / 错误处理）
//  4. 注册公开路由（注册、登录、文章列表/详情、评论列表）
//  5. 注册需 JWT 认证的路由组（当前用户、文章增删改、发表评论）
//  6. 监听端口，阻塞运行直到进程退出
func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// 加载 PORT / JWT_SECRET / DB_PATH 等配置
	cfg := config.Load()
	if err := config.InitDB(cfg.DBPath); err != nil {
		slog.Error("database init failed", "error", err)
		os.Exit(1)
	}

	// gin.New() 不含默认 Logger，便于自定义中间件
	r := gin.New()
	r.Use(gin.Recovery())              // panic 时恢复，避免进程崩溃
	r.Use(middleware.Logger())         // 记录每个请求的 method/path/status
	r.Use(middleware.ErrorHandler())   // 捕获 c.Errors 中的未处理错误

	authHandler := &handlers.AuthHandler{JWTSecret: cfg.JWTSecret}
	postHandler := &handlers.PostHandler{}
	commentHandler := &handlers.CommentHandler{}

	api := r.Group("/api")
	{
		// --- 公开路由（无需 Token）---
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.GET("/posts", postHandler.List)
		api.GET("/posts/:id", postHandler.Get)
		api.GET("/posts/:id/comments", commentHandler.ListByPost)

		// --- 需 JWT 认证的路由组 ---
		authed := api.Group("/")
		authed.Use(middleware.Auth(cfg.JWTSecret))
		{
			authed.GET("/me", authHandler.Me)
			authed.POST("/posts", postHandler.Create)
			authed.PUT("/posts/:id", postHandler.Update)
			authed.DELETE("/posts/:id", postHandler.Delete)
			authed.POST("/posts/:id/comments", commentHandler.Create)
		}
	}

	slog.Info("server starting", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
