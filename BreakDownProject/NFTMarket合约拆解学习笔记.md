# NFTMarket 合约拆解学习笔记

> 范围：只用学 **智能合约**。前端、后端不在本笔记里。  
> 代码目录：`EasySwapContract/contracts/`  
> 用法：按课时往下走。每课先写「打开哪个文件」，再写「看什么、想什么」，最后有过关题。过关再进下一课。

---

## 怎么用

1. 左右开：左边笔记，右边对应 `.sol`。
2. 每读一个函数只问三句：**谁能调、资产去哪、订单在盘口里还是已经关掉。**
3. 贯穿数字例子（建议抄在纸上）：

```
NFT：BAYC #42（示例，任意 ERC721 都行）
卖家 S 挂单价：1 ETH
买家 B 出价：1.2 ETH（单品 Bid）
协议费 protocolShare：250  （2.5%，因为 TOTAL_SHARE = 10000）
List.amount 必须 = 1
expiry = 0 表示永不过期
salt ≠ 0
```

4. 学习以 `contracts/` 下正式合约为准。`contracts/test/` 只在跑测试时看。

回看时先记住这四句：NFT 授权给 Vault，不是 OrderBook。Buy Now 不必先 `make` Bid，钱从 `msg.value` 出。字段不合格的单返回 `0x0`，其它单可以留下；没授权或 ETH 不够会整笔回滚。成交价用被吃掉的那一边。

---

## 学习地图（先看一眼）

```
EasySwapContract/contracts/
├── EasySwapOrderBook.sol     ★ 对外入口，撮合都在这里
├── EasySwapVault.sol         托管 ETH / NFT
├── OrderStorage.sol          价格红黑树 + 同价链表
├── OrderValidator.sol        filledAmount、取消标记
├── ProtocolManager.sol       协议费率
├── libraries/
│   ├── LibOrder.sol          订单结构与哈希
│   ├── LibPayInfo.sol        10000 = 100%，上限 10%
│   ├── RedBlackTreeLibrary.sol
│   └── LibTransferSafeUpgradeable.sol
└── interface/                用到再点
```

继承关系（OrderBook 是拼出来的）：

```
EasySwapOrderBook
  ├─ OrderStorage      盘口怎么存
  ├─ OrderValidator    能不能成交
  ├─ ProtocolManager   抽多少费
  └─ 指向 EasySwapVault 资产实际放哪
```

---

## 第 0 课：先搞懂产品（不看代码）

**目标：** 能讲清「链上订单簿 NFT 市场」和 OpenSea 差在哪。用时约 20 分钟。

OpenSea 一类：订单在链下签名，上架几乎不花 gas，有人买了才上链。  
EasySwap：订单写进合约，上架就要 gas，同时把 NFT 或 ETH **锁进金库**。好处是盘口可以在链上按价格排序查询，不必信中继器。

两个方向：

| side | 谁发起 | 锁什么 | 等人来干什么 |
| --- | --- | --- | --- |
| List（挂单） | 卖家 | NFT | 买家来 match，付 ETH |
| Bid（出价） | 买家 | ETH | 卖家来 match，交 NFT |

`List` 是 listing，把这一件挂出去卖。它不是「列表」。

两种标的：

- `FixedPriceForItem`：指定 collection + tokenId
- `FixedPriceForCollection`：只出现在 Bid，「这个合集任意一个 NFT」

三个必须先接受的设计：

1. **资产先托管，后成交。** 挂着的 NFT 不在卖家钱包里。
2. **改价 = 取消旧单 + 下一张新单。** 因为 OrderKey 是整单哈希。
3. **合约不会自动对敲。** 必须有人调用 `matchOrder(sell, buy)`。

### 这套合约管什么

EasySwap 只管用链上订单簿买卖 ERC721。卖家挂单时把某一件 NFT 锁进 `EasySwapVault`，买家出价时把 ETH 锁进同一座金库，然后必须有一方来调 `matchOrder`。合约同时完成 NFT 过户、ETH 结算和协议抽成。

它覆盖的是固定价、ETH 计价、一个合集一份盘口。卖单只能指定某一件，`amount` 必须为 1。买单可以是这一件，也可以是「这个合集里任意一件」。盘口按价格树加同价队列排：List 取最低卖价，Bid 取最高买价，同一价先挂的先成交。

合约不管这些：ERC20 报价、拍卖、一次打包买多件不同的 NFT；也不会自动把最优买卖价对上。没人调用 `matchOrder`，两张单就一直挂着。前端列表和地板价来自链下把 `LogMake` / `LogCancel` / `LogMatch` 扫进数据库，不在本笔记里。

### 为什么这课重要

这是课程里第一次把订单簿走通。Pledge 没有盘口，只是一个池子到期一次性配对。这里第一次出现三件后面还会用的东西：订单怎么存、资产先托管再成交、成交价取被吃那一边的报价。

后面的对照就靠这套：

- 永续的订单在链下签名，链上只结算。NFT 的订单写在合约的价格树里。
- 撮合引擎才把价格优先、时间优先、Maker 价一次做完。NFT 这边先看见同样的规则，只是簿在链上，而且要有人来 `match`。

### 读合约前只要记住的 ERC721

不用先读完 OpenZeppelin 的整份 ERC721。这条路径只用到下面几条：

- 一个合集是一份 ERC721 合约。每个 `tokenId` 只有一件，所以卖单 `amount` 必须是 1。
- 挂单前卖家 `approve` 或 `setApprovalForAll` 给 **Vault**。`depositNFT` 内部是 `safeTransferFrom(卖家, Vault, tokenId)`。
- Vault 实现了 `onERC721Received`。`safeTransferFrom` 转给合约时，接收方必须有这个函数。
- 挂单成功后 `ownerOf(tokenId)` 是 Vault，成交后变成买家。

