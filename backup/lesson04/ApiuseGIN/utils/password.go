// utils/password.go 封装 bcrypt 密码加密与校验，注册时哈希存储，登录时比对。
package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword 使用 bcrypt 对明文密码进行单向哈希，返回可存入数据库的字符串。
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 将明文密码与数据库中的 bcrypt 哈希比对，匹配返回 true。
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
