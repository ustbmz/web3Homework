# 老师可能提问的问题（Q&A）

本文档整理个人博客 API 项目中**常问的问题**及参考答案。

配套文档：[README.md](./README.md)（API 与启动） · [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)（手写思路）

---

## 一、项目整体

### Q1：这个项目做了什么？用了哪些技术？

**答：** 这是一个个人博客的后端 REST API，实现了用户注册/登录、文章 CRUD、评论功能。技术栈是 Go + Gin（Web 框架）+ GORM（ORM）+ SQLite（数据库），认证用 JWT，密码用 bcrypt 加密，配置从 `.env` 读取。

---

### Q2：为什么要把代码分成 config / models / handlers / middleware / utils 这些包？

**答：** 分层是为了**职责清晰、便于维护**：

- `models`：只定义数据结构，不关心 HTTP
- `config`：配置和数据库连接
- `utils`：可复用的工具（响应格式、密码、JWT）
- `middleware`：横切逻辑（认证、日志、错误），所有路由共用
- `handlers`：具体业务接口
- `main`：组装路由、启动服务

改认证逻辑只动 middleware，改表结构只动 models，互不影响。

---

### Q3：请求从进入到返回，经过哪些环节？

**答：**

1. 客户端发 HTTP 请求到 Gin
2. 全局中间件：`Recovery` → `Logger` → `ErrorHandler`
3. 若路由在认证组，再走 `Auth` 中间件验 JWT
4. 对应 `handler` 校验参数、操作数据库
5. 通过 `utils.Success` / `utils.Fail` 返回统一 JSON
6. `Logger` 记录 method、path、status

---

### Q4：为什么用 SQLite 而不是 MySQL？

**答：** 作业/个人博客规模小，SQLite 是文件型数据库，**零配置、部署简单**，适合本地开发和演示。GORM 换驱动即可迁移到 MySQL/PostgreSQL，模型层基本不用改。

---

## 二、Gin 框架

### Q5：Gin 是什么？和 Go 标准库 net/http 有什么区别？

**答：** Gin 是 Go 的高性能 Web 框架，基于 `httprouter` 提供路由、中间件、参数绑定、JSON 响应等能力。比直接用 `net/http` 写路由和解析请求**更省事**，社区生态成熟。

---

### Q6：`gin.Default()` 和 `gin.New()` 有什么区别？本项目为什么用 `gin.New()`？

**答：** `gin.Default()` 自带 Logger 和 Recovery 中间件；`gin.New()` 是空引擎，中间件自己挂。本项目用 `gin.New()` 是为了**自定义** `Logger` 和 `ErrorHandler`（用 `log/slog`），控制更灵活。

---

### Q7：什么是中间件（Middleware）？执行顺序是怎样的？

**答：** 中间件是包在 handler 外面的函数，签名是 `func(c *gin.Context)`，通过 `c.Next()` 调用下一个中间件或 handler。

本项目全局顺序：

```
Recovery → Logger → ErrorHandler → [Auth（仅认证路由）] → Handler
```

`Auth` 失败会 `Abort`，不再执行后面的 handler。

---

### Q8：路由分组 `r.Group("/api")` 有什么好处？

**答：** 同一前缀的路由写在一起，方便统一管理；认证路由可以单独挂 `Auth` 中间件，**公开接口和需登录接口分开**，不用在每个 handler 里重复验 Token。

---

### Q9：`c.ShouldBindJSON` 是干什么的？

**答：** 把请求体 JSON **自动解析**到 struct，并按 `binding` 标签做校验（如 `required`、`min`、`email`）。校验失败直接返回 400，减少手写 if 判断。

---

### Q10：Gin Context 是什么？为什么把 userID 存在 Context 里？

**答：** `*gin.Context` 封装了本次请求的所有信息（请求头、参数、响应写入等），贯穿整条处理链。`Auth` 中间件解析 JWT 后 `c.Set("userID", id)`，handler 用 `GetUserID(c)` 取出，**认证和业务解耦**。

---

## 三、数据库与 GORM

### Q11：三张表之间是什么关系？

**答：**

- **User 1 : N Post** — 一个用户可发多篇文章（`posts.user_id`）
- **User 1 : N Comment** — 一个用户可发多条评论（`comments.user_id`）
- **Post 1 : N Comment** — 一篇文章下有多条评论（`comments.post_id`）

---

### Q12：GORM 的 `AutoMigrate` 做了什么？

**答：** 根据 struct 和 gorm 标签**自动建表或补字段**，不会删已有列。本项目启动时对 `User`、`Post`、`Comment` 三张表执行迁移，无需手写 SQL 建表。