### libraries 四份各自干什么

四份都是工具，没有自己的资产账本。

| 文件 | 干什么 |
| --- | --- |
| `LibOrder.sol` | 订单结构、`hash` 出 OrderKey、`ORDERKEY_SENTINEL = 0x0` |
| `LibPayInfo.sol` | `TOTAL_SHARE = 10000` 表示 100%，费率上限 10%。同文件的 `PayInfo` 成交路径用不到 |
| `RedBlackTreeLibrary.sol` | 价格红黑树。只用 `insert` / `remove` / `first` / `last` / `next` / `prev`，旋转不用逐行读 |
| `LibTransferSafeUpgradeable.sol` | 把 ETH 和 NFT 转出去并检查是否成功。里面的 ERC20 转账，这条主线不用 |

**过关：** 卖家刚 `makeOrders` 成功，BAYC #42 现在在谁的地址上？买家「立即购买」一定要先自己挂一个 Bid 吗？

---

## 第 1 课：订单长什么样 —— LibOrder

**目标：** 能默写 Order 字段，并知道 OrderKey 怎么来。用时约 25 分钟。

打开 `libraries/LibOrder.sol`。

盯三个东西：

1. `enum Side { List, Bid }`。`List` = 挂单，卖家锁 NFT。`Bid` = 出价，买家锁 ETH。
2. `enum SaleKind`。`FixedPriceForItem` 指定某一件。`FixedPriceForCollection` 只出现在买单，合集里任意一件，成交时 tokenId 以卖单为准。
3. `struct Order` 和 `function hash(Order)`。

贯穿例子的卖单：

```
side      = List
saleKind  = FixedPriceForItem
maker     = S
nft       = { tokenId: 42, collection: BAYC, amount: 1 }
price     = 1e18          // 单价，单位 wei。1 ETH
expiry    = 0             // 0 = 永不过期
salt      = 12345         // 不能为 0。同样的价再挂一单，靠它得到不同的 key
```

`OrderKey` 怎么来。笔记里以前写的 `TYPEHASH`，就是本文件手写的常量 `ORDER_TYPEHASH`。它不是用户传入的，也不是编译器从 `struct Order` 自动生成的。

作者按字段类型写了一句字符串，编译时对它做 `keccak256`，结果固定进字节码：

```
ORDER_TYPEHASH = keccak256(
  "Order(uint8 side,uint8 saleKind,address maker,Asset nft,uint128 price,uint64 expiry,uint64 salt)"
  "Asset(uint256 tokenId,address collection,uint96 amount)"
)
```

这句字符串是开发人员自己写的，用来描述本项目的 `Order` 和 `Asset`。编译器不会从 `struct` 生成它，也不检查它和结构体是否一致。

`uint8`、`address`、`uint256`、`uint128`、`uint64`、`uint96` 是 Solidity 编码用的标准类型。`Side`、`SaleKind` 写成 `uint8`，`Price` 写成 `uint128`，因为链上按这个宽度编码。`Order`、`Asset`、`side`、`tokenId` 这些名字是项目自己起的。

这个写法来自 EIP-712：把「结构体名 + 字段名 + 字段类型」拼成固定字符串，再哈希一次，当作类型指纹。这里用它当订单号的前缀，把字段顺序和类型钉死。以后改了 `struct Order` 的字段，这句字符串和 `hash(Order)` 里的拼接顺序要一起改，否则同一张单会算出另一把 `OrderKey`。

`hash(Order)` 再做一次哈希：最前面放这个固定的 `ORDER_TYPEHASH`，后面依次拼这张单的 `side`、`saleKind`、`maker`、`hash(nft)`、`price`、`expiry`、`salt`。`nft` 不直接拼三个字段，先用另一个常量哈希：

```
ASSET_TYPEHASH = keccak256("Asset(uint256 tokenId,address collection,uint96 amount)")
hash(nft) = keccak256(ASSET_TYPEHASH, tokenId, collection, amount)
```

所以改 `price` 或 `salt`，变的是拼进去的值，`ORDER_TYPEHASH` 不变，`OrderKey` 变了。`editOrders` 因此是先取消旧单、再挂一张新单，旧 key 不能再 match。

`hash` 不算 `filledAmount`。成交了多少、取消没有，记在 `OrderValidator.filledAmount`。

`ORDERKEY_SENTINEL = 0x0` 表示「这一格是空的」。它不是 `revert` 的报错字符串。`makeOrders` 某一张没做成，返回数组对应位置就是它，整笔交易不一定 revert。盘口链表里也用同一个值：`head` 是 `0x0` 表示这一档还没有单，`next` 是 `0x0` 表示后面没有下一张。

返回 `0x0` 时，这张单哪儿都没去。`amount != 1` 发生在 `depositNFT` 和 `_addOrder` 之前，NFT 还在卖家钱包里，盘口和金库都没有它。链上只留下返回数组里的这个 `0x0`。挂成功才会 `_addOrder`，同时把 NFT 或 ETH 锁进金库。

List 的 `amount` 必须是 1，是因为 `makeOrders` 是公开函数。页面可以写死传 1，别人仍能直接发一笔 `amount = 2` 的交易。ERC721 一个 `tokenId` 只有一件，`depositNFT` 也只转这一件。`amount` 写成 2，金库里仍是一件，盘口却按「卖两件」记账，后面的 `filledAmount < amount` 会对不上。所以不等于 1 就返回 `0x0`，不进金库。

```solidity
if (order.side == LibOrder.Side.List) {
    if (order.nft.amount != 1) {
        return LibOrder.ORDERKEY_SENTINEL;
    }
    IEasySwapVault(_vault).depositNFT(newOrderKey, order.maker, order.nft.collection, order.nft.tokenId);
}
```

