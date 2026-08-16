# Lesson 03：GORM 模型、关联查询与钩子

本目录对应三道进阶 GORM 题。每个文件独立带 `main`，代码里有逐行注释；本文只讲**实现原理**。

```
lesson03/
├── README.md
├── go.mod / go.sum
├── models.go    # 题目 1：User / Post / Comment 一对多，AutoMigrate 建表
├── query.go     # 题目 2：预加载用户文章+评论；评论数最多的文章
└── hooks.go     # 题目 3：Post.AfterCreate、Comment.AfterDelete
```

运行（必须指定单个文件，每个文件都有自己的 `main`）：

```bash
cd lesson03
go run models.go
go run query.go
go run hooks.go
```

依赖：`gorm.io/gorm` + `github.com/glebarez/sqlite`（纯 Go SQLite，不需要 CGO）。演示用内存库，跑完即消失，不落盘。

---

## 1. 模型定义：关联是「外键 + 切片」，不是 JOIN 语句

### 1.1 三张表、两段一对多

```
User 1 ──────< Post 1 ──────< Comment
     UserID           PostID
```

| 关系 | 在「一」这边 | 在「多」这边 |
| ---- | ------------ | ------------ |
| 用户 → 文章 | `User.Posts []Post` | `Post.UserID` + `Post.User` |
| 文章 → 评论 | `Post.Comments []Comment` | `Comment.PostID` + `Comment.Post` |

- `UserID` / `PostID` 是**真正落库的外键列**。
- `Posts` / `Comments` / `User` 这些结构体字段**不占列**，只在查询时由 GORM 填值。
- `gorm:"foreignKey:UserID"` 写明「用哪一列去对 User.ID」。不写也能靠约定猜出来，写上更不容易看错。

评论还带了 `UserID`：评论作者和文章作者可以不是同一个人。题目没强制，但博客里这是自然字段。

### 1.2 为什么是「值嵌套」而不是指针

`User User` 和 `*User` 都能用。值类型零值是空结构体，JSON / 打印更省事；指针才能区分「没预加载」和「加载了但是空」。作业里用值类型即可。

### 1.3 AutoMigrate 做了什么

`db.AutoMigrate(&User{}, &Post{}, &Comment{})` 会：

1. 按结构体生成 `CREATE TABLE`（没有表时）。
2. 缺列就 `ALTER TABLE ADD`。
3. **不会**删列、改列类型、删表。所以它是「对齐到结构体」，不是完整迁移工具。

GORM 默认表名是结构体的蛇形复数：`User` → `users`。主键默认 `id`。`uniqueIndex`、`index` 会建成索引，一对多查询 `WHERE user_id = ?` 才快。

`models.go` 把 Logger 开成 Info，控制台能直接看到建表 SQL。

### 1.4 嵌套 Create

```go
db.Create(&User{Posts: []Post{{Comments: []Comment{{...}}}}})
```

GORM 会按外键顺序先插 `users`，拿到自增 ID，写进 `posts.user_id`，再插 `comments.post_id`。这是关联写入，不是三次手写 Insert。

---

## 2. 关联查询：Preload 解决 N+1，聚合用 GROUP BY

### 2.1 查某用户的全部文章和评论

错误写法（N+1）：

```
SELECT * FROM users WHERE username = 'alice'          -- 1 次
SELECT * FROM posts WHERE user_id = 1                 -- 1 次
SELECT * FROM comments WHERE post_id = 1              -- 每篇文章再 1 次
SELECT * FROM comments WHERE post_id = 2
```

文章一多，评论查询次数就线性涨。

`Preload("Posts.Comments")` 的点号表示**嵌套预加载**：

1. `SELECT * FROM users WHERE username = 'alice'`
2. `SELECT * FROM posts WHERE user_id IN (1)`
3. `SELECT * FROM comments WHERE post_id IN (1, 2)`

