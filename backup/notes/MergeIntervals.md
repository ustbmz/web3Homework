# Notes

Code walkthroughs live here. This directory is gitignored.

---

## 56. Merge Intervals

Sort by start, then scan left to right. Each interval only needs to be compared with the last merged one.

- `cur[0] <= last[1]`: overlap (touching endpoints count). Extend the end to the larger of the two.
- otherwise: no overlap, append the current interval.

```
[1,3] [2,6] [8,10] [15,18]
  ↓ sort, then scan
[1,3] overlaps [2,6] → [1,6]
[1,6] does not overlap [8,10] → append
[8,10] does not overlap [15,18] → append
result [[1,6],[8,10],[15,18]]
```

Unsorted input like `[[4,7],[1,4]]` becomes `[[1,4],[4,7]]` after sort; touching endpoints merge to `[1,7]`.

```go
func merge(intervals [][]int) [][]int {
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})

	merged := [][]int{intervals[0]}
	for i := 1; i < len(intervals); i++ {
		cur := intervals[i]
		last := merged[len(merged)-1]
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
```

Time O(n log n) for the sort. Space O(n) for the result (sort may use extra space too).
