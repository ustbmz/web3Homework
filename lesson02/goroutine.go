package main

import (
	"fmt"
	"sync"
	"time"
)

// printOdd 打印 1~10 的奇数。
// 作为独立协程运行时，与 printEven 交错输出，顺序不固定。
func printOdd() {
	for i := 1; i <= 10; i += 2 {
		fmt.Printf("[奇数协程] %d\n", i)
	}
}

// printEven 打印 2~10 的偶数。
func printEven() {
	for i := 2; i <= 10; i += 2 {
		fmt.Printf("[偶数协程] %d\n", i)
	}
}

// Task 表示调度器要执行的一个任务。
type Task struct {
	Name string
	Fn   func()
}

// TaskResult 记录单个任务的名称和耗时。
type TaskResult struct {
	Name     string
	Duration time.Duration
}

// Schedule 并发执行一组任务，并统计每个任务的执行时间。
//
// 原理：
//   每个任务用 go 启动一个协程；WaitGroup 保证 main 等到全部结束。
//   结果写入带下标的切片，各协程写不同下标，无需加锁。
//   time.Since(start) 得到该任务自身的墙钟时间，不是 CPU 时间。
func Schedule(tasks []Task) []TaskResult {
	results := make([]TaskResult, len(tasks))
	var wg sync.WaitGroup
	wg.Add(len(tasks))

	for i, t := range tasks {
		// 循环变量 i、t 每次迭代都会变。必须值捕获，
		// 否则多个协程可能读到最后一次循环的值。
		i, t := i, t
		go func() {
			defer wg.Done()
			start := time.Now()
			t.Fn()
			results[i] = TaskResult{
				Name:     t.Name,
				Duration: time.Since(start),
			}
		}()
	}

	wg.Wait()
	return results
}

func main() {
	// ---------- 题目 1：两个协程分别打印奇数 / 偶数 ----------
	fmt.Println("===== 奇数 / 偶数协程 =====")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		printOdd()
	}()
	go func() {
		defer wg.Done()
		printEven()
	}()
	wg.Wait() // 没有 Wait，main 返回会把未跑完的协程一起结束

	// ---------- 题目 2：任务调度器，并发执行并统计耗时 ----------
	fmt.Println("\n===== 任务调度器 =====")
	tasks := []Task{
		{"下载文件", func() { time.Sleep(80 * time.Millisecond) }},
		{"解析配置", func() { time.Sleep(30 * time.Millisecond) }},
		{"写日志", func() { time.Sleep(50 * time.Millisecond) }},
	}

	start := time.Now()
	results := Schedule(tasks)
	total := time.Since(start)

	for _, r := range results {
		fmt.Printf("任务 %-6s 耗时 %v\n", r.Name, r.Duration)
	}
	fmt.Printf("全部任务墙钟时间 %v（并发，应接近最慢那个任务，而不是三者之和）\n", total)
}
