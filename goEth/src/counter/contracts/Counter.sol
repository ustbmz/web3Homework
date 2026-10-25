// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

// Counter 是一个最小的计数器，用来演示 abigen 绑定后的读写。
contract Counter {
    uint256 public count;

    event Incremented(address indexed sender, uint256 newCount);

    function increment() external {
        count += 1;
        emit Incremented(msg.sender, count);
    }
}
