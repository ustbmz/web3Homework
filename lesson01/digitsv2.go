//go:build ignore

package main

import "fmt"

func plusOne(digits []int) []int {
	carry := 1 // 初始进位为 1，表示整体 +1
	for i := len(digits) - 1; i >= 0 && carry > 0; i-- {
		sum := digits[i] + carry
		digits[i] = sum % 10 // 当前位
		carry = sum / 10     // 进位
	}
	if carry > 0 {
		return append([]int{1}, digits...) // 最高位还有进位，前面补 1
	}
	return digits
}

func main() {
	fmt.Println(plusOne([]int{1, 2, 3}))       // [1 2 4]
	fmt.Println(plusOne([]int{4, 3, 2, 1}))    // [4 3 2 2]
	fmt.Println(plusOne([]int{9}))             // [1 0]
	fmt.Println(plusOne([]int{0, 0, 0, 9}))    // [0 0 1 0]
	fmt.Println(plusOne([]int{1, 2, 4, 9}))    // [1 2 5 0]
	fmt.Println(plusOne([]int{9, 9, 9, 9}))    // [1 0 0 0 0]
}