---

### Q13：struct 上的 gorm 标签和 json 标签分别有什么用？

**答：**

- **gorm 标签**：控制数据库行为，如 `primaryKey`、`uniqueIndex`、`not null`、`foreignKey`
- **json 标签**：控制序列化成 JSON 的字段名；`json:"-"` 表示不输出（如 Password）

---

### Q14：`Preload("User")` 是什么意思？

**答：** GORM 的**预加载关联**，查 Post 或 Comment 时一并查出关联的 User，避免 N+1 查询。例如文章列表返回时每篇文章带作者 username。

---

### Q15：删除文章时为什么要用 `Transaction`？

**答：** 删文章时要**先删该文章下所有评论**，再删文章。用事务保证两步要么全成功要么全失败，避免出现「文章删了评论还在」的数据不一致。

---

### Q16：为什么 Comment 没有 `UpdatedAt`？

**答：** 评论一般**只创建不修改**，业务上不需要更新时间；User 和 Post 可能被编辑，所以有 `UpdatedAt`。

---

## 四、认证与安全

### Q17：JWT 是什么？登录成功后怎么用它？

**答：** JWT（JSON Web Token）是一种**无状态**的令牌，服务端签发后客户端保存，之后请求在 Header 里带：

```
Authorization: Bearer <token>
```

服务端用密钥验证签名，从 Token 里取出 userID，无需每次查 Session 表。

---

### Q18：JWT 里存了哪些信息？有效期多久？

**答：** 自定义 Claims 存 `user_id`、`username`；标准字段有 `exp`（过期时间）、`iat`（签发时间）。本项目有效期 **7 天**，算法 **HS256**，密钥来自环境变量 `JWT_SECRET`。

---

### Q19：为什么用 bcrypt 存密码，不用 MD5？

**答：** MD5 快但**不安全**，易被彩虹表破解。bcrypt 是专为密码设计的慢哈希，带随机 salt，**不可逆**，即使数据库泄露也难以还原明文。登录时用 `CompareHashAndPassword` 比对。

---

### Q20：登录时「用户不存在」和「密码错误」为什么返回同样的 401？

**答：** 防止攻击者通过不同错误信息**枚举有效用户名**。统一返回 `invalid username or password`，不泄露用户是否存在。

---

### Q21：401 和 403 有什么区别？

**答：**

- **401 Unauthorized**：未登录或 Token 无效/过期 — 「你是谁？」
- **403 Forbidden**：已登录但**没有权限** — 「知道你是谁，但不能做这件事」

例如：没带 Token 创建文章 → 401；用他人 Token 改别人文章 → 403。

---

### Q22：作者权限是怎么实现的？

**答：** Update/Delete 文章时，先查出 `post.UserID`，与 `middleware.GetUserID(c)` 得到的当前用户 ID 比较；不等则返回 403 `only author can update/delete this post`。

---

### Q23：JWT 存在哪里？服务端需要存 Session 吗？

**答：** Token 由**客户端保存**（内存/localStorage 等），服务端不存 Session 表。服务端只保留 `JWT_SECRET` 用于签发和验证。这是 JWT **无状态**的特点，便于水平扩展。

---

### Q24：`JWT_SECRET` 为什么必须改？写在代码里行不行？

**答：** 密钥泄露则任何人可伪造 Token。应放在 `.env`，生产环境用**强随机字符串**，且不提交到 Git。写在代码里不安全，也不便不同环境用不同密钥。

---

## 五、API 设计

### Q25：什么是 REST API？本项目哪些接口是 RESTful 的？

**答：** REST 用 HTTP 方法和 URL 表示资源操作。例如：

| 方法 | 路径 | 含义 |
| --- | --- | --- |
| GET | `/api/posts` | 列表 |
| GET | `/api/posts/:id` | 详情 |
| POST | `/api/posts` | 创建 |
| PUT | `/api/posts/:id` | 更新 |
| DELETE | `/api/posts/:id` | 删除 |

评论嵌套在 `/api/posts/:id/comments` 下，表示「某篇文章的评论」资源。

---

### Q26：统一响应格式 `{ code, msg, data }` 有什么好处？

**答：** 前端只需判断 `code === 0` 即成功，错误信息在 `msg`；`data` 承载业务数据。格式一致，**前后端约定清晰**，便于封装 axios/fetch。

---

### Q27：为什么登录接口的 JSON 里 `token` 和 `data` 分开返回？

**答：** 登录响应除了用户基本信息（`data`），还需要顶层 `token` 字段方便客户端**直接取用**存本地；这是本项目的小约定，也可统一放进 `data.token`。