三次 SQL，把评论按 `PostID` 填回对应的 `Post.Comments`。这是 GORM 的「预加载」，本质是 IN 查询 + 内存里拼图，不是一条大 JOIN。

`Joins("JOIN posts ...")` 会把多行拼进同一张结果表，一篇文章三条评论就会把 User 打成三行，还要自己去重。要「对象图」时用 Preload；要「扁平报表」时用 Joins。

### 2.2 评论数最多的文章

目标：`GROUP BY post_id ORDER BY COUNT(*) DESC LIMIT 1`。

SQLite 对 `SELECT posts.* ... GROUP BY posts.id` 比较挑剔（非聚合列必须在 GROUP BY 里）。所以拆两步：

1. 只在 `comments` 上聚合，得到冠军 `post_id` 和 `cnt`。
2. 再 `First(&post, id)` + `Preload` 把文章、作者、评论装回来。

并列第一时 `Limit(1)` 只取一行，作业不要求打破平局。

---

## 3. 钩子：插在「写库成功之后」，并且必须走同一个 tx

GORM 模型钩子是约定方法名，不用注册：

| 方法 | 何时调用 |
| ---- | -------- |
| `BeforeCreate` / `AfterCreate` | `Create` 前后 |
| `BeforeDelete` / `AfterDelete` | `Delete` 前后 |
| `BeforeUpdate` / `AfterUpdate` | `Save` / `Updates` 前后 |

签名必须是 `func (m *T) AfterXxx(tx *gorm.DB) error`。返回 error 会让整次操作失败并回滚。

### 3.1 `Post.AfterCreate`：给用户文章数 +1

```
Create(post)
  → INSERT posts
  → AfterCreate(tx)  →  UPDATE users SET post_count = post_count + 1
```

要点：

- **用 `tx`，不要用全局 `db`。** `Create` 若在事务里失败，钩子里那次 UPDATE 必须一起回滚，否则计数会漂。
- **用 `UpdateColumn` + `gorm.Expr`。** `UpdateColumn` 不触发 User 的钩子，避免「改 User → 又触发钩子」循环；`Expr` 是数据库侧自增，不会先读后写丢更新。
- 钩子里看到的 `p.ID`、`p.UserID` 已经是插入后的值。

题目只要求「创建文章时更新用户文章数」。删文章时没有对称的 `-1`，作业范围外不必补。

### 3.2 `Comment.AfterDelete`：没评论了就把状态改成「无评论」

```
Delete(comment)
  → DELETE comments WHERE id = ?
  → AfterDelete(tx)
       COUNT(comments WHERE post_id = ?)
       若为 0 → UPDATE posts SET comment_status = '无评论'
```

要点：

- 删除**已经发生**，`Count` 数的是剩余评论。最后一条删掉后就是 0。
- **必须先 `First` 再 `Delete(&comment)`。** `db.Delete(&Comment{}, id)` 传进去的结构体常常只有 ID，`c.PostID == 0`，钩子会去数 `post_id = 0`，状态永远改不对。
- 题目只要求「删到 0 条时改成无评论」。建评论时把状态改成「有评论」是演示数据，不是钩子职责。

`CommentStatus` 默认值在标签里写成 `default:无评论`，新文章一落地就是这个状态。

### 3.3 钩子不适合做什么

- 不适合发 HTTP、发邮件：失败了整笔 Create 会回滚，副作用却收不回来。
- 不适合很重的统计：高并发下「读 Count 再 Update」仍可能竞态；本题演示用 SQLite 单连接，够用。真正的计数更常见的是 `UpdateColumn` 加减，或单独的计数表。

---

## 4. 三道题怎么串起来

```
models.go   把「一对多」落成外键和表
query.go    用 Preload / GROUP BY 把关联读回来
hooks.go    在写路径上维护冗余字段（文章数、评论状态）
```

冗余字段（`PostCount`、`CommentStatus`）是用空间换查询：列表页不用每次 `COUNT`。代价是必须靠钩子（或应用层）和真实数据保持一致，这就是题目 3 存在的原因。
