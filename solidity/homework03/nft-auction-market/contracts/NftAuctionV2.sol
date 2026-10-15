// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {NftAuction} from "./NftAuction.sol";

/// @title NftAuctionV2
/// @notice 升级演示。代理地址不变，V1 里已经创建的拍卖还在。
///         新状态 minIncrementBps 加在存储末尾。不要把新变量插进 NftAuction 原来的变量中间。
contract NftAuctionV2 is NftAuction {
    /// 新出价至少要比当前美元价格再高这么多万分比。0 时和 V1 一样，只要更高即可。
    uint256 public minIncrementBps;

    error IncrementTooHigh(uint256 bps);

    function version() public pure override returns (string memory) {
        return "v2";
    }

    /// @param bps 万分比。1000 表示新出价至少是当前价格的 110%。上限 20%，避免所有者把拍卖锁死。
    function setMinIncrementBps(uint256 bps) external onlyOwner {
        if (bps > 2_000) revert IncrementTooHigh(bps);
        minIncrementBps = bps;
    }

    function _isHigherBid(uint256 newUsd, uint256 currentUsd) internal view override returns (bool) {
        if (currentUsd == 0) return newUsd > 0;
        uint256 required = currentUsd + (currentUsd * minIncrementBps) / 10_000;
        return newUsd >= required && newUsd > currentUsd;
    }
}
