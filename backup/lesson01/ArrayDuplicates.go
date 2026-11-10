//go:build ignore

package main

// 删除有序数组中的重复项
import "fmt"

func removeDuplicates(nums []int) int {
	// 空数组没有唯一元素
	if len(nums) == 0 {
		return 0
	}

	// slow(i)：去重后「有效区」的最后一个下标，nums[0..i] 始终是不重复且有序的
	// fast(j)：扫描指针，从 1 开始向后找「下一个与 slow 不同的数」
	i := 0
	for j := 1; j < len(nums); j++ {
		if nums[j] != nums[i] {
			// 发现新数字：先扩展有效区，再把新数写进有效区末尾
			i++
			nums[i] = nums[j]
		}
		// 若 nums[j] == nums[i]，说明重复，j 继续往后扫即可（不写、不移动 i）
	}

	// i 是最后一个有效元素的下标，唯一元素个数 k = i + 1
	return i + 1
}

func main() {
	nums := []int{1, 1, 2}
	k := removeDuplicates(nums)
	fmt.Println(k, nums[:k]) // 2 [1 2]

	nums = []int{0, 0, 1, 1, 1, 2, 2, 3, 3, 4}
	k = removeDuplicates(nums)
	fmt.Println(k, nums[:k]) // 5 [0 1 2 3 4]
}
