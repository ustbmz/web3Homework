// config 包负责应用配置加载和数据库初始化，对外暴露全局 DB 实例供各 handler 使用。
package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"blog-api/models"
)

// DB 是全局 GORM 数据库连接，InitDB 成功后即可在各 handler 中直接使用。
var DB *gorm.DB

// AppConfig 保存从环境变量读取的运行时配置。
type AppConfig struct {
	Port      string // HTTP 服务监听端口，默认 3001
	JWTSecret string // JWT 签名密钥，用于签发和验证 Token
	DBPath    string // SQLite 数据库文件路径
}

// Load 从 .env 文件和系统环境变量中读取配置。
// 优先使用环境变量；若未设置则使用默认值（见 getEnv）。
func Load() AppConfig {
	_ = godotenv.Load()
	return AppConfig{
		Port:      getEnv("PORT", "3001"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-me"),
		DBPath:    getEnv("DB_PATH", "db/blog.db"),
	}
}

// InitDB 建立 SQLite 数据库连接并自动迁移表结构。
// 步骤：创建 db 目录 → 打开 SQLite → AutoMigrate 三张表 → 赋值给全局 DB。
func InitDB(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}); err != nil {
		return err
	}
	DB = db
	slog.Info("database connected", "path", dbPath)
	return nil
}

// getEnv 读取单个环境变量；若为空则返回 fallback 默认值。
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
