// 66. - 加一
package main

import "fmt"

func plusOne(digits []int) []int {
	// 从最低位（数组末尾）向高位遍历，模拟加 1 的进位
	for i := len(digits) - 1; i >= 0; i-- {
		digits[i]++ // 当前位加 1
		if digits[i] < 10 {
			return digits // 无进位，直接返回
		}
		digits[i] = 0 // 当前位变成 0，进位继续传给前一位
	}
	return append([]int{1}, digits...)
}

func main() {
	fmt.Println(plusOne([]int{1, 2, 3}))    // 示例 1 -> [1 2 4]
	fmt.Println(plusOne([]int{4, 3, 2, 1})) // 示例 2 -> [4 3 2 2]
	fmt.Println(plusOne([]int{9}))          // 示例 3 -> [1 0]
	fmt.Println(plusOne([]int{0, 0, 0, 9}))
	fmt.Println(plusOne([]int{3, 2, 4}))
	fmt.Println(plusOne([]int{1, 2, 4, 9}))
	fmt.Println(plusOne([]int{9, 9, 9, 9}))

}