旁边 `amount == 0` 的判断在 `side == Bid` 的分支：买单件数不能是 0，可以大于 1。那是集合出价。

**过关：** List 的 `amount` 为什么必须是 1？`hash` 里有没有 `filledAmount`？（没有。成交进度在 `OrderValidator.filledAmount`。）

---

## 第 2 课：金库 —— 资产实际住在哪

**目标：** 搞清 Vault 只认 OrderBook，以及按 OrderKey 记账。用时约 30 分钟。

打开 `EasySwapVault.sol`。先看修饰器 `onlyEasySwapOrderBook`，再按「存 / 取 / 改」读。

| 函数 | 方法说明 | 何时发生 | 记账 |
| --- | --- | --- | --- |
| `depositNFT(orderKey, from, collection, tokenId)` | 把 `from` 手里的这一件转进 Vault，记在这张卖单下。`collection` 是 ERC721 合约地址 | 挂 List | `NFTBalance[orderKey] = tokenId` |
| `withdrawNFT(orderKey, to, collection, tokenId)` | 金库里记下的 tokenId 必须和参数一致，然后把这一件转给 `to` | 取消卖单，或卖单成交给买家 | 先 `require(NFTBalance[orderKey] == tokenId)`，再 `delete` |
| `depositETH(orderKey, ETHAmount)` | 调用要带上 ETH。`ETHAmount` 只检查「至少带了这么多」，余额加的是 `msg.value` | 挂 Bid。OrderBook 传入的 value 是 `单价 × amount` | `ETHBalance[orderKey] += msg.value` |
| `withdrawETH(orderKey, ETHAmount, to)` | 从这张单的余额里扣 `ETHAmount`，ETH 转给 `to` | 取消买单退剩余 ETH，或成交时把 ETH 抽到 OrderBook | `ETHBalance[orderKey] -= ETHAmount` |
| `editNFT(oldOrderKey, newOrderKey)` | 改价后 OrderKey 变了，把旧单记下的 tokenId 改记到新单。NFT 仍在 Vault | 卖家改 List 的价格或数量 | `NFTBalance[newOrderKey] = NFTBalance[oldOrderKey]`，再删旧 key |
| `editETH(oldOrderKey, newOrderKey, oldETHAmount, newETHAmount, to)` | 旧单余额清零，按新旧金额多退少补，记到新单。`to` 是买单的 maker | 买家改 Bid | 更低：新 key 记 `newETHAmount`，差额退 `to`。更高：新 key 记 `msg.value + oldETHAmount`。相等：新 key 记 `oldETHAmount` |
| `transferERC721(from, to, assets)` | Vault 代 `from` 把 `assets` 里的那一件直接转给 `to`，不写 `NFTBalance`。卖家要事先 approve 给 Vault | 卖家接受 Bid，但自己的卖单不在盘口 | 不记账 |





表里这些函数都有 `onlyEasySwapOrderBook`，用户直接调会失败。`batchTransferERC721` 没有这道限制，用户可以自己批量转 NFT，主线不用看它。

`setOrderBook` 只有 Owner 能调。部署顺序：先 `Vault.initialize`，再 `OrderBook.initialize(vault)`，再 `Vault.setOrderBook(orderBook)`。顺序反了，OrderBook 调 Vault 会被 `onlyEasySwapOrderBook` 拒绝。

授权提醒：ERC721 要 `approve` **Vault 地址**，不是 OrderBook。`depositNFT` 内部是 `safeTransferFrom(from, Vault, tokenId)`。

`depositETH` 是挂买单时把 ETH 锁进金库。`ETHAmount` 只检查「至少带了这么多」，真正加上去的是这次调用的 `msg.value`。OrderBook 的调用是 `depositETH{value: 单价 × amount}(orderKey, 单价 × amount)`，所以余额增加的就是这笔出价。你钱包里多付的 ETH 不在这个函数里退，由 `makeOrders` 在所有单处理完之后退差额。

**过关：** 买家出价 1.2 ETH、amount=1，Vault 里这个 OrderKey 的 ETHBalance 是多少？取消时退多少？（1.2 ETH；还没成交就取消，退 1.2 ETH。）

---

## 第 3 课：校验与成交状态 —— OrderValidator / ProtocolManager

**目标：** 知道一笔单怎样算「还能成交」。用时约 20 分钟。

打开 `OrderValidator.sol`。

`filledAmount[orderKey]` 记这张单成交了几件。它不在 `Order` 里，也不参与 `OrderKey`。

```
filledAmount[orderKey]
  0                 未成交。新挂出来的单就是 0
  1 .. amount-1     部分成交。卖单 amount 只能是 1，不会停在这里。只有集合 Bid 每配成一件 +1
  amount            完全成交。卖单成交后写成 1。之后 filledAmount < amount 过不了
  type(uint256).max 已取消。常量 CANCELLED。_getFilledAmount 看到它会 revert "OVa: canceled"
```

取消过的这把 key 不能复用。`makeOrders` 要求 `filledAmount[hash] == 0`。

`_validateOrder` 只检查字段合不合法。不转资产，也不看成交了多少。任意一条 `require` 失败，整笔回滚。

```solidity
function _validateOrder(LibOrder.Order memory order, bool isSkipExpiry) internal view {
    require(order.maker != address(0), "OVa: miss maker");
    if (!isSkipExpiry) {
        require(order.expiry == 0 || order.expiry > block.timestamp, "OVa: expired");
    }
    require(order.salt != 0, "OVa: zero salt");
    if (order.side == LibOrder.Side.List) {
        require(order.nft.collection != address(0), "OVa: unsupported nft asset");
    } else if (order.side == LibOrder.Side.Bid) {
        require(Price.unwrap(order.price) > 0, "OVa: zero price");
    }
}
```

