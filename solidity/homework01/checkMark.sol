// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract Voting {
    // 候选人得票数
    mapping(string => uint256) private votes;
    // 已出现过的候选人，供 resetVotes 遍历
    string[] private candidates;
    mapping(string => bool) private isCandidate;

    // 给候选人投一票。第一次出现时记入名单，方便之后重置。
    function vote(string calldata candidate) public {
        if (!isCandidate[candidate]) {
            isCandidate[candidate] = true;
            candidates.push(candidate);
        }
        votes[candidate] += 1;
    }

    // 查询某个候选人当前的得票数。没投过的候选人返回 0。
    function getVotes(string calldata candidate) public view returns (uint256) {
        return votes[candidate];
    }

    // 把所有候选人的得票数清零，并清空名单，下次投票会重新登记。
    function resetVotes() public {
        for (uint256 i = 0; i < candidates.length; i++) {
            votes[candidates[i]] = 0;
            isCandidate[candidates[i]] = false;
        }
        delete candidates;
    }
}

contract ReverseString {
    // 反转字符串。例如输入 "abcde"，返回 "edcba"。
    function reverseString(string memory str) public pure returns (string memory) {
        bytes memory chars = bytes(str);
        uint256 left = 0;
        uint256 right = chars.length;

        // 从两端向中间交换字符，直到相遇
        while (left < right) {
            right--;
            bytes1 temp = chars[left];
            chars[left] = chars[right];
            chars[right] = temp;
            left++;
        }

        return string(chars);
    }
}


