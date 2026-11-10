// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";

/// @title MockV3Aggregator
/// @notice 本地测试用的 Chainlink 价格源。线上用的是 Chainlink 已经部署好的代理地址，
///         这个合约只实现同样的 latestRoundData / decimals，让拍卖合约不用改代码就能在本地跑。
///         freeze 为 false 时，更新时间永远是当前区块，避免测试里拨快时间后被误判为过期。
contract MockV3Aggregator is AggregatorV3Interface {
    uint8 private immutable _decimals;
    int256 private _answer;
    uint256 private _updatedAt;
    bool public freeze;

    constructor(uint8 decimals_, int256 initialAnswer) {
        _decimals = decimals_;
        _answer = initialAnswer;
        _updatedAt = block.timestamp;
    }

    function decimals() external view override returns (uint8) {
        return _decimals;
    }

    function setAnswer(int256 nextAnswer) external {
        _answer = nextAnswer;
        if (!freeze) _updatedAt = block.timestamp;
    }

    /// 把价格时间冻在过去，用来测试“预言机数据过期就拒绝出价”。
    function setUpdatedAt(uint256 timestamp) external {
        freeze = true;
        _updatedAt = timestamp;
    }

    function description() external pure override returns (string memory) {
        return "Mock V3 Aggregator";
    }

    function version() external pure override returns (uint256) {
        return 0;
    }

    function getRoundData(uint80)
        external
        view
        override
        returns (uint80, int256, uint256, uint256, uint80)
    {
        return _round();
    }

    function latestRoundData()
        external
        view
        override
        returns (uint80, int256, uint256, uint256, uint80)
    {
        return _round();
    }

    function _round() private view returns (uint80, int256, uint256, uint256, uint80) {
        uint256 ts = freeze ? _updatedAt : block.timestamp;
        return (1, _answer, ts, ts, 1);
    }
}
