// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Initializable} from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts/proxy/utils/UUPSUpgradeable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {IERC721} from "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import {IERC721Receiver} from "@openzeppelin/contracts/token/ERC721/IERC721Receiver.sol";
import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";

/// @title NftAuction
/// @notice 用美元比较 ETH 和 ERC20 出价的 NFT 拍卖。用户调用的是代理地址，本合约是逻辑实现。
///
/// 需求先拆成这张表，再写函数：
///
/// | 要填的 | 这个合约 |
/// | 角色 | 卖家、出价人、所有者（改价格源、升级、收手续费） |
/// | 要记住的数据 | 每一场拍卖的结构体；token => Chainlink 价格源 |
/// | 动作 | createAuction、bid、endAuction、quoteUsd |
/// | 检查 | 是不是 NFT 主人、拍卖是否还在进行、美元价格是否更高、价格是否过期 |
/// | 转账顺序 | 先改拍卖记录，再转 NFT / 退款 / 付款 |
///
/// 存储：OpenZeppelin v5 的 owner、初始化标记、重入锁都在独立槽位里。
/// 下面这些变量从 0 号槽开始。升级时只能在子合约末尾追加变量，不能插入，也不能给 Auction 加字段。
contract NftAuction is Initializable, OwnableUpgradeable, ReentrancyGuard, UUPSUpgradeable, IERC721Receiver {
    using SafeERC20 for IERC20;

    /// 美元价格统一换成 18 位小数后再比较。这里的 1e18 表示 1 美元，不是 1 ETH。
    uint256 public constant USD_UNIT = 1e18;
    uint256 public constant TIER_LOW = 100 * USD_UNIT;
    uint256 public constant TIER_MID = 1000 * USD_UNIT;
    /// 拍卖最短 60 秒，避免截止时间就是当前区块、别人来不及出价。
    uint256 public constant MIN_DURATION = 60;

    struct Auction {
        address seller;
        address nftContract;
        uint256 tokenId;
        uint256 startTime;
        uint256 endTime;
        bool settled;
        address highestBidder;
        uint256 highestBidAmount;
        address bidToken; // address(0) 表示这一笔是 ETH
        uint256 highestBidUsd; // 18 位小数的美元
    }

    /// 下一场拍卖的编号。从 1 开始，0 表示“这场不存在”。
    uint256 public nextAuctionId;
    mapping(uint256 auctionId => Auction) public auctions;
    /// 出价代币 => Chainlink Aggregator。address(0) 这一格存的是 ETH/USD。
    mapping(address token => address feed) public priceFeeds;
    address public feeRecipient;
    /// 价格超过这么多秒没更新，就拒绝新的出价。结算不用重新读价格。
    uint256 public maxPriceAge;

    error ZeroAddress();
    error ZeroAmount();
    error DurationTooShort(uint256 duration);
    error NotNftOwner(address caller, uint256 tokenId);
    error SellerCannotBid();
    error AuctionNotFound(uint256 auctionId);
    error AuctionAlreadySettled(uint256 auctionId);
    error AuctionStillRunning(uint256 auctionId);
    error AuctionFinished(uint256 auctionId);
    error EthAmountMismatch(uint256 value, uint256 amount);
    error UnexpectedEth();
    error PriceFeedMissing(address token);
    error InvalidPrice(address token);
    error StalePrice(address token, uint256 updatedAt);
    error BidTooLow(uint256 newUsd, uint256 currentUsd);
    error PayoutFailed(address to);

    event AuctionCreated(
        uint256 indexed auctionId,
        address indexed seller,
        address indexed nftContract,
        uint256 tokenId,
        uint256 endTime
    );
    event BidPlaced(
        uint256 indexed auctionId,
        address indexed bidder,
        address bidToken,
        uint256 amount,
        uint256 usdAmount
    );
    event AuctionEnded(uint256 indexed auctionId, address indexed winner, uint256 proceeds, uint256 fee);
    event PriceFeedSet(address indexed token, address indexed feed);

    /// 实现合约自己的构造函数。这里的存储写在实现合约上，不写进代理。
    /// 目的是锁死实现合约，别人不能对实现合约调用 initialize 把自己设成 owner。
    /// 代理的存储是另一份，所以代理仍然可以初始化。
    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /// 代替构造函数。只能通过代理调用一次。
    /// @param ethUsdFeed ETH/USD 的 Chainlink 价格源。ETH 在本合约里用 address(0) 表示。
    function initialize(address initialOwner, address ethUsdFeed, address feeRecipient_, uint256 maxPriceAge_)
        external
        initializer
    {
        if (initialOwner == address(0) || ethUsdFeed == address(0) || feeRecipient_ == address(0)) {
            revert ZeroAddress();
        }
        if (maxPriceAge_ == 0) revert ZeroAmount();

        __Ownable_init(initialOwner);
        feeRecipient = feeRecipient_;
        maxPriceAge = maxPriceAge_;
        nextAuctionId = 1;
        priceFeeds[address(0)] = ethUsdFeed;
        emit PriceFeedSet(address(0), ethUsdFeed);
    }

    function version() public pure virtual returns (string memory) {
        return "v1";
    }

    /// 设置某种代币的美元价格源。token 传 address(0) 就是在改 ETH/USD。
    function setPriceFeed(address token, address feed) external onlyOwner {
        if (feed == address(0)) revert ZeroAddress();
        priceFeeds[token] = feed;
        emit PriceFeedSet(token, feed);
    }

    function setFeeRecipient(address recipient) external onlyOwner {
        if (recipient == address(0)) revert ZeroAddress();
        feeRecipient = recipient;
    }

    function setMaxPriceAge(uint256 maxPriceAge_) external onlyOwner {
        if (maxPriceAge_ == 0) revert ZeroAmount();
        maxPriceAge = maxPriceAge_;
    }

    /// 卖家把 NFT 交给拍卖合约保管。调用前必须 approve 这个代理地址，不是实现合约地址。
    function createAuction(address nftContract, uint256 tokenId, uint256 duration)
        external
        nonReentrant
        returns (uint256 auctionId)
    {
        if (nftContract == address(0)) revert ZeroAddress();
        if (duration < MIN_DURATION) revert DurationTooShort(duration);
        if (IERC721(nftContract).ownerOf(tokenId) != msg.sender) revert NotNftOwner(msg.sender, tokenId);

        auctionId = nextAuctionId++;
        Auction storage auction = auctions[auctionId];
        auction.seller = msg.sender;
        auction.nftContract = nftContract;
        auction.tokenId = tokenId;
        auction.startTime = block.timestamp;
        auction.endTime = block.timestamp + duration;

        // 先写记录，再把 NFT 拉进来。对方若在回调里再次进入，重入锁会挡住。
        IERC721(nftContract).safeTransferFrom(msg.sender, address(this), tokenId);
        emit AuctionCreated(auctionId, msg.sender, nftContract, tokenId, auction.endTime);
    }

    /// 用 ETH 或 ERC20 出价。两种资产的原始数量不能直接比，先换成美元。
    /// ETH：bidToken 传 address(0)，amount 必须等于 msg.value。
    /// ERC20：msg.value 必须是 0，并且调用者已经 approve 足够的数量。
    function bid(uint256 auctionId, address bidToken, uint256 amount) external payable nonReentrant {
        Auction storage auction = _openAuction(auctionId);
        if (msg.sender == auction.seller) revert SellerCannotBid();

        if (bidToken == address(0)) {
            if (amount == 0 || msg.value != amount) revert EthAmountMismatch(msg.value, amount);
        } else {
            if (msg.value != 0) revert UnexpectedEth();
            if (amount == 0) revert ZeroAmount();
        }

        uint256 usdAmount = quoteUsd(bidToken, amount);
        if (usdAmount == 0 || !_isHigherBid(usdAmount, auction.highestBidUsd)) {
            revert BidTooLow(usdAmount, auction.highestBidUsd);
        }

        address previousBidder = auction.highestBidder;
        address previousToken = auction.bidToken;
        uint256 previousAmount = auction.highestBidAmount;

        auction.highestBidder = msg.sender;
        auction.bidToken = bidToken;
        auction.highestBidAmount = amount;
        auction.highestBidUsd = usdAmount;

        if (bidToken != address(0)) {
            IERC20(bidToken).safeTransferFrom(msg.sender, address(this), amount);
        }
        // 退的是上一笔的代币和数量，不是合约里的全部余额。同时进行的其他拍卖不受影响。
        if (previousBidder != address(0)) {
            _payout(previousToken, previousBidder, previousAmount);
        }

        emit BidPlaced(auctionId, msg.sender, bidToken, amount, usdAmount);
    }

    /// 到时间后谁都可以结算，避免卖家不上线导致 NFT 和钱一起被卡住。
    /// 手续费按出价当时记下的美元价格计算，不再读预言机，所以结算时价格过期也不改变赢家。
    function endAuction(uint256 auctionId) external nonReentrant {
        Auction storage auction = _existingAuction(auctionId);
        if (auction.settled) revert AuctionAlreadySettled(auctionId);
        if (block.timestamp < auction.endTime) revert AuctionStillRunning(auctionId);

        auction.settled = true;

        address winner = auction.highestBidder;
        if (winner == address(0)) {
            IERC721(auction.nftContract).safeTransferFrom(address(this), auction.seller, auction.tokenId);
            emit AuctionEnded(auctionId, address(0), 0, 0);
            return;
        }

        uint256 fee = (auction.highestBidAmount * feeBps(auction.highestBidUsd)) / 10_000;
        uint256 proceeds = auction.highestBidAmount - fee;

        IERC721(auction.nftContract).safeTransferFrom(address(this), winner, auction.tokenId);
        _payout(auction.bidToken, auction.seller, proceeds);
        _payout(auction.bidToken, feeRecipient, fee);
        emit AuctionEnded(auctionId, winner, proceeds, fee);
    }

    /// 把代币数量换成 18 位小数的美元。
    /// usd = 数量 * 预言机价格 * 10^(18 - 价格小数位) / 10^(代币小数位)
    /// ETH 的小数位按 18 算。Chainlink 的美元价格通常是 8 位小数。
    function quoteUsd(address token, uint256 amount) public view returns (uint256) {
        address feed = priceFeeds[token];
        if (feed == address(0)) revert PriceFeedMissing(token);

        AggregatorV3Interface oracle = AggregatorV3Interface(feed);
        (uint80 roundId, int256 answer,, uint256 updatedAt, uint80 answeredInRound) = oracle.latestRoundData();
        if (answer <= 0 || answeredInRound < roundId) revert InvalidPrice(token);
        if (updatedAt == 0 || updatedAt > block.timestamp || block.timestamp - updatedAt > maxPriceAge) {
            revert StalePrice(token, updatedAt);
        }

        uint8 priceDecimals = oracle.decimals();
        if (priceDecimals > 18) revert InvalidPrice(token);

        uint8 tokenDecimals = 18;
        if (token != address(0)) {
            tokenDecimals = IERC20Metadata(token).decimals();
            if (tokenDecimals > 18) revert InvalidPrice(token);
        }

        return (amount * uint256(answer) * (10 ** (18 - priceDecimals))) / (10 ** tokenDecimals);
    }

    /// 按成交额的美元价值取手续费万分比。100 = 1%，200 = 2%，300 = 3%。
    /// 不足 100 美元收 1%，100 美元（含）到 1000 美元收 2%，1000 美元（含）以上收 3%。
    function feeBps(uint256 usdAmount) public pure returns (uint256) {
        if (usdAmount < TIER_LOW) return 100;
        if (usdAmount < TIER_MID) return 200;
        return 300;
    }

    /// 子合约可以改“怎样才算更高的出价”。V1 只要美元价格严格更高。
    function _isHigherBid(uint256 newUsd, uint256 currentUsd) internal view virtual returns (bool) {
        return newUsd > currentUsd;
    }

    function _openAuction(uint256 auctionId) internal view returns (Auction storage auction) {
        auction = _existingAuction(auctionId);
        if (auction.settled) revert AuctionAlreadySettled(auctionId);
        if (block.timestamp >= auction.endTime) revert AuctionFinished(auctionId);
    }

    function _existingAuction(uint256 auctionId) internal view returns (Auction storage auction) {
        auction = auctions[auctionId];
        if (auction.seller == address(0)) revert AuctionNotFound(auctionId);
    }

    /// ETH 用 call，并把返回值检查掉。transfer 只转发 2300 gas，收款方是合约时可能失败。
    function _payout(address token, address to, uint256 amount) internal {
        if (amount == 0) return;
        if (token == address(0)) {
            (bool ok,) = to.call{value: amount}("");
            if (!ok) revert PayoutFailed(to);
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
    }

    /// 收到 safeTransferFrom 的 NFT 时必须返回这个选择器，否则转账会回滚。
    function onERC721Received(address, address, uint256, bytes calldata) external pure returns (bytes4) {
        return IERC721Receiver.onERC721Received.selector;
    }

    /// UUPS 把升级函数放在实现合约里。只有 owner 能换实现。
    function _authorizeUpgrade(address) internal override onlyOwner {}
}