卖单这里不查价格，买单这里不查 `collection`。`amount` 和 `filledAmount` 也不在这个函数里。

`isSkipExpiry` 只在撮合时传入。某一边已经在盘口上，校验这一边时传 `true`，过期了也允许继续配。还没进盘口的那一边，以及对手方从盘口里读出来的那张单，都传 `false`。函数注释里写的「取消时跳过」和调用点不一致：取消走 `_cancelOrderTry`，并不调用 `_validateOrder`。

再打开 `ProtocolManager.sol` + `LibPayInfo.sol`。费率不写进 `Order`。成交时 `_shareToAmount` 读的是 `protocolShare` 的当前值，Owner 可以事后用 `setProtocolShare` 改，上限 1000。

```
fee = fillPrice × protocolShare / TOTAL_SHARE
TOTAL_SHARE = 10000   // 100%
protocolShare = 250   // 2.5%
上限 MAX_PROTOCOL_SHARE = 1000   // 10%
```

贯穿例子：成交价 1 ETH → 手续费 0.025 ETH，卖家到账 0.975 ETH。费留在 OrderBook，成交当时不打给 Owner。

`LibPayInfo` 里还有 `PayInfo` 和它自己的 `TYPE_HASH`。成交路径只用 `TOTAL_SHARE` 和 `MAX_PROTOCOL_SHARE`，`PayInfo` 跳过。

**过关：** 已取消的单再 match，会在哪一层被拦住？`_getFilledAmount` 里那句 `require != CANCELLED`。

---

## 第 4 课：盘口怎么存 —— OrderStorage

**目标：** 能在纸上画出「同一合集三个卖价」的树和链表。用时约 45 分钟。这是本项目最「订单簿」的一课。

打开 `OrderStorage.sol`。先看三个 mapping，再读 `_addOrder` / `_removeOrder`，最后 `getBestPrice`。`RedBlackTreeLibrary.sol` 的旋转不用逐行读，这里只用 `insert`、`remove`、`first`（最小价）、`last`（最大价）、`next` / `prev`（下一档）。

结构：

```
priceTrees[collection][side]                 红黑树，节点是价格
orderQueues[collection][side][price]         该价格档的链表头尾
orders[orderKey] = { order, next }           节点
```

这个合约不锁 NFT、不收 ETH。`EasySwapOrderBook` 继承它，挂单成功后调 `_addOrder`，取消或成交后调 `_removeOrder`。

### getBestPrice / getNextBestPrice

卖单取 `first()`，树上最小的卖价。买单取 `last()`，树上最大的买价。这一边还没有单时，返回的价格是空。

```solidity
function getBestPrice(address collection, LibOrder.Side side) public view returns (Price price) {
    price = (side == LibOrder.Side.Bid)
        ? priceTrees[collection][side].last()
        : priceTrees[collection][side].first();
}
```

`getNextBestPrice` 从当前这一档走到下一档更差的价格。买单用 `prev`，价格变小。卖单用 `next`，价格变大。传入的价格是空时，退回最优价。`getOrders` 翻页用它。

卖价 1.5、1.0、2.0 三档，`getBestPrice(collection, List)` 是 **1.0**。同一 1.0 档有两单，先成交 `head`，也就是先挂上的那张。后挂的接在 `tail`。

### _addOrder

先 `hash` 出 OrderKey。这个 key 已经有 `maker`，说明同一张单挂过，`CannotInsertDuplicateOrder`。这个价第一次出现就 `tree.insert`。队列空则 `head = tail = 本单`，否则旧 `tail.next` 指向新单，再把 `tail` 改成新单。新单自己的 `next` 是 `0x0`。

```solidity
orderKey = LibOrder.hash(order);
if (orders[orderKey].order.maker != address(0)) {
    revert CannotInsertDuplicateOrder(orderKey);
}
if (!priceTree.exists(order.price)) {
    priceTree.insert(order.price);
}
if (LibOrder.isSentinel(orderQueue.tail)) {
    orderQueue.head = orderKey;
    orderQueue.tail = orderKey;
} else {
    orders[orderQueue.tail].next = orderKey;
    orderQueue.tail = orderKey;
}
```

### _removeOrder

已经按 `collection + side + price` 找到那一档，从 `head` 往后对 `maker`、`saleKind`、`expiry`、`salt`、`tokenId`、`amount`。价格和方向不用再比。摘的是 `head` 就让 `head` 变成它的 `next`；摘的是 `tail` 就让 `tail` 退回前一张。然后 `delete orders[这张]`。整档空了，从红黑树上删掉这个价格。找不到就 `revert "Cannot remove missing order"`。

### getOrders / getBestOrder

两个都只读。`getOrders` 给前端翻页：`price` 传空就从最优价开始，`count` 是这一页最多几张，`firstOrderKey` 是上一页返回的下一张。`getBestOrder` 从最优价往下找第一张现在还能用的单。

都会跳过：

- 已过期（`expiry != 0` 且 `< now`）
- 查单品 Bid 时 `tokenId` 不符的单
- 查集合 Bid 时，把单品 Bid 滤掉

卖单侧没有后两道过滤。`getOrders` 返回这一页订单，以及下一页要从哪张 `OrderKey` 开始。数组长度固定是 `count`，不够的位置是空单。

**过关：** 合集里有卖价 1.5、1.0、2.0 三档，`getBestPrice(collection, List)` 是哪个？同一 1.0 档有两单，先成交哪一单？（1.0；先挂的那张，队列的 `head`。）

---

## 第 5 课：挂单 makeOrders

**目标：** 看懂「批量、部分成功、多余 ETH 退回」。用时约 30 分钟。

打开 `EasySwapOrderBook.sol` 的 `makeOrders` 和 `_makeOrderTry`。准入是 `whenNotPaused` + `nonReentrant`。一次传入多张单，返回同样长度的 `OrderKey` 数组。

