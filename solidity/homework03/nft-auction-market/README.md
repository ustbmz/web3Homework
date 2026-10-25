# NFT 拍卖市场

用 Hardhat 3 写的拍卖市场。卖家把 ERC721 交给拍卖合约，买家可以用 ETH 或 ERC20 出价。合约用 Chainlink 价格源把两种出价都换成美元再比较。拍卖逻辑放在 UUPS 代理后面，换实现合约时地址和已有拍卖都不变。

题目里写的 `@nomiclabs/hardhat-ethers` 和 `hardhat-deploy` 是 Hardhat 2 的包。这个仓库用的是 Hardhat 3，所以测试用 Mocha + ethers，部署用 `scripts/deploy.ts` 加上 OpenZeppelin 的 ERC1967 代理。`@chainlink/contracts` 固定在 `0.8.0`：更新的版本会去拉一个装不上的 git 依赖，而拍卖合约实际只用到其中的 `AggregatorV3Interface`。

## 项目结构

```
contracts/AuctionNFT.sol            ERC721，负责铸造
contracts/NftAuction.sol            拍卖逻辑，V1
contracts/NftAuctionV2.sol          升级后的逻辑，只在末尾加了最小加价比例
contracts/AuctionProxy.sol          用户真正调用的代理
contracts/mocks/                    本地测试用的 ERC20 和 Chainlink 价格源
test/NftAuction.ts                  功能测试
scripts/deploy.ts                   本地或 Sepolia 部署
```

用户调用的是代理地址。授权 NFT、批准 ERC20，都批准这个代理，不是实现合约。

## 功能

角色有三个：卖家、出价人、所有者。所有者可以改价格源、改手续费接收地址、升级合约。

一场拍卖要记住的数据都在 `Auction` 结构体里：卖家、NFT、截止时间、当前最高出价人、出价用的代币、原始数量、换算后的美元。`priceFeeds` 单独记每种代币的 Chainlink 价格源，`address(0)` 那一格是 ETH/USD。

三个会改状态的函数：

1. `createAuction`：卖家必须是这枚 NFT 的主人，并且已经 `approve` 给代理。合约把 NFT 转进来保管。
2. `bid`：ETH 出价时 `bidToken` 传 `address(0)`，`amount` 必须等于 `msg.value`。ERC20 出价时 `msg.value` 必须是 0，并提前 `approve`。美元价格更高才算覆盖上一笔，然后退回上一笔的原代币。
3. `endAuction`：截止后谁都可以调用。有赢家就把 NFT 转给赢家，扣除手续费后把钱给卖家；没人出价就把 NFT 还给卖家。

转账顺序是先改拍卖记录，再转 NFT 或退款。付款用的是这场拍卖记下的数量，不是合约的全部余额，所以两场拍卖的钱不会串。

## 价格

Chainlink 的 `latestRoundData` 返回整数价格，美元喂价一般是 8 位小数。本合约把结果统一成 18 位小数的美元：

```text
美元 = 数量 * 预言机价格 * 10^(18 - 价格小数位) / 10^(代币小数位)
```

ETH 按 18 位小数算。本地测试里 1 ETH = 2000 美元，所以 0.01 ETH = 20 美元。USDC 是 6 位小数、1 枚 = 1 美元，所以 30 USDC = 30 美元，可以压过 0.01 ETH。原始数量不能直接比：0.0001 ETH 的 wei 数量比 1 USDC 的最小单位大，但美元只有 0.2，1 USDC 仍会赢。

价格小于等于 0，或者超过 `maxPriceAge` 没有更新，新的出价会被拒绝。结算不再读预言机，用的是出价当时记下的美元，所以结算时价格过期也不会改赢家。

## 手续费

按出价当时的美元价值分档，从这笔出价的代币里扣：

| 美元金额 | 手续费 |
| --- | --- |
| 不足 100 | 1% |
| 100 以上、不足 1000 | 2% |
| 1000 以上 | 3% |

1 ETH 在 2000 美元的价格下收 3%，卖家得到 0.97 ETH，0.03 ETH 进 `feeRecipient`。

## 升级

用的是 UUPS。代理保存数据，`NftAuction` 保存逻辑，`upgradeToAndCall` 只有所有者能调用。实现合约的构造函数里调用了 `_disableInitializers`，所以不能直接对实现合约执行 `initialize`。

`NftAuctionV2` 只在存储末尾增加 `minIncrementBps`。设成 1000 时，已有出价之后的新出价至少要再高 10%。不要往 `NftAuction` 已有变量中间插入新变量，也不要给 `Auction` 结构体加字段，否则旧拍卖会读错。

## 测试

```bash
cd solidity/homework03/nft-auction-market
npm test
npx hardhat test --coverage
```

20 个测试全部通过。业务合约行覆盖率：`AuctionNFT` 100%，`NftAuction` 91.58%，`NftAuctionV2` 100%。完整数字在 `测试报告.md`。

## 部署

本地模拟网络会部署假的价格源和 USDC，用来确认脚本能跑完。这个网络在脚本结束时就消失了。

```bash
npm run deploy:local
```

Sepolia 使用 Chainlink 上已经存在的喂价，不部署假价格源。

| 用途 | 地址 |
| --- | --- |
| ETH/USD | `0x694AA1769357215DE4FAC081bf1f309aDC325306` |
| LINK/USD | `0xc59E3633BAAC79493d908e63626716e204A45EdF` |
| LINK 代币 | `0x779877A7B0D9E8603169DdbD7836e478b4624789` |

部署前到 [Chainlink 喂价地址页](https://docs.chain.link/data-feeds/price-feeds/addresses) 再核对一次。密钥放在 Hardhat keystore，不写进 `.env`：

```bash
npx hardhat keystore set SEPOLIA_RPC_URL
npx hardhat keystore set SEPOLIA_PRIVATE_KEY
npm run deploy:sepolia
```

成功后地址写在 `deployments/sepolia.json`。用户交互和授权都用里面的 `auctionProxy`。
