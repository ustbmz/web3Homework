package main

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User 用户。Has Many Posts：一个用户可以发多篇文章。
//
// GORM 靠「外键字段名约定」建关联：Post.UserID 指向 User.ID。
// Posts 切片本身不占数据库列，只是查询时用来装关联结果。
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"size:64;uniqueIndex;not null"`
	Email    string `gorm:"size:128;uniqueIndex;not null"`
	Posts    []Post `gorm:"foreignKey:UserID"`
}

// Post 文章。Belongs To User，同时 Has Many Comments。
type Post struct {
	ID       uint      `gorm:"primaryKey"`
	Title    string    `gorm:"size:200;not null"`
	Content  string    `gorm:"type:text;not null"`
	UserID   uint      `gorm:"not null;index"` // 外键：属于哪个用户
	User     User      `gorm:"foreignKey:UserID"`
	Comments []Comment `gorm:"foreignKey:PostID"`
}

// Comment 评论。Belongs To Post（以及发表评论的 User）。
type Comment struct {
	ID      uint   `gorm:"primaryKey"`
	Content string `gorm:"type:text;not null"`
	PostID  uint   `gorm:"not null;index"` // 外键：属于哪篇文章
	UserID  uint   `gorm:"not null;index"` // 外键：谁写的评论
	Post    Post   `gorm:"foreignKey:PostID"`
	User    User   `gorm:"foreignKey:UserID"`
}

func main() {
	// 内存库即可演示建表；Logger Info 会把 CREATE TABLE 打到控制台。
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	// AutoMigrate：按结构体标签生成 / 对齐表结构（只增不删列）。
	// 顺序无硬性要求，GORM 会按外键依赖建表。
	if err := db.AutoMigrate(&User{}, &Post{}, &Comment{}); err != nil {
		log.Fatal(err)
	}

	// 从 sqlite_master 读出实际表名，确认三张表已落地。
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name").Scan(&tables)
	fmt.Println("已创建的表:", tables)

	// 插一条完整链路，验证外键能写进去。
	user := User{
		Username: "alice",
		Email:    "alice@example.com",
		Posts: []Post{
			{
				Title:   "第一篇文章",
				Content: "Hello GORM",
				Comments: []Comment{
					{Content: "写得真好", UserID: 1},
				},
			},
		},
	}
	if err := db.Create(&user).Error; err != nil {
		log.Fatal(err)
	}
	fmt.Printf("写入成功: user.id=%d  post.id=%d  comment 数=%d\n",
		user.ID, user.Posts[0].ID, len(user.Posts[0].Comments))
}