```solidity
for (uint256 i = 0; i < orderAmount; ++i) {
    uint128 buyPrice;
    if (newOrders[i].side == LibOrder.Side.Bid) {
        buyPrice = Price.unwrap(newOrders[i].price) * newOrders[i].nft.amount;
    }
    OrderKey newOrderKey = _makeOrderTry(newOrders[i], buyPrice);
    newOrderKeys[i] = newOrderKey;
    if (OrderKey.unwrap(newOrderKey) != OrderKey.unwrap(LibOrder.ORDERKEY_SENTINEL)) {
        ETHAmount += buyPrice;
    }
}
if (msg.value > ETHAmount) {
    _msgSender().safeTransferETH(msg.value - ETHAmount);
}
```

卖单的 `buyPrice` 是 0。买单先算出 `单价 × amount`。

`_makeOrderTry` 必须同时成立：`maker == msg.sender`、价格不是 0、`salt` 不是 0、没过期（`expiry == 0` 表示永不过期）、这把 key 的 `filledAmount` 还是 0。然后：

```solidity
if (order.side == LibOrder.Side.List) {
    if (order.nft.amount != 1) return LibOrder.ORDERKEY_SENTINEL;
    IEasySwapVault(_vault).depositNFT(newOrderKey, order.maker, order.nft.collection, order.nft.tokenId);
} else if (order.side == LibOrder.Side.Bid) {
    if (order.nft.amount == 0) return LibOrder.ORDERKEY_SENTINEL;
    IEasySwapVault(_vault).depositETH{value: uint256(ETHAmount)}(newOrderKey, ETHAmount);
}
_addOrder(order);
emit LogMake(...);
```

字段不合格走另一个出口：返回值默认就是 `0x0`，只发 `LogSkipOrder`。

### 部分成功

`amount` 不对、`maker` 不是自己、价格为 0、`salt` 为 0、已过期、这把 key 用过或取消过：这一格返回 `0x0`，循环继续。前面已经挂上的单留在金库和盘口里。

两张 List，第一张 `amount = 1` 且已授权，第二张 `amount = 2`。返回值是 `[第一张的 OrderKey, 0x0]`。第一张的 NFT 已经在 Vault 里。第二张在 `depositNFT` 之前就返回了。

有两种失败会把整笔交易回滚，前面做成的单也一起撤销：

- List 的 NFT 没授权给 Vault。`depositNFT` 里的转账失败，不是返回 `0x0`。
- 附带的 ETH 不够付成功的那些 Bid。`depositETH{value: 单价 × amount}` 转不出去。

### 多余 ETH 退回

只有成功的 Bid 才把 `单价 × amount` 累加进 `ETHAmount`。循环结束后，`msg.value` 多出来的部分退给调用者。

附带 3 ETH，两张 Bid 各要 1.2 ETH，第二张 `amount = 0` 失败。`ETHAmount` 只有 1.2，Vault 里第一张记下 1.2 ETH，退回 1.8 ETH。第二张没有金库余额。

**过关：** 一次 `makeOrders` 里两张 List，第二张 `amount = 2`，第一张会不会已经锁进 Vault？会，这是部分成功。如果第二张是「没授权」呢？整笔回滚，第一张也不会留下。

---

## 第 6 课：取消与改价

**目标：** 看懂退资产，以及 edit 为什么是「取消 + 新建」。用时约 35 分钟。

两个入口都是批量的，合约暂停时不能调。取消把资产退回 maker。改价是先取消旧单、再挂一张新单，OrderKey 一定变。

### 6.1 取消

`cancelOrders` 逐张调用 `_cancelOrderTry`。某一张失败只把对应位置记成 `false`，后面的继续。

只有 maker 本人能取消，而且 `filledAmount < amount`。别人来取消，或者已经买满，不 `revert`，只发 `LogSkipOrder`。

```solidity
if (order.maker == _msgSender() && filledAmount[orderKey] < order.nft.amount) {
    _removeOrder(order);
    if (order.side == LibOrder.Side.List) {
        IEasySwapVault(_vault).withdrawNFT(orderHash, order.maker, order.nft.collection, order.nft.tokenId);
    } else if (order.side == LibOrder.Side.Bid) {
        uint256 availNFTAmount = order.nft.amount - filledAmount[orderKey];
        IEasySwapVault(_vault).withdrawETH(
            orderHash,
            Price.unwrap(order.price) * availNFTAmount,
            order.maker
        );
    }
    _cancelOrder(orderKey); // filledAmount = type(uint256).max
}
```

卖单：NFT 从 Vault 转回卖家。能取消时就是整件退回。

买单：只退还没成交的部分，`单价 × (amount - filledAmount)`。集合出价买 3 件、已成交 1 件，取消时退 2 件的 ETH。

然后这把旧 key 不能再 `match`，也不能拿原内容重新 `makeOrders`。

### 6.2 改价

`editOrders` 的每一项是旧 key 加一张新 `Order`。失败返回 `0x0`，旧单还在盘口上。买单如果新金额更高，差额累加进 `bidETHAmount`；调用者多付的 ETH 最后退一次。

不能改的是 `saleKind`、`side`、`maker`、`collection`、`tokenId`。完全成交的也不能改。`price`、`amount`、`expiry`、`salt` 可以换成新的。`salt` 仍不能为 0，新单的哈希也不能已经被用过。

```solidity
if (
    oldOrder.saleKind != newOrder.saleKind ||
    oldOrder.side != newOrder.side ||
    oldOrder.maker != newOrder.maker ||
    oldOrder.nft.collection != newOrder.nft.collection ||
    oldOrder.nft.tokenId != newOrder.nft.tokenId ||
    filledAmount[oldOrderKey] >= oldOrder.nft.amount
) {
    return (LibOrder.ORDERKEY_SENTINEL, 0);
}
```

