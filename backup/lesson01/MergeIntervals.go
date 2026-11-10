//go:build ignore

package main

import (
	"fmt"
	"sort"
)

func mergeIntervals(intervals [][]int) [][]int {
	if len(intervals) == 0 {
		return [][]int{}
	}

	// 按区间起点升序排序，保证从左到右扫描时只需和「上一个合并结果」比较
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})

	// merged 存已合并的区间；先把第一个区间放进去
	merged := [][]int{intervals[0]}

	for i := 1; i < len(intervals); i++ {
		cur := intervals[i]
		lastIdx := len(merged) - 1

		// 当前区间起点 <= 上一个区间终点 → 有重叠（含端点相接），合并
		if cur[0] <= merged[lastIdx][1] {
			// 必须直接改 merged 里的元素；last := merged[lastIdx] 只是副本，改副本无效
			merged[lastIdx][1] = max(merged[lastIdx][1], cur[1])
		} else {
			// 不重叠，把当前区间追加到结果
			merged = append(merged, cur)
		}
	}

	return merged
}

func main() {
	intervals := [][]int{{1, 3}, {2, 6}, {8, 10}, {15, 18}}
	fmt.Println(mergeIntervals(intervals)) // [[1 6] [8 10] [15 18]]

	intervals = [][]int{{1, 4}, {4, 5}}
	fmt.Println(mergeIntervals(intervals)) // [[1 5]]

	intervals = [][]int{{1, 4}, {0, 4}}
	fmt.Println(mergeIntervals(intervals)) // [[0 4]] 未排序输入也会先 sort
}
