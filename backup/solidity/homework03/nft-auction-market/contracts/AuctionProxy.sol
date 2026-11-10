// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

/// @title AuctionProxy
/// @notice 用户实际交互的地址是这个代理，不是实现合约。
///         代理自己几乎没有业务代码：它把每次调用 delegatecall 到当前的实现合约，
///         数据存在代理里。换实现合约时，这个地址和里面的拍卖记录都不变。
contract AuctionProxy is ERC1967Proxy {
    constructor(address implementation, bytes memory data) ERC1967Proxy(implementation, data) {}
}
