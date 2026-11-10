package main

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 模型与 models.go 一致，本文件单独可运行。

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"size:64;uniqueIndex;not null"`
	Email    string `gorm:"size:128;uniqueIndex;not null"`
	Posts    []Post `gorm:"foreignKey:UserID"`
}

type Post struct {
	ID       uint      `gorm:"primaryKey"`
	Title    string    `gorm:"size:200;not null"`
	Content  string    `gorm:"type:text;not null"`
	UserID   uint      `gorm:"not null;index"`
	User     User      `gorm:"foreignKey:UserID"`
	Comments []Comment `gorm:"foreignKey:PostID"`
}

type Comment struct {
	ID      uint   `gorm:"primaryKey"`
	Content string `gorm:"type:text;not null"`
	PostID  uint   `gorm:"not null;index"`
	UserID  uint   `gorm:"not null;index"`
	Post    Post   `gorm:"foreignKey:PostID"`
	User    User   `gorm:"foreignKey:UserID"`
}

// seed 造出可对比的数据：alice 两篇（3 评 / 1 评），bob 一篇（5 评）。
func seed(db *gorm.DB) {
	alice := User{Username: "alice", Email: "alice@example.com"}
	bob := User{Username: "bob", Email: "bob@example.com"}
	if err := db.Create(&alice).Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Create(&bob).Error; err != nil {
		log.Fatal(err)
	}

	posts := []Post{
		{Title: "Alice 的 Go 笔记", Content: "指针与切片", UserID: alice.ID},
		{Title: "Alice 的 SQL 笔记", Content: "JOIN 与 GROUP BY", UserID: alice.ID},
		{Title: "Bob 的热门文章", Content: "评论最多的那篇", UserID: bob.ID},
	}
	if err := db.Create(&posts).Error; err != nil {
		log.Fatal(err)
	}

	comments := []Comment{
		{Content: "讲得很清楚", PostID: posts[0].ID, UserID: bob.ID},
		{Content: "求更新", PostID: posts[0].ID, UserID: bob.ID},
		{Content: "收藏了", PostID: posts[0].ID, UserID: alice.ID},
		{Content: "SQL 例子不错", PostID: posts[1].ID, UserID: bob.ID},
		{Content: "赞 1", PostID: posts[2].ID, UserID: alice.ID},
		{Content: "赞 2", PostID: posts[2].ID, UserID: alice.ID},
		{Content: "赞 3", PostID: posts[2].ID, UserID: alice.ID},
		{Content: "赞 4", PostID: posts[2].ID, UserID: alice.ID},
		{Content: "赞 5", PostID: posts[2].ID, UserID: alice.ID},
	}
	if err := db.Create(&comments).Error; err != nil {
		log.Fatal(err)
	}
}

// queryUserPostsWithComments 查出某用户的全部文章，并带上每篇文章的评论。
//
// 原理：
//   Preload("Posts") 会再发一条 WHERE user_id IN (...) 把文章装进 user.Posts。
//   Preload("Posts.Comments") 再按 post_id IN (...) 把评论装进每篇文章。
//   这是「预加载」，避免循环里对每篇文章再查一次评论（N+1）。
func queryUserPostsWithComments(db *gorm.DB, username string) (User, error) {
	var user User
	err := db.Preload("Posts.Comments").
		Where("username = ?", username).
		First(&user).Error
	return user, err
}

// queryMostCommentedPost 查出评论数最多的一篇文章。
//
// 原理：
//   先在 comments 上 GROUP BY post_id + ORDER BY COUNT(*)，得到「冠军」文章 id；
//   再按 id 把文章、作者、评论完整加载出来。
//   拆成两步是为了避开 SQLite 对 SELECT * + GROUP BY 的限制。
func queryMostCommentedPost(db *gorm.DB) (Post, int64, error) {
	type agg struct {
		PostID uint
		Cnt    int64
	}
	var a agg
	if err := db.Model(&Comment{}).
		Select("post_id AS post_id, COUNT(*) AS cnt").
		Group("post_id").
		Order("cnt DESC").
		Limit(1).
		Scan(&a).Error; err != nil {
		return Post{}, 0, err
	}

	var post Post
	err := db.Preload("User").Preload("Comments").First(&post, a.PostID).Error
	return post, a.Cnt, err
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
	seed(db)

	// ---------- 题目 2.1：某用户的全部文章 + 评论 ----------
	user, err := queryUserPostsWithComments(db, "alice")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("用户 %s 共 %d 篇文章\n", user.Username, len(user.Posts))
	for _, p := range user.Posts {
		fmt.Printf("  《%s》 评论 %d 条\n", p.Title, len(p.Comments))
		for _, c := range p.Comments {
			fmt.Printf("    - %s\n", c.Content)
		}
	}

	// ---------- 题目 2.2：评论数最多的文章 ----------
	post, cnt, err := queryMostCommentedPost(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n评论最多：《%s》（作者 %s，%d 条评论）\n",
		post.Title, post.User.Username, cnt)
}
