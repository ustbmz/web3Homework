package main

import (
	"fmt"
	"sync"
)

// produceUnbuffered 向无缓冲通道发送 1~10，发完后关闭通道。
//
// 原理：
//   无缓冲通道（make(chan int)）的发送和接收必须同时就绪，否则会阻塞。
//   这保证了「发一个、收一个」的同步交接。
//   发送方关闭通道后，接收方 range 会正常结束；重复 close 会 panic。
func produceUnbuffered(ch chan<- int) {
	defer close(ch)
	for i := 1; i <= 10; i++ {
		ch <- i
		fmt.Printf("[生产者] 发送 %d\n", i)
	}
}

// consumeUnbuffered 从通道接收直到通道被关闭。
func consumeUnbuffered(ch <-chan int) {
	for n := range ch {
		fmt.Printf("[消费者] 收到 %d\n", n)
	}
}

// produceBuffered 向带缓冲通道发送 100 个整数。
//
// 原理：
//   缓冲通道 make(chan int, N) 内部有长度为 N 的队列。
//   队列未满时发送不阻塞；未空时接收不阻塞。
//   缓冲满了发送方才等消费者，缓冲空了消费者才等生产者。
func produceBuffered(ch chan<- int, n int) {
	defer close(ch)
	for i := 1; i <= n; i++ {
		ch <- i
	}
	fmt.Printf("[缓冲生产者] 已发送 %d 个整数并关闭通道\n", n)
}

// consumeBuffered 从缓冲通道接收并打印。
func consumeBuffered(ch <-chan int) {
	count := 0
	for n := range ch {
		fmt.Printf("[缓冲消费者] %d\n", n)
		count++
	}
	fmt.Printf("[缓冲消费者] 共收到 %d 个整数\n", count)
}

func main() {
	// ---------- 题目 1：无缓冲通道，两协程通信 1~10 ----------
	fmt.Println("===== 无缓冲通道 =====")
	unbuf := make(chan int) // 容量 0：同步通道
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		produceUnbuffered(unbuf)
	}()
	go func() {
		defer wg.Done()
		consumeUnbuffered(unbuf)
	}()
	wg.Wait()

	// ---------- 题目 2：缓冲通道，生产 100 个整数 ----------
	fmt.Println("\n===== 缓冲通道（容量 16）=====")
	buf := make(chan int, 16) // 容量 16：发送可先于接收积压最多 16 个
	wg.Add(2)
	go func() {
		defer wg.Done()
		produceBuffered(buf, 100)
	}()
	go func() {
		defer wg.Done()
		consumeBuffered(buf)
	}()
	wg.Wait()
}