通过之后：旧单摘出盘口并标成取消，新单 `_addOrder`。卖单只换记账，NFT 不离开 Vault：

```solidity
if (oldOrder.side == LibOrder.Side.List) {
    IEasySwapVault(_vault).editNFT(oldOrderKey, newOrderKey);
}
```

卖家把 1 ETH 改成 1.5 ETH：NFT 仍在金库，OrderKey 变了，旧 key 不能再 match。改成 0.8 ETH 也一样，卖单没有 ETH 要退。

### 买单可以挂高，也可以挂低

未成交的 ETH 本来就在 Vault 里，改价不会先退回钱包再整笔重存。补进去的差额也进金库，不是 OrderBook。

旧的剩余 = `旧单价 × (旧 amount − 已成交件数)`。新的一笔 = `新单价 × 新 amount`。

```solidity
if (newRemainingPrice > oldRemainingPrice) {
    deltaBidPrice = newRemainingPrice - oldRemainingPrice;
    vault.editETH{value: deltaBidPrice}(oldKey, newKey, oldRemainingPrice, newRemainingPrice, maker);
} else {
    // 新的更低或相等：不补钱，多出来的由 Vault 退给 maker
    vault.editETH(oldKey, newKey, oldRemainingPrice, newRemainingPrice, maker);
}
```

金库里，新的更高时新 key 记 `msg.value + 旧剩余`。新的更低时新 key 只留下 `newETHAmount`，差额打给 maker。

例子：单价 1 ETH、买 3 件，已成交 1 件，金库里还锁着 2 ETH。

- 改成单价 1.5 ETH、数量 2 件，新单要 3 ETH。这次再付 1 ETH，新 key 记下 3 ETH。少付则整笔回滚，旧单还在。
- 改成单价 0.5 ETH、数量 2 件，新单只要 1 ETH。新 key 记下 1 ETH，另外 1 ETH 退回买家。

**过关：** 卖家把 1 ETH 改成 1.5 ETH，NFT 有没有离开 Vault？OrderKey 变了没有？旧 key 还能 match 吗？（没有离开；变了；不能。）买单能不能改成更低的单价？（能。多出来的 ETH 从金库退回。）

---

## 第 7 课：撮合 —— 全项目最核心

**目标：** 能不看代码讲清两个分支的资金流。用时约 50 分钟。

打开 `_isMatchAvailable`，再打开 `_matchOrder`。建议把函数打印出来，左右标「卖家吃 Bid / 买家吃 List」。

### 为什么撮合是核心

挂单、取消、改价都是在整理盘口。NFT 和 ETH 还锁在金库里，所有权没有换人。`makeOrders` 只做托管，`cancelOrders` 把没成交的退回去，`editOrders` 换一把 OrderKey。这三步结束时，BAYC #42 还不是买家的。

`matchOrder` 才同时做完三件事：NFT 交给买家，ETH 按成交价付给卖家，协议费从成交价里扣下留在 OrderBook。`OrderKey`、`filledAmount`、价格树都是为这一步准备的。没有人调用 `matchOrder`，盘口可以一直挂着，但没有发生过一次成交。

### 7.1 `_isMatchAvailable`：只判断能不能配

这个函数不转 NFT、不转 ETH，也不决定成交价。`_matchOrder` 一进来就先调用它，任意一条不过就整笔回滚。六条都不看价格，也不看调用者是谁。

```solidity
require(sellOrderKey != buyOrderKey, "HD: same order");
require(sellOrder.side == List && buyOrder.side == Bid, "HD: side mismatch");
require(sellOrder.saleKind == FixedPriceForItem, "HD: kind mismatch");
require(sellOrder.maker != buyOrder.maker, "HD: same maker");
require(
    buyOrder.saleKind == FixedPriceForCollection ||
        (sellOrder.nft.collection == buyOrder.nft.collection &&
            sellOrder.nft.tokenId == buyOrder.nft.tokenId),
    "HD: asset mismatch"
);
require(
    filledAmount[sellOrderKey] < sellOrder.nft.amount &&
        filledAmount[buyOrderKey] < buyOrder.nft.amount,
    "HD: order closed"
);
```

1. 两把 OrderKey 不能相同。
2. 卖单必须是 List，买单必须是 Bid。
3. 卖单必须是 `FixedPriceForItem`。集合出价只能出现在买侧。
4. 两边 maker 不能是同一个人。
5. 买单是集合出价时，不看买单的 tokenId，交割用卖单上的那一件。否则两边的 collection 和 tokenId 必须相同。卖单是 BAYC `#42`、买单写着「这个合集任意一件」，可以配，买家拿到的是 `#42`。
6. 两边都还没关单。取消后 `filledAmount` 是 `type(uint256).max`，过不了。集合出价买 3 件、只成交 1 件，`1 < 3`，还可以再配。

价格和「谁来吃单」在通过之后：卖家来吃，成交价用买价；买家来吃，成交价用卖价。两边都不是 `msg.sender`，`revert "HD: sender invalid"`。

### 7.2 分支一：`msg.sender == sellOrder.maker`（卖家接受出价）

```
require(msg.value == 0)
成交价 fillPrice = buyOrder.price        ← 用买价
Bid 必须已在盘口（从 Vault 拿得出 ETH）
List 可以已在盘口，也可以是「现挂现吃」（isSellExist == false）
  已在盘口：Vault.withdrawNFT 给买家
  不在盘口：Vault.transferERC721 从卖家钱包直转（卖家要已授权 Vault）
ETH：Vault.withdrawETH(bidKey, fillPrice, OrderBook)
     → 卖家收到 fillPrice - fee
     → fee 留在 OrderBook
```

