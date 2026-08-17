//go:build ignore

package main

import "fmt"

func twoSum(nums []int, target int) []int {
	// map[数值] = 下标，记录已经遍历过的数
	seen := make(map[int]int)

	for i, x := range nums {
		need := target - x // 当前数 x 需要的「另一半」

		// 先查 map：之前是否出现过 need
		if j, ok := seen[need]; ok {
			return []int{j, i} // 找到配对，返回两个下标
		}

		// 没找到再把当前数记下；必须「先查后存」，避免同一元素用两次
		seen[x] = i
	}

	return nil // 题目保证有解，正常不会走到这里
}

func main() {
	fmt.Println(twoSum([]int{2, 7, 11, 15}, 9))  // [0 1]
	fmt.Println(twoSum([]int{3, 2, 4}, 6))        // [1 2]
	fmt.Println(twoSum([]int{3, 3}, 6))           // [0 1]
}
