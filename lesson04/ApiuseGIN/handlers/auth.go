// handlers/auth.go 处理用户认证相关 HTTP 请求：注册、登录、获取当前用户信息。
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

// AuthHandler 认证处理器，持有 JWT 密钥用于登录时签发 Token。
type AuthHandler struct {
	JWTSecret string
}

// registerRequest 注册接口的请求体结构，binding 标签用于 Gin 参数校验。
type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Email    string `json:"email" binding:"required,email"`
}

// loginRequest 登录接口的请求体结构。
type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 用户注册接口 POST /api/register
// 流程：校验 JSON 参数 → bcrypt 加密密码 → 写入 users 表 → 返回新用户信息（不含密码）。
// 若用户名或邮箱已存在，返回 409 Conflict。
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	user := models.User{Username: req.Username, Password: hash, Email: req.Email}
	if err := config.DB.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			utils.Fail(c, http.StatusConflict, 409, "username or email already exists")
			return
		}
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	utils.SuccessMsg(c, "register success", gin.H{"id": user.ID, "username": user.Username, "email": user.Email})
}

// Login 用户登录接口 POST /api/login
// 流程：校验参数 → 按 username 查用户 → bcrypt 比对密码 → 签发 JWT → 返回 token 和用户信息。
// 用户名不存在或密码错误均返回 401，避免泄露具体原因。
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	var user models.User
	if err := config.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusUnauthorized, 401, "invalid username or password")
			return
		}
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	if !utils.CheckPassword(user.Password, req.Password) {
		utils.Fail(c, http.StatusUnauthorized, 401, "invalid username or password")
		return
	}
	token, err := utils.GenerateToken(user.ID, user.Username, h.JWTSecret)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, 500, "internal server error")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0, "msg": "login success", "token": token,
		"data": gin.H{"id": user.ID, "username": user.Username, "email": user.Email},
	})
}

// Me 获取当前登录用户信息 GET /api/me（需 JWT 认证）
// 从中间件写入的 Context 中读取 userID，再查库返回用户基本信息。
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, 401, "unauthorized")
		return
	}
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, 404, "user not found")
		return
	}
	utils.Success(c, gin.H{"id": user.ID, "username": user.Username, "email": user.Email})
}
