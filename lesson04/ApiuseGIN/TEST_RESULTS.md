# 接口测试结果

测试时间：2026-08-16  
环境：`go run main.go`，端口 `3001`，SQLite `db/blog.db`（测试前清空）  
工具：curl

## 测试结论

| 序号 | 场景 | 预期 | 结果 |
| ---- | ---- | ---- | ---- |
| 1 | 注册用户 | `code: 0` | 通过 |
| 2 | 登录获取 JWT | 返回 `token` | 通过 |
| 3 | 带 Token 创建文章 | `code: 0` | 通过 |
| 4 | 获取文章列表 | 返回文章数组 | 通过 |
| 5 | 获取单篇文章 | 返回文章详情 | 通过 |
| 6 | 作者更新文章 | 标题/内容已更新 | 通过 |
| 7 | 发表评论 | `code: 0` | 通过 |
| 8 | 获取评论列表 | 返回评论数组 | 通过 |
| 9 | 作者删除文章 | `post deleted` | 通过 |
| 10 | 无 Token 创建文章 | HTTP 401 | 通过 |

---

## 详细输出

### 1. 注册用户

```bash
curl -X POST http://localhost:3001/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"123456","email":"demo@example.com"}'
```

```json
{"code":0,"msg":"register success","data":{"email":"demo@example.com","id":1,"username":"demo"}}
```

### 2. 登录

```bash
curl -X POST http://localhost:3001/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"123456"}'
```

```json
{"code":0,"data":{"email":"demo@example.com","id":1,"username":"demo"},"msg":"login success","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

### 3. 创建文章（需 Token）

```json
{"code":0,"msg":"post created","data":{"id":1,"title":"我的第一篇文章","content":"Hello, Blog!","user_id":1,...}}
```

### 4. 文章列表

```json
{"code":0,"msg":"ok","data":[{"id":1,"title":"我的第一篇文章",...}]}
```

### 5. 单篇文章

```json
{"code":0,"msg":"ok","data":{"id":1,"title":"我的第一篇文章",...}}
```

### 6. 更新文章

```json
{"code":0,"msg":"post updated","data":{"id":1,"title":"更新后的标题","content":"更新后的内容",...}}
```

### 7. 发表评论

```json
{"code":0,"msg":"comment created","data":{"id":1,"content":"写得不错！","post_id":1,...}}
```

### 8. 评论列表

```json
{"code":0,"msg":"ok","data":[{"id":1,"content":"写得不错！",...}]}
```

### 9. 删除文章

```json
{"code":0,"msg":"post deleted"}
```

### 10. 未带 Token 创建文章（错误场景）

```json
{"code":401,"msg":"missing authorization header"}
HTTP 401
```

---

## 复现方式

```bash
cd lesson04/ApiuseGIN
cp .env.example .env
go run main.go
```

另开终端，按 [README.md](./README.md) 中的 curl 命令逐条执行即可。
