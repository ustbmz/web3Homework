// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract BeggingContract {
    // 合约所有者。只有这个地址可以调用 withdraw 把钱提走。
    address public owner;

    // 每个捐赠者累计捐了多少 wei。没捐过的地址读出来是 0。
    mapping(address => uint256) private donations;

    // 部署合约的人成为所有者。msg.sender 是发起这笔部署交易的地址。
    constructor() {
        owner = msg.sender;
    }

    // 捐赠。payable 表示调用时可以附带以太币，附带的金额会进入本合约。
    // msg.sender 是捐赠者，msg.value 是这次附带的 wei。同一地址多次捐赠会累加。
    function donate() public payable {
        donations[msg.sender] += msg.value;
    }

    // 查询某个地址累计捐赠了多少 wei。view 表示只读，不修改状态。
    function getDonation(address donor) public view returns (uint256) {
        return donations[donor];
    }

    // 所有者提取合约里的全部以太币。
    // address(this).balance 是合约当前余额。
    // transfer 要求接收方是 payable 地址，失败时整笔交易回滚。
    function withdraw() public {
        require(msg.sender == owner, "only owner can withdraw");

        payable(owner).transfer(address(this).balance);
    }
}
