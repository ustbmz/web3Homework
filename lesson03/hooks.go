package main

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User 增加 PostCount，给 Post.AfterCreate 回写。
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"size:64;uniqueIndex;not null"`
	Email     string `gorm:"size:128;uniqueIndex;not null"`
	PostCount int64  `gorm:"not null;default:0"` // 文章数统计，由钩子维护
	Posts     []Post `gorm:"foreignKey:UserID"`
}

// Post 增加 CommentStatus，给 Comment.AfterDelete 回写。
type Post struct {
	ID            uint      `gorm:"primaryKey"`
	Title         string    `gorm:"size:200;not null"`
	Content       string    `gorm:"type:text;not null"`
	UserID        uint      `gorm:"not null;index"`
	CommentStatus string    `gorm:"size:16;not null;default:无评论"`
	User          User      `gorm:"foreignKey:UserID"`
	Comments      []Comment `gorm:"foreignKey:PostID"`
}

type Comment struct {
	ID      uint   `gorm:"primaryKey"`
	Content string `gorm:"type:text;not null"`
	PostID  uint   `gorm:"not null;index"`
	UserID  uint   `gorm:"not null;index"`
	Post    Post   `gorm:"foreignKey:PostID"`
	User    User   `gorm:"foreignKey:UserID"`
}

// AfterCreate 在 INSERT posts 成功之后执行，仍处在同一次事务 tx 里。
//
// 必须用 tx 而不是外面的 db：创建失败回滚时，这里的 +1 也会一起回滚。
// UpdateColumn 只改一列且不触发 User 的钩子，避免钩子互相调用。
func (p *Post) AfterCreate(tx *gorm.DB) error {
	return tx.Model(&User{}).
		Where("id = ?", p.UserID).
		UpdateColumn("post_count", gorm.Expr("post_count + ?", 1)).Error
}

// AfterDelete 在 DELETE comments 成功之后执行。
//
// 注意：钩子里要用 c.PostID，所以删除前必须先把记录查出来，
// 只写 db.Delete(&Comment{}, id) 时结构体往往只有 ID，PostID 是 0。
func (c *Comment) AfterDelete(tx *gorm.DB) error {
	var remain int64
	if err := tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&remain).Error; err != nil {
		return err
	}
	if remain == 0 {
		return tx.Model(&Post{}).
			Where("id = ?", c.PostID).
			Update("comment_status", "无评论").Error
	}
	return nil
}

func printUser(db *gorm.DB, id uint, tag string) {
	var u User
	db.First(&u, id)
	fmt.Printf("[%s] 用户 %s 的 PostCount = %d\n", tag, u.Username, u.PostCount)
}

func printPost(db *gorm.DB, id uint, tag string) {
	var p Post
	db.First(&p, id)
	var n int64
	db.Model(&Comment{}).Where("post_id = ?", id).Count(&n)
	fmt.Printf("[%s] 《%s》 剩余评论 %d，状态 = %s\n", tag, p.Title, n, p.CommentStatus)
}

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&User{}, &Post{}, &Comment{}); err != nil {
		log.Fatal(err)
	}

	user := User{Username: "alice", Email: "alice@example.com"}
	if err := db.Create(&user).Error; err != nil {
		log.Fatal(err)
	}
	printUser(db, user.ID, "建用户后")

	// ---------- 题目 3.1：创建文章时自动累加用户文章数 ----------
	p1 := Post{Title: "钩子与事务", Content: "AfterCreate", UserID: user.ID}
	p2 := Post{Title: "删除评论", Content: "AfterDelete", UserID: user.ID}
	if err := db.Create(&p1).Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Create(&p2).Error; err != nil {
		log.Fatal(err)
	}
	printUser(db, user.ID, "发两篇文章后") // 期望 PostCount = 2

	// 先有评论，并把状态标成「有评论」（删除钩子只负责改回「无评论」）。
	cs := []Comment{
		{Content: "第一条", PostID: p1.ID, UserID: user.ID},
		{Content: "第二条", PostID: p1.ID, UserID: user.ID},
	}
	if err := db.Create(&cs).Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&p1).Update("comment_status", "有评论").Error; err != nil {
		log.Fatal(err)
	}
	printPost(db, p1.ID, "两条评论")

	// ---------- 题目 3.2：删评论，最后一条删完后状态变为「无评论」----------
	// 必须 First 再 Delete，AfterDelete 才能拿到 PostID。
	var first Comment
	if err := db.First(&first, cs[0].ID).Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Delete(&first).Error; err != nil {
		log.Fatal(err)
	}
	printPost(db, p1.ID, "删掉 1 条") // 还剩 1 条，状态仍是「有评论」

	var second Comment
	if err := db.First(&second, cs[1].ID).Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Delete(&second).Error; err != nil {
		log.Fatal(err)
	}
	printPost(db, p1.ID, "全部删完") // 期望状态 = 无评论
}
