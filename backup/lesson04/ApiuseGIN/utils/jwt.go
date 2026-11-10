// utils/jwt.go 封装 JWT 的签发与解析，Token 有效期 7 天。
package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 是 JWT 的自定义载荷，嵌入 jwt.RegisteredClaims 以支持 exp/iat 等标准字段。
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 为指定用户签发 JWT，使用 HS256 算法和传入的 secret 签名。
// Token 中包含 user_id、username，有效期 7 天。
func GenerateToken(userID uint, username, secret string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析并验证 JWT 字符串，返回自定义 Claims。
// 签名不匹配、过期或格式错误时返回 error。
func ParseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
