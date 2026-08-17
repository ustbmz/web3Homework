### 字符串

- **有效的括号** 
考察：字符串处理、栈的使用
题目：给定一个只包括 '('，')'，'{'，'}'，'['，']' 的字符串，判断字符串是否有效
链接：[https://leetcode-cn.com/problems/valid-parentheses/](https://leetcode-cn.com/problems/valid-parentheses/)

- **最长公共前缀**
考察：字符串处理、循环嵌套
题目：查找字符串数组中的最长公共前缀
链接：[https://leetcode-cn.com/problems/longest-common-prefix/](https://leetcode-cn.com/problems/longest-common-prefix/)
题解: [https://leetcode.cn/problems/valid-parentheses/solutions/4007260/you-xiao-gua-hao-zhan-jie-fa-by-ustbmz-2os4/](https://leetcode.cn/problems/valid-parentheses/solutions/4007260/you-xiao-gua-hao-zhan-jie-fa-by-ustbmz-2os4/)

### 基本值类型

- **加一**
难度：简单
考察：数组操作、进位处理
题目：给定一个由整数组成的非空数组所表示的非负整数，在该数的基础上加一
链接：[https://leetcode-cn.com/problems/plus-one/](https://leetcode-cn.com/problems/plus-one/)
题解：[https://leetcode.cn/problems/plus-one/solutions/4012745/jia-yi-ti-jie-by-ustbmz-l8pf/](https://leetcode.cn/problems/plus-one/solutions/4012745/jia-yi-ti-jie-by-ustbmz-l8pf/)

### 引用类型：切片

- **删除有序数组中的重复项**

- ([https://leetcode.cn/problems/remove-duplicates-from-sorted-array/)**：给你一个有序数组](https://leetcode.cn/problems/remove-duplicates-from-sorted-array/)**：给你一个有序数组) `nums` ，请你原地删除重复出现的元素，使每个元素只出现一次，返回删除后数组的新长度。不要使用额外的数组空间，你必须在原地修改输入数组并在使用 O(1) 额外空间的条件下完成。可以使用双指针法，一个慢指针 `i` 用于记录不重复元素的位置，一个快指针 `j` 用于遍历数组，当 `nums[i]` 与 `nums[j]` 不相等时，将 `nums[j]` 赋值给 `nums[i + 1]`，并将 `i` 后移一位。
链接：[https://leetcode-cn.com/problems/remove-duplicates-from-sorted-array/](https://leetcode-cn.com/problems/remove-duplicates-from-sorted-array/)
题解：[https://leetcode.cn/problems/remove-duplicates-from-sorted-array/solutions/4012789/26-shan-chu-you-xu-shu-zu-zhong-de-zhong-bcq0/](https://leetcode.cn/problems/remove-duplicates-from-sorted-array/solutions/4012789/26-shan-chu-you-xu-shu-zu-zhong-de-zhong-bcq0/)

-  **合并区间**

- (https://leetcode.cn/problems/merge-intervals/)**：以数组) `intervals` 表示若干个区间的集合，其中单个区间为 `intervals[i] = [starti, endi]` 。请你合并所有重叠的区间，并返回一个不重叠的区间数组，该数组需恰好覆盖输入中的所有区间。可以先对区间数组按照区间的起始位置进行排序，然后使用一个切片来存储合并后的区间，遍历排序后的区间数组，将当前区间与切片中最后一个区间进行比较，如果有重叠，则合并区间；如果没有重叠，则将当前区间添加到切片中。
题解：[https://leetcode.cn/problems/two-sum/solutions/4012802/liang-shu-zhi-he-ti-jie-by-ustbmz-snbz/](https://leetcode.cn/problems/two-sum/solutions/4012802/liang-shu-zhi-he-ti-jie-by-ustbmz-snbz/)