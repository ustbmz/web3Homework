// 56. Merge Intervals
package main

import (
	"fmt"
	"sort"
)

func merge(intervals [][]int) [][]int {
	// Sort by start so we only compare with the last merged interval.
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})

	merged := [][]int{intervals[0]}

	for i := 1; i < len(intervals); i++ {
		cur := intervals[i]
		last := merged[len(merged)-1]

		// Overlap (including touching endpoints) → extend the end.
		if cur[0] <= last[1] {
			if cur[1] > last[1] {
				last[1] = cur[1]
			}
		} else {
			merged = append(merged, cur)
		}
	}

	return merged
}

func main() {
	fmt.Println(merge([][]int{{1, 3}, {2, 6}, {8, 10}, {15, 18}})) // [[1 6] [8 10] [15 18]]
	fmt.Println(merge([][]int{{1, 4}, {4, 5}}))                    // [[1 5]]
	fmt.Println(merge([][]int{{4, 7}, {1, 4}}))                    // [[1 7]]
}
