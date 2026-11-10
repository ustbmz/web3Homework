package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	goroutineCount = 10   // 协程数
	incrementsEach = 1000 // 每个协程递增次数
)

// mutexAdd 在互斥锁保护下把 counter 加 1。
//
// 原理：
//   counter++ 实际是「读 → 加 1 → 写」三步，多协程交错会丢更新。
//   Mutex 保证同一时刻只有一个协程进入临界区，读改写是原子的。
func mutexAdd(mu *sync.Mutex, counter *int) {
	mu.Lock()
	*counter++
	mu.Unlock()
}

// runWithMutex 启动 10 个协程，各对共享计数器 +1000 次。
func runWithMutex() int {
	var (
		mu      sync.Mutex
		counter int
		wg      sync.WaitGroup
	)
	wg.Add(goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsEach; j++ {
				mutexAdd(&mu, &counter)
			}
		}()
	}
	wg.Wait()
	return counter
}

// runWithAtomic 用原子加实现无锁计数器。
//
// 原理：
//   atomic.AddInt64 由 CPU 的 LOCK 前缀指令（如 LOCK XADD）保证
//   「读-改-写」在硬件层一次完成，不需要 Mutex。
//   适合这种简单计数；复杂临界区仍要用锁。
func runWithAtomic() int64 {
	var (
		counter int64
		wg      sync.WaitGroup
	)
	wg.Add(goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsEach; j++ {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}
	wg.Wait()
	return counter
}

func main() {
	// ---------- 题目 1：sync.Mutex 保护共享计数器 ----------
	gotMutex := runWithMutex()
	fmt.Printf("Mutex 计数器 = %d（期望 %d）\n",
		gotMutex, goroutineCount*incrementsEach)

	// ---------- 题目 2：sync/atomic 无锁计数器 ----------
	gotAtomic := runWithAtomic()
	fmt.Printf("Atomic 计数器 = %d（期望 %d）\n",
		gotAtomic, goroutineCount*incrementsEach)
}
