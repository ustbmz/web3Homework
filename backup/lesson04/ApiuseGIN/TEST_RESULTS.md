

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
{"code":0,"msg":"register success","data":{"email":"demo@example.com","id":1,"username":"demo"}}
```

![image-20260817183035087](https://files.seeusercontent.com/2026/08/17/Wbk4/image-20260817183035087.png)

### 2. 登录

```bash
curl -X POST http://localhost:3001/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
{"code":0,"data":{"email":"demo@example.com","id":1,"username":"demo"},"msg":"login success","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

![image-20260817183240588](https://files.seeusercontent.com/2026/08/17/n2Wc/image-20260817183240588.png)

### 3. 创建文章（需 Token）

```bash
curl -X POST http://localhost:3001/api/posts 
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"title":"我的第一篇文章","content":"Hello, Blog!"}'
{"code":0,"msg":"post created","data":{"id":1,"title":"我的第一篇文章","content":"Hello, Blog!","user_id":1,...}}
```



![image-20260817183734117](https://files.seeusercontent.com/2026/08/17/o1Ki/image-20260817183734117.png)

### 4. 文章列表

```bash
curl http://localhost:3001/api/posts
{"code":0,"msg":"ok","data":[{"id":1,"title":"我的第一篇文章",...}]}
```

![image-20260817183809385](https://files.seeusercontent.com/2026/08/17/o1Oi/image-20260817183809385.png)

### 5. 单篇文章

```bash
curl http://localhost:3001/api/posts/1
{"code":0,"msg":"ok","data":{"id":1,"title":"我的第一篇文章",...}}
```

![image-20260817190147957](https://files.seeusercontent.com/2026/08/17/3Tfm/image-20260817190147957.png)

### 6. 更新文章

```bash
curl -X PUT http://localhost:3001/api/posts/1 
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"title":"更新后的标题","content":"更新后的内容"}'
{"code":0,"msg":"post updated","data":{"id":1,"title":"更新后的标题","content":"更新后的内容",...}}
```

### 7. 发表评论

```bash
curl -X POST http://localhost:3001/api/posts/2/comments
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"content":"写得不错！"}'
{"code":0,"msg":"comment created","data":{"id":1,"content":"写得不错！","post_id":1,...}}
```

![image-20260817190633726](https://files.seeusercontent.com/2026/08/17/m7Iw/image-20260817190633726.png)

### 8. 评论列表

```bash
curl http://localhost:3001/api/posts/1/comments
{"code":0,"msg":"ok","data":[{"id":1,"content":"写得不错！",...}]}
```

![image-20260817190944436](https://files.seeusercontent.com/2026/08/17/aD4v/image-20260817190944436.png)

### 9. 删除文章

```bash
curl -X DELETE http://localhost:3001/api/posts/1 
  -H "Authorization: Bearer YOUR_TOKEN"
{"code":0,"msg":"post deleted"}
```

![image-20260817191025677](https://files.seeusercontent.com/2026/08/17/xhR2/image-20260817191025677.png)

### 10. 未带 Token 创建文章（错误场景）

```bash
curl -X POST http://localhost:3001/api/posts 
  -H "Content-Type: application/json" \
  -d '{"title":"未登录发文","content":"should fail"}'
{"code":401,"msg":"missing authorization header"}
HTTP 401

```

![image-20260817191207963](https://files.seeusercontent.com/2026/08/17/j4kK/image-20260817191207963.png)





## 复现方式

```bash
cd lesson04/ApiuseGIN
cp .env.example .env
go run main.go
```

另开终端，按 [README.md](./README.md) 中的命令逐条执行即可。