贯穿例子：B 事先锁了 1.2 ETH，S 接受。成交价 = 1.2 ETH。S 得 1.17 ETH，协议费 0.03 ETH 留在 OrderBook，NFT #42 给 B。

买单 `filledAmount + 1`。集合 Bid 没买满就还留在盘口；买满后第 6 条会拒绝再配。卖单如果本来在盘口，整张摘掉，filled 写成 amount。

### 7.3 分支二：`msg.sender == buyOrder.maker`（买家接受挂单）

```
List 必须已在盘口（orders[sellKey].maker != 0），否则 _validateOrder 对空单失败
成交价 fillPrice = sellOrder.price       ← 用卖价，单价
Bid 可以不在盘口，这就是 Buy Now，不必事先 make Bid：
  不在盘口：msg.value >= fillPrice。外层退款按买单上的 price 算，所以这张临时买单的 price 写成准备付的金额
  在盘口：买单 price >= fillPrice，从 Vault 抽出这一件的单价
NFT：Vault.withdrawNFT 给买家
卖家收到 fillPrice - fee
买单 price > 卖价时，多出的 ETH 退买家
costValue：Bid 已在盘口时是 0；Buy Now 时是买单 price。matchOrder 用 msg.value - costValue 退余款
```

贯穿例子：S 挂 1 ETH，B 直接 match，买单 price 写成 1 ETH，`msg.value >= 1 ETH`。S 得 0.975 ETH，#42 给 B。B 多付的部分退回。

### 7.4 谁不能调

`msg.sender` 既不是卖单 maker 也不是买单 maker → revert `"HD: sender invalid"`。第三方不能替别人成交（没有经纪人角色）。

**过关：** 卖家吃 Bid 时成交价用哪边？买家吃 List 时呢？为什么不对称？

（卖家吃单用买价：卖家愿意按对方出的价卖。买家吃单用卖价：买家愿意按挂牌价买。各取「被吃那一侧」的报价。）

---

## 第 8 课：批量 match 与 multicall

**目标：** 知道 delegatecall 在这里干什么。用时约 25 分钟。

两个函数都在 `EasySwapOrderBook.sol`。合约暂停时都不能调。`delegatecall` 让里面的 `msg.sender` 仍是用户，盘口数据也仍是这份订单簿。

### matchOrders

每一项是一对卖单加买单。循环里 `delegatecall` 到 `matchOrderWithoutPayback`，不直接调用 `_matchOrder`。内部直接调用的话，一对单 `revert`，前面已经成交的也会被撤掉。`delegatecall` 把失败收成 `success = false`。

```solidity
(bool success, bytes memory data) = address(this).delegatecall(
    abi.encodeWithSignature(
        "matchOrderWithoutPayback(...)",
        matchDetail.sellOrder,
        matchDetail.buyOrder,
        msg.value - buyETHAmount
    )
);
if (success) {
    successes[i] = true;
    if (matchDetail.buyOrder.maker == _msgSender()) {
        buyETHAmount += abi.decode(data, (uint128));
    }
} else {
    emit BatchMatchInnerError(i, data);
}
if (msg.value > buyETHAmount) {
    _msgSender().safeTransferETH(msg.value - buyETHAmount);
}
```

`matchOrderWithoutPayback` 只做撮合，自己不退 ETH。它带 `onlyDelegateCall`：`address(this) != self` 才通过。别人直接 `call` 这个函数，两个地址相同，调用会被拒绝。

传进去的 ETH 是 `msg.value - 已经花掉的 buyETHAmount`。只有调用者是买单 maker 时，才把返回的 `costValue` 累加进去。卖家来吃 Bid 时 `costValue` 是 0，不从附带的 ETH 里扣。循环结束后多付的一次退回。

附带 3 ETH，买两件各 1 ETH 的卖单，第二件已经被人买走。第一件成交，`buyETHAmount = 1`；第二件发出 `BatchMatchInnerError`。最后退回 2 ETH。

### multicall

它不新写业务，只把已有入口串进一笔交易。允许的只有 `makeOrders`、`cancelOrders`、`editOrders`、`matchOrder`、`matchOrders`。别的 selector 会 `revert "HD: unsupported selector"`。

```solidity
require(_isSupportedMulticallSelector(selector), "HD: unsupported selector");
if (_isValueSensitiveSelector(selector)) {
    ++valueSensitiveCallAmount;
    require(valueSensitiveCallAmount <= 1, "HD: multi value calls");
}
(bool success, bytes memory result) = address(this).delegatecall(callData);
```

`delegatecall` 时，每个子调用看到的 `msg.value` 都是一开始付的那一整笔。两个子调用都会花钱的话，同一笔 ETH 可能被花两次。所以 `makeOrders`、`editOrders`、`matchOrder`、`matchOrders` 里，一次最多只能有一个。`cancelOrders` 不碰 `msg.value`，可以和其中任意一个放在一起。

`revertOnFail = true` 时，任意一步失败，整笔回滚。`false` 时只发 `MulticallInnerError`，后面的调用继续。退款仍由里面的函数自己做，`multicall` 不再退一次。

**过关：** 「先挂一个 Bid，再立刻吃一张卖单」能不能放进同一次 multicall？不能。两个都要付款，会被 `"HD: multi value calls"` 拦住。要分两笔交易，或者买的时候直接走 Buy Now，只付一次。

---

## 第 9 课：把一张资金流图画完（合上代码）

**目标：** 不看文件，用贯穿例子走通两条路。用时约 20 分钟。

**路 A：挂 1 ETH，别人立即买**

```
S: NFT #42 ──make List──► Vault
B: 1 ETH ──match（当场 Bid）──► OrderBook
OrderBook ──0.975 ETH──► S
Vault ──#42──► B
0.025 ETH 留在 OrderBook（Owner 以后 withdrawETH）
```

