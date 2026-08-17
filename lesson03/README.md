# Lesson 03

每个文件独立 `main`，必须指定文件运行：

```bash
go run models.go
go run query.go
go run hooks.go
```

内存 SQLite，跑完不落盘。

---

## models.go

- `User` / `Post` / `Comment`：一对多靠 `UserID`、`PostID` 外键；`Posts`、`Comments` 切片不占列
- `AutoMigrate`：按结构体建表
- `Create(&user)`：嵌套写入用户 → 文章 → 评论，打印三个 id

## query.go

- `seed`：alice 两篇（3 评 / 1 评），bob 一篇（5 评）
- `queryUserPostsWithComments`：`Preload("Posts.Comments")` 查出某用户全部文章和评论
- `queryMostCommentedPost`：`GROUP BY post_id ORDER BY cnt DESC LIMIT 1` 拿冠军 id，再 `Preload` 文章、作者、评论
- `main`：打印 alice 的文章评论，再打印评论最多的那篇（Bob 的热门文章，5 条）

## hooks.go

- `User.PostCount`、`Post.CommentStatus` 由钩子维护
- `Post.AfterCreate`：`post_count + 1`（用 `tx` + `UpdateColumn`）
- `Comment.AfterDelete`：该文剩余评论为 0 时把状态改成「无评论」；删除前必须 `First`，否则 `PostID` 是 0
- `main`：发两篇后 `PostCount=2`；两条评论删完后状态变为「无评论」