---

### Q28：哪些接口需要 Token，哪些不需要？

**答：**

**不需要：** 注册、登录、文章列表/详情、评论列表

**需要：** `/api/me`、创建/更新/删除文章、发表评论

读多写少，浏览公开；写操作需身份。

---

## 六、错误处理与日志

### Q29：项目里用了哪些 HTTP 状态码？

**答：**

| 状态码 | 场景 |
| --- | --- |
| 200 | 成功 |
| 400 | 参数校验失败 |
| 401 | 未认证 / Token 无效 |
| 403 | 无权限（非作者） |
| 404 | 用户/文章不存在 |
| 409 | 用户名或邮箱重复注册 |
| 500 | 服务器内部错误 |

---

### Q30：`Fail` 和 `AbortFail` 有什么区别？

**答：** 都返回错误 JSON；`AbortFail` 额外调用 `c.Abort()`，**中断后续中间件和 handler**。认证中间件必须用 `AbortFail`，否则 Token 无效时仍可能执行到业务 handler。

---

### Q31：为什么用 `log/slog` 而不是 `fmt.Println`？

**答：** `slog` 是 Go 1.21+ 标准库**结构化日志**，带 level、键值对，便于过滤和接入日志系统；比 `Println` 更适合生产环境。

---

### Q32：`gin.Recovery()` 是干什么的？

**答：** 捕获 handler 里的 **panic**，返回 500 并记录堆栈，**避免整个进程崩溃**。是生产环境必备中间件。

---

## 七、配置与部署

### Q33：`.env` 里配置了哪些项？各有什么用？

**答：**

| 变量 | 作用 |
| --- | --- |
| `PORT` | 服务端口，默认 3001 |
| `JWT_SECRET` | JWT 签名密钥 |
| `DB_PATH` | SQLite 文件路径，如 `db/blog.db` |

---

### Q34：如何重置数据库？

**答：** 停止服务 → 删除 `db/blog.db` → 重新启动，`AutoMigrate` 会重建空表。

---

### Q35：`go run main.go` 和 `go build` 再运行有什么区别？

**答：** `go run` 编译到临时目录并立即运行，适合开发；`go build -o bin/blog-api` 生成**可执行文件**，适合部署到服务器反复运行，无需每次编译。

---

## 八、扩展思考题（老师可能追问）

### Q36：如果 Token 被盗用了怎么办？

**答：** 当前方案无服务端黑名单，被盗 Token 在过期前仍有效。改进方向：缩短有效期 + Refresh Token、服务端维护 Token 黑名单、HTTPS 传输、敏感操作二次验证。

---

### Q37：为什么不用 Session + Cookie？

**答：** Session 需在服务端存状态，多机部署要共享 Session 存储。JWT 无状态，适合 API 和前后端分离；Cookie 还有跨域问题，移动端常用 Header 带 Token。

---

### Q38：评论能不能删？能不能改别人的文章？

**答：** 当前版本**没有**删评论接口；**不能**改别人的文章，Update/Delete 有作者校验。这是有意控制 scope，作业中说明设计取舍即可。

---

### Q39：GORM 的 N+1 问题是什么？本项目怎么避免？

**答：** 查 N 篇文章再循环查每篇的作者，会产生 1+N 次查询。本项目用 `Preload("User")` **一次 JOIN 或批量查询**关联数据，减少数据库往返。

---

### Q40：如果并发很高，SQLite 够用吗？

**答：** SQLite 写并发有限，适合读多写少的个人博客/本地 demo。线上高并发应换 PostgreSQL/MySQL，并考虑连接池、缓存等；GORM 换 `Open` 的 driver 即可。

---

## 九、快速自测（合上文档能答出来就稳了）

1. 说清 **User / Post / Comment** 表关系和字段含义  
2. 画一条请求链：**HTTP → 中间件 → Handler → DB → JSON**  
3. 解释 **JWT 签发、携带、验证** 全流程  
4. 解释 **bcrypt** 为什么比 MD5 适合存密码  
5. 举例 **401 vs 403** 各对应哪个接口场景  
6. 说明 **路由分组 + Auth 中间件** 如何保护接口  
7. 说明 **Preload** 和 **Transaction** 在本项目里的用途  

---

## 答辩小技巧

- 先答**结论**，再补**一两句原理或代码位置**（如「在 `middleware/auth.go` 里解析 Bearer Token」）
- 不会的扩展题（Q36–Q40）诚实说「当前版本没做，可以改进为…」，比瞎编好
- 结合 **curl 或 Postman 演示** 注册 → 登录 → 带 Token 发文章，印象分更高
