// handlers/comment.go 处理评论相关 HTTP 请求：按文章查询评论、发表评论。
package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blog-api/config"
	"blog-api/middleware"
	"blog-api/models"
	"blog-api/utils"
)

// CommentHandler 评论处理器，无额外状态。
type CommentHandler struct{}

// commentRequest 发表评论时的请求体结构。
type commentRequest struct {
	Content string `json:"content" binding:"required,min=1"`
}

// ListByPost 获取某篇文章下的所有评论 GET /api/posts/:id/comments（公开）
// 先确认文章存在，再按 post_id 查询评论，Preload 评论者信息，按时间升序排列。
func (h *CommentHandler) ListByPost(c *gin.Context) {
	postID, err := parseUintParam(c, "id")
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, 400, "invalid post id")
		return
	}
	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, 404, "post not found")
			return
		}
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	var comments []models.Comment
	if err := config.DB.Where("post_id = ?", postID).Preload("User").Order("created_at asc").Find(&comments).Error; err != nil {
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	utils.Success(c, comments)
}

// Create 发表评论 POST /api/posts/:id/comments（需 JWT 认证）
// 确认文章存在后，将当前用户 ID 和评论内容写入 comments 表。
func (h *CommentHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, 401, "unauthorized")
		return
	}
	postID, err := parseUintParam(c, "id")
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, 400, "invalid post id")
		return
	}
	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, 404, "post not found")
			return
		}
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	var req commentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	comment := models.Comment{Content: req.Content, UserID: userID, PostID: postID}
	if err := config.DB.Create(&comment).Error; err != nil {
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	config.DB.Preload("User").First(&comment, comment.ID)
	utils.SuccessMsg(c, "comment created", comment)
}
