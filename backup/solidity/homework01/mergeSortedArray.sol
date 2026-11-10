// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract MergeSortedArray {
    // 合并两个升序数组，返回一个新的升序数组。
    function merge(uint256[] memory a, uint256[] memory b) public pure returns (uint256[] memory) {
        uint256[] memory result = new uint256[](a.length + b.length);
        uint256 i = 0;
        uint256 j = 0;
        uint256 k = 0;

        // 两边都还有数时，每次取较小的那个放进结果
        while (i < a.length && j < b.length) {
            if (a[i] <= b[j]) {
                result[k] = a[i];
                i++;
            } else {
                result[k] = b[j];
                j++;
            }
            k++;
        }

        // 把还没走完的那一边原样接到后面
        while (i < a.length) {
            result[k] = a[i];
            i++;
            k++;
        }
        while (j < b.length) {
            result[k] = b[j];
            j++;
            k++;
        }

        return result;
    }
}