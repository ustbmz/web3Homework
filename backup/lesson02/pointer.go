package main

import "fmt"

// addTen 接收一个整数指针，把指针指向的值增加 10。
//
// 原理：
//   Go 函数参数一律是「值传递」。如果直接传 int，函数拿到的是副本，
//   改副本不会影响调用方。传 *int 时，拷贝的是「地址」本身，
//   通过 *p 解引用就能改到同一块内存，效果等价于引用传递。
func addTen(p *int) {
	if p == nil { // 防御：空指针解引用会 panic
		return
	}
	*p += 10 // *p 取出指针指向的值，加 10 后再写回原地址
}

// doubleSlice 接收「切片的指针」，把切片里每个元素乘以 2。
//
// 原理：
//   切片头（slice header）本身就含有指向底层数组的指针。
//   只改元素时，直接传 []int 也能改到原数组。
//   这里按题目要求传 *[]int：通过 (*s)[i] 访问元素。
//   若函数里要 append 并让调用方看到新切片，就必须用指针改 header。
func doubleSlice(s *[]int) {
	if s == nil {
		return
	}
	for i := range *s {
		(*s)[i] *= 2
	}
}

func main() {
	// ---------- 题目 1：整数指针 +10 ----------
	n := 5
	fmt.Printf("调用前 n = %d，地址 = %p\n", n, &n)
	addTen(&n) // 把 n 的地址传进去
	fmt.Printf("调用后 n = %d（期望 15）\n\n", n)

	// ---------- 题目 2：切片指针，元素 ×2 ----------
	nums := []int{1, 2, 3, 4, 5}
	fmt.Printf("调用前 nums = %v\n", nums)
	doubleSlice(&nums)
	fmt.Printf("调用后 nums = %v（期望 [2 4 6 8 10]）\n", nums)
}
