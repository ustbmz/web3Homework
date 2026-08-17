# Lesson 02

每个文件独立 `main`，必须指定文件运行：

```bash
go run pointer.go
go run goroutine.go
go run oop.go
go run channel.go
go run lock.go
```

---

## pointer.go

- `addTen(p *int)`：`*p += 10`，改调用方的 `n`
- `doubleSlice(s *[]int)`：`(*s)[i] *= 2`，每个元素乘 2
- `main`：`n=5` → 15；`[1 2 3 4 5]` → `[2 4 6 8 10]`

## goroutine.go

- `printOdd` / `printEven`：两个 `go` 分别打 1~10 的奇偶数；`wg.Wait()` 等它们结束
- `Schedule`：每个任务一个协程，`i, t := i, t` 避免循环变量共享，结果写 `results[i]`
- `main`：先跑奇偶打印，再调度三个 sleep 任务并打印各自耗时

## oop.go

- `Shape` 接口：`Area()`、`Perimeter()`
- `Rectangle` / `Circle`：值接收者，算面积和周长；`printShape` 只接收 `Shape`
- `Employee` 嵌入 `Person`：可直接用 `e.Name`、`e.Age`；`PrintInfo` 打印员工信息

## channel.go

- 无缓冲 `make(chan int)`：生产者发 1~10 后 `close`，消费者 `range` 收完
- 缓冲 `make(chan int, 16)`：生产者发 100 个数后 `close`，消费者收完打印总数
- 两边都用 `WaitGroup` 等结束；`chan<-` 只发，`<-chan` 只收

## lock.go

- `runWithMutex`：10 个协程各 `counter++` 1000 次，每次 `Lock` / `Unlock`
- `runWithAtomic`：同样次数，用 `atomic.AddInt64(&counter, 1)`
- 两种结果都应为 10000
