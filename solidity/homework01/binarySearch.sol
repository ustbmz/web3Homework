// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract BinarySearch {
    // 在升序数组中查找 target，找到返回下标，找不到返回 -1。
    function search(uint256[] memory nums, uint256 target) public pure returns (int256) {
        uint256 left = 0;
        uint256 right = nums.length;

        // 每次只看中间元素，把查找范围缩小一半
        while (left < right) {
            uint256 mid = left + (right - left) / 2;
            if (nums[mid] == target) {
                return int256(mid);
            } else if (nums[mid] < target) {
                left = mid + 1;
            } else {
                right = mid;
            }
        }


        return -1;
    }
}