**路 B：先出价 1.2 ETH，卖家接受**

```
B: 1.2 ETH ──make Bid──► Vault
S: 可以事先 make List，也可以当场交 NFT
S 调 match，msg.value = 0
Vault ──1.2 ETH──► OrderBook ──1.17 ETH──► S
NFT ──► B
协议费 0.03 ETH 留在 OrderBook
```

两条路的协议费都按 **成交价** 算：`fillPrice × 250 / 10000`。路 A 的成交价是卖价 1 ETH，路 B 的成交价是买价 1.2 ETH。

---

## 第 10 课：管理员

**目标：** 知道 Owner 能做什么。用时约 15 分钟。

OrderBook 上只有 Owner 能调：

- `setVault`：换金库地址，不能是 `address(0)`
- `setProtocolShare`：改费率，不能超过 1000。已经挂着的单下次成交按新费率算
- `pause` / `unpause`
- `withdrawETH(recipient, amount)`：把留在 OrderBook 里的协议费转走

Vault 上 Owner 调 `setOrderBook`。

**过关：** pause 之后 `makeOrders` 还能用吗？不能，`make` / `cancel` / `edit` / `match` 都有 `whenNotPaused`。已经锁在 Vault 里的 NFT 出不来，等 `unpause` 之后才能 cancel 或 match。

---

## 读代码时建议的文件顺序（复习用）

1. `LibOrder.sol` 结构 + hash
2. `EasySwapVault.sol` 存取
3. `OrderStorage.sol` `_addOrder` 与 `getBestPrice`
4. `EasySwapOrderBook.sol` `makeOrders` → `_cancelOrderTry` → `_editOrderTry`
5. `_isMatchAvailable` + `_matchOrder` 两个分支
6. `ProtocolManager` 费率
7. `matchOrders` / `multicall` 的 delegatecall

---

## 总自检（闭卷）

1. NFT 授权给谁？挂单后 NFT 在哪个合约？

   授权给 Vault，不是 OrderBook。挂单成功后 NFT 在 Vault，`NFTBalance[orderKey] = tokenId`。

2. OrderKey 包含哪些字段？改价为什么换 key？

   前缀是手写常量 `ORDER_TYPEHASH`。后面拼 `side`、`saleKind`、`maker`、`nft`（`tokenId`、`collection`、`amount`）、`price`、`expiry`、`salt`。`filledAmount` 不算。`price` 一变，哈希就是一把新 key。旧 key 在 `editOrders` 里被标成取消。卖单用 `editNFT` 把 tokenId 从旧 key 改记到新 key，NFT 不离开 Vault。

   ![image-20261122121119553](https://files.seeusercontent.com/2026/11/22/i2Yw/image-20261122121119553.png)

3. 价格树 List 侧取最高还是最低？

   最低。`getBestPrice(collection, List)` 走 `first()`。卖价 1.5、1.0、2.0 里最优是 1.0。同一档先成交 `head`。

4. `makeOrders` 某一单失败，其它单会不会回滚？

   字段不合格（`amount` 不对、`salt` 为 0、已过期）不会，这一格返回 `0x0`。两种会整笔回滚：NFT 没授权给 Vault；买单带来的 ETH 不够。

5. 买家 Buy Now 要不要先 make Bid？钱从 msg.value 还是 Vault 出？

    Buy Now 不用先 `make Bid`。卖单必须已经在盘口上，钱从这次调用的 `msg.value` 出。只有买家事先挂过 Bid、ETH 已经锁在 Vault 里，成交时才从 Vault 抽。

6. 卖家接受 Bid 时 msg.value 为什么必须是 0？

   卖家接受 Bid 时，ETH 已经在 Vault 里，成交价用的是买价，从 Bid 那一格 `withdrawETH` 抽出来。卖家自己不付钱，所以 `msg.value` 必须是 0，否则 `"HD: value > 0"` 回滚。

7. 集合 Bid 和单品 List 成交时，tokenId 以谁为准？

   以卖单为准。 集合 Bid 不看买单上的 `tokenId`，交割用单品 List 上的那一件

8. 协议费当时打给 Owner 吗？还是积在 OrderBook？

   OrderBook

9. `filledAmount == max` 还能 match 吗？

   不能

10. 第三方能不能调用 `matchOrder` 帮两边成交？

    不能。`msg.sender` 必须是卖单或买单的 maker。没有 relayer 角色——这和 Seaport 允许他人 fill 签名单不一样。

---

## 可选：跑测试

```
cd EasySwapContract
npm install
npx hardhat test
```

先看与 `makeOrders`、`matchOrder` 相关的用例，对照第 5、7 课。测试里的地址、mock NFT 可以忽略，盯「先授权 Vault → make → match → ownerOf 变成买家」。

---

## 以后再用：换一个订单簿项目，按这张表去搜

现在不用看。以后打开另一个 NFT 市场时，用中间一列回忆 EasySwap，用右边的词去新项目里搜。

| 问题 | EasySwap 的答案 | 换项目时看哪里 |
| --- | --- | --- |
| 订单在链上还是链下 | 链上 | 搜 Order 是 `mapping` 还是 `signature` |
| 资产谁托管 | Vault，仅 OrderBook 可动 | 搜 custody / vault / escrow |
| 价格怎么排 | 红黑树 + FIFO | 搜 bestPrice / heap / sorted |
| 成交谁发起 | Taker 调 match | 搜 fill / match / take |
| 协议费 | protocolShare / 10000 | 搜 fee / royalty |
| 后端是否管钱 | 否，扫事件 | 搜 FilterLogs vs 私钥发交易 |
| 索引方式 | 事件 + 断点区块 | 搜 last_indexed_block / subgraph |

先画「用户 → 前端 → 合约 / 索引器」三条线，再补「List vs Bid」两条资金流。
