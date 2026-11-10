// handlers/post.go 处理文章 CRUD 相关 HTTP 请求。
package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blog-api/config"
	"blog-api/middleware"
	"blog-api/models"
	"blog-api/utils"
)

// PostHandler 文章处理器，无额外状态，直接使用全局 config.DB。
type PostHandler struct{}

// postRequest 创建/更新文章时的请求体结构。
type postRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,min=1"`
}

// List 获取文章列表 GET /api/posts（公开，无需认证）
// 按 created_at 降序排列，Preload 作者信息一并返回。
func (h *PostHandler) List(c *gin.Context) {
	var posts []models.Post
	if err := config.DB.Preload("User").Order("created_at desc").Find(&posts).Error; err != nil {
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	utils.Success(c, posts)
}

// Get 获取单篇文章 GET /api/posts/:id（公开）
// 解析 URL 中的 id 参数，查库并 Preload 作者；不存在返回 404。
func (h *PostHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, 400, "invalid post id")
		return
	}
	var post models.Post
	if err := config.DB.Preload("User").First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, 404, "post not found")
			return
		}
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	utils.Success(c, post)
}

// Create 创建文章 POST /api/posts（需 JWT 认证）
// 从 Context 取当前用户 ID 作为 author，写入 posts 表后返回完整文章（含作者）。
func (h *PostHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, 401, "unauthorized")
		return
	}
	var req postRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	post := models.Post{Title: req.Title, Content: req.Content, UserID: userID}
	if err := config.DB.Create(&post).Error; err != nil {
		slog.Error("create post failed", "error", err)
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	config.DB.Preload("User").First(&post, post.ID)
	utils.SuccessMsg(c, "post created", post)
}

// Update 更新文章 PUT /api/posts/:id（需 JWT，仅作者可操作）
// 先查文章是否存在，再校验 post.UserID == 当前用户 ID，通过后更新 title/content。
func (h *PostHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, 401, "unauthorized")
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, 400, "invalid post id")
		return
	}
	var post models.Post
	if err := config.DB.First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, 404, "post not found")
			return
		}
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	if post.UserID != userID {
		utils.Fail(c, http.StatusForbidden, 403, "only author can update this post")
		return
	}
	var req postRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	post.Title, post.Content = req.Title, req.Content
	if err := config.DB.Save(&post).Error; err != nil {
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	config.DB.Preload("User").First(&post, post.ID)
	utils.SuccessMsg(c, "post updated", post)
}

// Delete 删除文章 DELETE /api/posts/:id（需 JWT，仅作者可操作）
// 使用事务：先删除该文章下所有评论，再删除文章本身，保证数据一致性。
func (h *PostHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, 401, "unauthorized")
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, 400, "invalid post id")
		return
	}
	var post models.Post
	if err := config.DB.First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, 404, "post not found")
			return
		}
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	if post.UserID != userID {
		utils.Fail(c, http.StatusForbidden, 403, "only author can delete this post")
		return
	}
	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", post.ID).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&post).Error
	}); err != nil {
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	utils.SuccessMsg(c, "post deleted", nil)
}

// parseUintParam 从 URL 路径参数（如 :id）解析无符号整数。
// 供 post 和 comment handler 共用，解析失败时返回 error。
func parseUintParam(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
