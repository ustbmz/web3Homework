# 个人博客系统后端 (Gin + GORM + SQLite)

基于 Go 语言开发的个人博客 REST API，使用 Gin 框架、GORM ORM 和 SQLite 数据库，实现用户认证、文章 CRUD 和评论功能。

## 技术栈


| 组件     | 库                                       |
| ------ | --------------------------------------- |
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) |
| ORM    | [GORM](https://gorm.io)                 |
| 数据库    | SQLite (`gorm.io/driver/sqlite`)        |
| 密码加密   | `golang.org/x/crypto/bcrypt`            |
| JWT    | `github.com/golang-jwt/jwt/v5`          |
| 配置     | `github.com/joho/godotenv`              |
| 日志     | Go 标准库 `log/slog`                       |




## 项目结构

```
ApiuseGIN/
├── main.go              # 入口，路由注册
├── go.mod / go.sum      # 依赖管理
├── .env.example         # 环境变量示例
├── config/
│   └── config.go        # 配置加载、数据库连接、AutoMigrate
├── models/
│   └── models.go        # User / Post / Comment 模型
├── handlers/
│   ├── auth.go          # 注册、登录、当前用户
│   ├── post.go          # 文章 CRUD
│   └── comment.go       # 评论增查
├── middleware/
│   ├── auth.go          # JWT 认证中间件
│   └── error.go         # 错误处理、请求日志
├── utils/
│   ├── jwt.go           # JWT 签发与解析
│   ├── password.go      # bcrypt 密码加密
│   └── response.go      # 统一 JSON 响应格式
└── db/
    └── blog.db          # SQLite 数据库（运行后自动生成）
```



## 数据库设计



### users 表


| 字段         | 类型       | 说明            |
| ---------- | -------- | ------------- |
| id         | uint     | 主键            |
| username   | string   | 用户名，唯一        |
| password   | string   | bcrypt 加密后的密码 |
| email      | string   | 邮箱，唯一         |
| created_at | datetime | 创建时间          |
| updated_at | datetime | 更新时间          |




### posts 表


| 字段         | 类型       | 说明             |
| ---------- | -------- | -------------- |
| id         | uint     | 主键             |
| title      | string   | 文章标题           |
| content    | text     | 文章内容           |
| user_id    | uint     | 外键，关联 users.id |
| created_at | datetime | 创建时间           |
| updated_at | datetime | 更新时间           |




### comments 表


| 字段         | 类型       | 说明             |
| ---------- | -------- | -------------- |
| id         | uint     | 主键             |
| content    | text     | 评论内容           |
| user_id    | uint     | 外键，关联 users.id |
| post_id    | uint     | 外键，关联 posts.id |
| created_at | datetime | 创建时间           |




## 安装与启动



### 1. 克隆 / 进入项目

```bash
cd homeWork/lesson04/ApiuseGIN
```



### 2. 安装依赖

```bash
go mod tidy
```



### 3. 配置环境变量

```bash
cp .env.example .env
```

`.env` 内容说明：

```env
PORT=3001                          # 服务端口
JWT_SECRET=change-me-in-production # JWT 签名密钥
DB_PATH=db/blog.db                 # SQLite 文件路径
```



### 4. 启动服务

```bash
go run main.go
```

或编译后运行：

```bash
cd homeWork/lesson04/ApiuseGIN
go build -o bin/blog-api .
./bin/blog-api
```

启动成功后输出：

```
level=INFO msg="database connected" path=db/blog.db
level=INFO msg="server starting" port=3001
```

服务地址：**[http://localhost:3001](http://localhost:3001)**

## API 接口

统一响应格式：

```json
{
  "code": 0,
  "msg": "ok",
  "data": { ... }
}
```

错误时 `code` 为非 0，`msg` 为错误描述。

### 认证相关（无需 Token）


| 方法   | 路径              | 说明          |
| ---- | --------------- | ----------- |
| POST | `/api/register` | 用户注册        |
| POST | `/api/login`    | 用户登录，返回 JWT |




### 文章（部分需 Token）


| 方法     | 路径               | 认证    | 说明     |
| ------ | ---------------- | ----- | ------ |
| GET    | `/api/posts`     | 否     | 获取所有文章 |
| GET    | `/api/posts/:id` | 否     | 获取单篇文章 |
| POST   | `/api/posts`     | ✅     | 创建文章   |
| PUT    | `/api/posts/:id` | ✅ 仅作者 | 更新文章   |
| DELETE | `/api/posts/:id` | ✅ 仅作者 | 删除文章   |




### 评论


| 方法   | 路径                        | 认证  | 说明        |
| ---- | ------------------------- | --- | --------- |
| GET  | `/api/posts/:id/comments` | 否   | 获取某篇文章的评论 |
| POST | `/api/posts/:id/comments` | ✅   | 发表评论      |




### 当前用户


| 方法  | 路径        | 认证  | 说明         |
| --- | --------- | --- | ---------- |
| GET | `/api/me` | ✅   | 获取当前登录用户信息 |


认证方式：请求头携带

```
Authorization: Bearer <token>
```



## Postman / curl 测试用例



### 1. 注册用户

```bash
curl -X POST http://localhost:3001/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo",
    "password": "123456",
    "email": "demo@example.com"
  }'
```

预期：`code: 0`，返回用户 id、username、email。

### 2. 登录

```bash
curl -X POST http://localhost:3001/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo",
    "password": "123456"
  }'
```

预期：`code: 0`，返回 `token` 字段，后续请求使用该 token。

### 3. 创建文章（需 Token）

```bash
curl -X POST http://localhost:3001/api/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "我的第一篇文章",
    "content": "Hello, Blog!"
  }'
```



### 4. 获取文章列表

```bash
curl http://localhost:3001/api/posts
```



### 5. 获取单篇文章

```bash
curl http://localhost:3001/api/posts/1
```



### 6. 更新文章（需 Token，仅作者）

```bash
curl -X PUT http://localhost:3001/api/posts/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "更新后的标题",
    "content": "更新后的内容"
  }'
```



### 7. 发表评论（需 Token）

```bash
curl -X POST http://localhost:3001/api/posts/1/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "content": "写得不错！"
  }'
```



### 8. 获取文章评论

```bash
curl http://localhost:3001/api/posts/1/comments
```



### 9. 删除文章（需 Token，仅作者）

```bash
curl -X DELETE http://localhost:3001/api/posts/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```



### 10. 错误场景测试


| 场景            | 请求                                | 预期状态码 |
| ------------- | --------------------------------- | ----- |
| 未带 Token 创建文章 | POST `/api/posts` 无 Authorization | 401   |
| 非作者更新他人文章     | PUT `/api/posts/:id` 用他人 Token    | 403   |
| 访问不存在的文章      | GET `/api/posts/9999`             | 404   |
| 重复注册相同用户名     | POST `/api/register` 重复 username  | 409   |




## 功能说明

- **密码安全**：注册时使用 bcrypt 加密存储，登录时比对哈希值，明文密码不写入数据库。
- **JWT 认证**：登录成功后签发 JWT（有效期 7 天），受保护接口通过中间件验证 Token。
- **作者权限**：更新和删除文章时校验 `post.user_id == 当前用户 id`，非作者返回 403。
- **错误处理**：参数错误 400、未认证 401、无权限 403、资源不存在 404、服务器错误 500。
- **日志记录**：使用 `log/slog` 记录数据库连接、HTTP 请求、错误信息。



## 常见问题

**Q:** `go mod tidy` **下载超时？**

设置 `GOPROXY=https://goproxy.cn,direct` 后重试。

**Q: 数据库文件在哪？**

默认在 `db/blog.db`，首次启动自动创建。

**Q: 如何重置数据？**

停止服务后删除 `db/blog.db`，重新启动即可。