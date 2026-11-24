# Pledge 合约拆解学习笔记

> 范围：只用学 **智能合约**。前端、后端不在本笔记里。  
> 代码目录：`pledgev2/contracts/`  
> 用法：按课时往下走。每课先写「打开哪个文件」，再写「看什么、想什么」，最后有过关题。过关再进下一课。

---

## 怎么用

1. 在编辑器里打开本笔记和对应 `.sol`，左右对照。
2. 不要背函数名。每读一个函数，只问三件事：**谁能调、什么时候能调、钱有没有动。**
3. 贯穿全笔记的数字例子（建议抄在纸上）：

```
抵押物：1 BTC
BTC 价格：30000 U    稳定币价格：1 U
抵押率 martgageRate：150% = 1.5e8
年化利率 interestRate：10% = 1e7
期限：settle 到 end = 182.5 天（半年）
放款人存入：50000 U（后面会发现只能成交 20000）
最小存款 minAmount：100e18
精度：calDecimal = 1e18，费率/利率 baseDecimal = 1e8
```

4. 学习以 `pledge/` 和 `multiSignature/` 为准。`mock/` 和测试里的函数签名可能不一样（例如测试用的 `MockPledgePool.depositBorrow` 多一个 deadline，`settle` 用 `onlyOwner` 而不是多签）。

回看时先记住这四句：`depositLend` / `depositBorrow` 不铸 sp/jp。`settle` 不转钱，成交额取两边较小的一边。到期拿走的是 `finishAmountLend`，不是 `stakeAmount`。管理员不调用 `finish`，钱就留在池子里。

---

## 学习地图（先看一眼，不必此刻读懂）

```
contracts/
├── pledge/
│   ├── PledgePool.sol          ★ 核心，约 960 行，大部分课时在这里
│   ├── DebtToken.sol           sp / jp 凭证
│   ├── AddressPrivileges.sol   谁能 mint/burn
│   └── BscPledgeOracle.sol     价格
├── multiSignature/
│   ├── multiSignature.sol      多签本身
│   └── multiSignatureClient.sol  validCall 门禁
├── library/SafeTransfer.sol    转入转出（含原生币）
└── interface/                  先当目录，用到再点开
```

池状态机（从第 0 课开始就要能画）：

```
createPoolInfo          settle()                 finish()
   ┌──────┐  双方存款  ┌──────┐              ┌────────┐
   │ MATCH│ ─────────► │EXECUTION│ ─────────► │ FINISH │ ──► withdraw
   └──────┘            └──────┘              └────────┘
       │                   │
       │ 有一方没人存       │ 抵押物贬值 liquidate()
       ▼                   ▼
    UNDONE              LIQUIDATION
```

---

## 第 0 课：先搞懂产品（不看代码）

**目标：** 能用自己的话讲清 Pledge 和 Aave 差在哪。用时约 20 分钟。

Pledge 要解决的事：DeFi 借贷大多是**浮动利率、随时存取**。Pledge 做成**固定利率、固定期限**：建池时把利率和到期日写死，募集期内两边把资产打进合约，到点一次性撮合。

三个角色：

| 角色 | 做什么 |
| --- | --- |
| 放款人 | 存稳定币，到期拿本金 + 事先锁定的利息 |
| 借款人 | 押 BTC 一类资产，撮合成功后提走稳定币 |
| 管理员 | 建池、到点撮合、到期结算或爆仓清算 |

三个必须先接受的设计（后面读代码会反复撞上）：

1. **`depositLend` 当时不铸凭证。** 能不能成交要等 `settle`。成交的部分由之后的 `claimLend` 按 `settleAmountLend` 铸 sp，没成交的由 `refundLend` 退回稳定币。
2. **利息在建池时就定了。** 到期利息 = 期限占比 × 年化 × 成交本金。
3. **合约里没有闹钟。** `settle` / `finish` / `liquidate` 必须有人发交易，合约只检查「时间到了没、状态对不对」。

**过关：** 合上笔记，回答：用户刚 `depositLend` 成功，手里有没有 spToken？成交的 sp 在哪个函数铸？管理员如果一直不调用 `settle`，池子停在哪？

---

## 第 1 课：权限骨架（先别进 PledgePool）

**目标：** 知道「敏感操作」是怎么被拦住的。用时约 40 分钟。

### 1.1 打开 `multiSignature/multiSignatureClient.sol`

盯 `validCall` 和 `checkMultiSignature`。

执行敏感函数前，合约算一个哈希：

```
msgHash = keccak256(msg.sender, address(this))
```

然后问多签合约：这个哈希有没有凑够签名。  
含义：**批准的是「某个人可以操作某个合约」**，不是「批准调用某一个具体函数」。钥匙一旦给了，建池、settle、改费率都能调。把它理解成管理员钥匙，而不是链上提案机。

### 1.2 打开 `multiSignature/multiSignature.sol`

按调用顺序读三个函数：

1. `createApplication(address to)`：申请「我（msg.sender）要操作合约 to」
2. `signApplication(bytes32 msghash)`：owner 签名
3. `getValidSignature`：签名人数 ≥ `threshold` 才返回非 0

构造函数里 `owners` 是管理员列表，`limitedSignNum` 是阈值（例如 3 人里要 2 人签）。

### 1.3 打开 `pledge/AddressPrivileges.sol`

这是第二道门：`onlyMinter`。  
`addMinter` / `delMinter` 自己还挂着 `validCall`。部署后要把 **PledgePool 地址**加进 minter 白名单，否则后面 `claim` 铸币会失败。

权限串起来是：

```
多签 owners 签名
    → validCall 放行（settle、建池、写价格、加 minter）
        → onlyMinter 放行（mint / burn sp、jp）
```

**过关：** `settle` 挂的是哪道门？`DebtToken.mint` 挂的是哪道门？普通放款人调 `depositLend` 要不要多签？

---

## 第 2 课：DebtToken —— 两种凭证其实是同一个合约

**目标：** 搞清 sp / jp 是什么，以及为什么 mint/burn 必须由池子来做。用时约 25 分钟。

打开 `pledge/DebtToken.sol`，只有 50 行。再打开 `interface/IDebtToken.sol` 对照。

一个池子会部署 **两个** DebtToken 实例：

| 实例 | 给谁 | 何时 mint | 数量含义 | 何时 burn |
| --- | --- | --- | --- | --- |
| spCoin | 放款人 | `claimLend` | 对成交本金的份额 | `withdrawLend` |
| jpCoin | 借款人 | `claimBorrow` | 对应「该还多少抵押物」的欠条 | `withdrawBorrow` |

用贯穿例子：成交借款 20000 U、抵押率 150%

```
sp 总量 ≈ 20000
jp 总量 = 20000 × 1.5 = 30000
```

jp 不是「借到了 30000 U」，而是「这笔债对应 30000 份抵押物份额」。到期 burn 多少 jp，就按比例拿回卖剩的抵押物。

读 `mint` / `burn`：都是 `onlyMinter`，而且 `burn(address _from, uint256)` 可以从别人账上销毁。所以用户 redeem 前，要先把 sp/jp **approve 给 PledgePool**（这是主合约能 burn 你的前提）。

**过关：** 为什么不在 `depositLend` 里 mint sp？如果任何人都能 mint sp，会发生什么？

---

## 第 3 课：预言机 —— 撮合和清算的命门

**目标：** 知道价格从哪来，settle 和清算为什么都要调它。用时约 30 分钟。

打开 `pledge/BscPledgeOracle.sol`，只精读 `getUnderlyingPrice`。

决策树：

```
这个资产配了 Chainlink Aggregator？
  是 → latestRoundData()
  否 → 读管理员写入的 priceMap
```

再看两个写价函数：

- `setPrice(address asset, uint256 price)`：按 **资产地址** 写入，和读取路径一致，**这条是能用的**。
- `setPrices(assets[], prices[])`：循环里写的是 `priceMap[i]`（下标 0,1,2…），读取却按资产地址取。**批量设价等于没生效。** 读到这里停一下，在函数旁边标注「bug」。

PledgePool 里真正取价的入口是 `getUnderlyingPriceView(pid)`，返回 `[稳定币价格, 抵押物价格]`。后面 settle、清算都用这两个数做单位换算：

```
抵押物折合稳定币 = 数量 × (prices[1] / prices[0])
```

**过关：** 测试网如果没接 Chainlink，价格谁说了算？价格被写高或写低，会分别伤害放款人还是借款人？

---

## 第 4 课：PledgePool 骨架 —— 结构体 + 修饰器

**目标：** 不读业务函数，先能根据修饰器判断「这个函数属于哪一段生命周期」。用时约 45 分钟。

打开 `pledge/PledgePool.sol`。先不要往下翻到 `depositLend`。按这个顺序看：

### 4.1 文件头常量与状态

- `calDecimal = 1e18`：份额计算
- `baseDecimal = 1e8`：利率、抵押率、费率
- `baseYear = 365 days`
- `enum PoolState { MATCH, EXECUTION, FINISH, LIQUIDATION, UNDONE }`
- `oracle` / `swapRouter` / `feeAddress` / `lendFee` / `borrowFee`

### 4.2 四个结构体（建议对着纸画表）

`PoolBaseInfo`：建池时定下的规则 + 实时募集量

| 字段 | 记住什么 |
| --- | --- |
| settleTime / endTime | 撮合时刻 / 到期时刻 |
| interestRate | 固定年化，1e8 = 100% |
| maxSupply | 放款上限 |
| lendSupply / borrowSupply | 已存入量（含尚未成交） |
| martgageRate | 抵押率 |
| lendToken / borrowToken | 稳定币 / 抵押物；`address(0)` = 原生币 |
| spCoin / jpCoin | 该池专属凭证 |
| autoLiquidateThreshold | 清算缓冲 |
| state | 当前状态 |

`PoolDataInfo`：settle / finish / liquidate 写下的成交快照（真正分钱看这里，不看 supply）

`LendInfo` / `BorrowInfo`：每个用户在每个池的仓位。注意两个布尔名是反直觉的：

- `hasNoRefund == true` 表示 **已经退过**，不能再退
- `hasNoClaim == true` 表示 **已经领过**凭证

### 4.3 文件底部六个修饰器（倒着读也行）

打开文件最底部：

| 修饰器 | 含义 | 典型函数 |
| --- | --- | --- |
| `notPause` | 没被紧急暂停 | 几乎所有用户函数 |
| `timeBefore` | now < settleTime | deposit* |
| `timeAfter` | now > settleTime | refund* / claim* |
| `stateMatch` | 状态是 MATCH | deposit* |
| `stateNotMatchUndone` | EXECUTION 或 FINISH 或 LIQUIDATION | refund* / claim* |
| `stateFinishLiquidation` | FINISH 或 LIQUIDATION | withdraw* |
| `stateUndone` | UNDONE | emergency* |

读任何用户函数，**先看修饰器再看函数体**。修饰器就是准入条件。

### 4.4 `createPoolInfo`

挂 `validCall`。只做两件事：往 `poolBaseInfo`、`poolDataInfo` 各 push 一条，初始状态 MATCH。不转钱。

**过关：** 把下面函数抄下来，只根据修饰器填「何时可调」：`depositLend`、`claimLend`、`withdrawLend`、`emergencyLendWithdrawal`。先不要翻函数体，填完再去核对。

---

## 第 5 课：募集期 —— 钱怎么进池子

**目标：** 看懂存款只记账、不铸币。用时约 30 分钟。

打开顺序：

1. `library/SafeTransfer.sol` 的 `getPayableAmount` / `_redeem`
2. `PledgePool.sol` 的 `depositLend`、`depositBorrow`

`getPayableAmount`：token 是 `address(0)` 就收 `msg.value`，否则 `transferFrom`。所以前端（你这阶段不用管）要先 ERC20 `approve`。

`depositLend` 做的事只有：

- 检查没超过 `maxSupply - lendSupply`
- 转入资产
- `userLendInfo[msg.sender][pid].stakeAmount += 金额`
- `pool.lendSupply += 金额`
- 把 `hasNoClaim` / `hasNoRefund` 置回 false（允许同一用户多次存）

**没有 mint。** `depositBorrow` 同理，只是没有 maxSupply 限制。

用贯穿例子，假设 settle 前：

```
放款人 A 存 50000 U   → lendSupply = 50000
借款人 B 押 1 BTC     → borrowSupply = 1
state 仍是 MATCH
```

**过关：** 存款成功后，`poolDataInfo` 里的 `settleAmountLend` 是多少？为什么？

---

## 第 6 课：settle —— 全项目最核心的一课

**目标：** 能手算成交额，并说清「settle 不转钱」。用时约 50 分钟。

打开 `settle` 和它上面的 `checkoutSettle`。挂 `validCall`，要求 `now > settleTime` 且状态 MATCH。

### 6.1 极端情况

`lendSupply == 0` 或 `borrowSupply == 0` → 状态变 `UNDONE`，两边之后走紧急提取。募集失败。

### 6.2 正常撮合（对着纸算）

```
prices = [稳定币价, 抵押物价] = [1, 30000]   （示例，忽略精度）

抵押物总价值 totalValue
  = borrowSupply × (prices[1] / prices[0])
  = 1 × 30000 = 30000 U

可支持借款 actualValue
  = totalValue × 1e8 / martgageRate
  = 30000 / 1.5 = 20000 U
  即「押 1 BTC 最多借 2 万 U」
```

然后两边取小：

| 情形 | 条件 | 成交 |
| --- | --- | --- |
| 放款人过多 | lendSupply 50000 > 20000 | settleAmountLend = 20000，抵押物全部成交 |
| 放款人不足 | lendSupply 例如 15000 < 20000 | settleAmountLend = 15000，抵押物只锁成交所需部分 |

贯穿例子走第一行：

```
settleAmountLend   = 20000 U
settleAmountBorrow = 1 BTC
state              = EXECUTION
```

多出来的 30000 U 还在合约里，但已经不计入成交，等用户自己 `refundLend`。

### 6.3 settle 之后合约里有什么

- 50000 U 还在池子（20000 待借给借款人，30000 待退）
- 1 BTC 还在池子
- **谁也还没拿到凭证，借款人也还没拿到稳定币**

**过关：** 用第二行「放款人只存 15000」手算 `settleAmountBorrow`。公式在 `settle` 的 else 分支。算完再和代码对一下。

---

## 第 7 课：refund 与 claim —— 成交之后怎么分

**目标：** 看懂 share 模式，以及借款人是在 claim 时才拿到稳定币。用时约 45 分钟。

四个函数成对读：`refundLend` ↔ `refundBorrow`，`claimLend` ↔ `claimBorrow`。  
修饰器都是 `timeAfter` + `stateNotMatchUndone`（settle 已经跑过）。

### 7.1 份额公式（本文件反复出现，背下来）

```
userShare = 我的 stakeAmount × 1e18 / 池子总 supply
金额     = 池子某个字段 × userShare / 1e18
```

贯穿例子里只有一个放款人，所以 userShare = 100%。

### 7.2 refundLend

未成交 = `lendSupply - settleAmountLend` = 30000 U，按份额退。  
`hasNoRefund` 防重复领。钱从池子回到用户。

### 7.3 claimLend

```
spAmount = settleAmountLend × userShare = 20000
pool.spCoin.mint(msg.sender, 20000)
```

只铸 sp，**不把稳定币还给放款人**。放款人的 20000 U 已经借出去了，凭证代表「到期来分本息」。

### 7.4 claimBorrow（这课最重要的函数）

```
jp 总量 = settleAmountLend × martgageRate / 1e8 = 20000 × 1.5 = 30000
jpAmount = jp 总量 × userShare
同时：borrowAmount = settleAmountLend × userShare = 20000
_redeem(用户, lendToken, 20000)   ← 稳定币这时才离开池子
```

从这一刻起：

- 借款人口袋里是 20000 U + 30000 jp
- 放款人口袋里是 20000 sp（以及退回的 30000 U）
- 池子里剩下 1 BTC（托管中）

**过关：** claim 的顺序有没有强制？先 claim 再 refund 可不可以？`hasNoClaim` 为 true 之后还能不能再 claim？

---

## 第 8 课：finish —— 到期卖抵押物还本付息

**目标：** 手算利息，并画出「BTC → DEX → 稳定币」的资金流。用时约 40 分钟。

打开 `finish`。`validCall`，要求 `now > endTime` 且状态 EXECUTION。

### 8.1 固定利率公式（和 liquidate 共用，必须会算）

```
timeRatio = (endTime - settleTime) × 1e8 / 365天
interest  = timeRatio × interestRate × settleAmountLend / 1e16
          = (期限/一年) × (年化) × 本金
lendAmount = 本金 + 利息
```

贯穿例子：半年、10%、本金 20000

```
interest  = 0.5 × 0.1 × 20000 = 1000 U
lendAmount = 21000 U
```

再加 lendFee 得到要向 DEX「买到」的稳定币数量，然后 `_sellExactAmount`：用抵押物换稳定币。

### 8.2 钱怎么走（按代码顺序）

1. 算出需要卖多少 BTC，才能换回至少 21000 U（加手续费）
2. 去 PancakeSwap 换
3. `require(amountIn >= lendAmount)` —— **换不够就整笔 revert**，这是到期路径的滑点保护
4. 多换出来的稳定币进 `feeAddress`
5. `finishAmountLend` = 给放款人准备的稳定币（本+息）
6. 没卖掉的 BTC 扣 `borrowFee` 后记入 `finishAmountBorrow`
7. 状态 → FINISH

池子此刻：

```
稳定币：约 21000 U（等放款人 burn sp 来领）
剩余 BTC：1 - 卖掉的数量（等借款人 burn jp 来领）
```

**过关：** 如果 DEX 流动性极差，换不回 21000 U，finish 会成功吗？用户的 sp 能 redeem 吗？

---

## 第 9 课：liquidate —— 和 finish 对照着读

**目标：** 找出三个差异，尤其是「没有滑点保护」。用时约 35 分钟。

先读 `checkoutLiquidate`（view），再读 `liquidate`。

清算条件：

```
抵押物现值 = settleAmountBorrow × (抵押物价 / 稳定币价)
清算线     = settleAmountLend × (1 + autoLiquidateThreshold)
现值 < 清算线 → 可清算
```

贯穿例子：成交借 20000 U，阈值 10%（`1e7`）

```
清算线 = 20000 × 1.1 = 22000 U
当 1 BTC 现价 < 22000 U，可触发
```

注意两点：

- 用的是 settle 锁定的数量，借款人**不能**事后单独补仓
- 清算线按**本金**算，不含已计利息，和 finish 的口径不同

然后把 `liquidate` 和 `finish` 并排：

| | finish | liquidate |
| --- | --- | --- |
| 触发 | 时间到 endTime | 价格跌破阈值（仍要多签来调） |
| 利息公式 | 相同 | 相同 |
| 滑点保护 | 有 `require(amountIn >= lendAmount)` | **没有**，换少了记在 `liquidationAmounLend` |
| 分钱字段 | finishAmount* | liquidationAmoun* |
| 粒度 | 整池 | 整池（不是按单个借款人） |

`liquidate` 里注释也写了：`amountIn - lendAmount` 可能小于 0。亏损由放款人承担。这是本合约最值得讲的实现取舍。

**过关：** 为什么说 Pledge 的清算比 Aave「粗」？一个池里两个借款人，其中一人的抵押物其实还够，会被一起清算吗？

---

## 第 10 课：withdraw 与紧急提取 —— 钱怎么回到用户

**目标：** 分清「按凭证份额赎回」和「募集失败原路退」。用时约 30 分钟。

### 10.1 `withdrawLend` / `withdrawBorrow`

修饰器：`stateFinishLiquidation`。先 burn 凭证，再按凭证占总量的比例，从 `finishAmount*` 或 `liquidationAmoun*` 里分。

```
spShare = 我烧的 sp / settleAmountLend
赎回稳定币 = finishAmountLend × spShare     （FINISH）
         或 liquidationAmounLend × spShare （LIQUIDATION）
```

关键：**赎回看的是池子结算后的余额，不是你当初存的 stakeAmount。** 清算亏了，放款人按份额承担。

FINISH 还要求 `now > endTime`。LIQUIDATION 只要求过了 settleTime。

### 10.2 `emergencyLendWithdrawal` / `emergencyBorrowWithdrawal`

只有 `UNDONE`。直接把 `stakeAmount` 原路退回，不经过份额计算。因为根本没成交。

**过关：** 状态还在 EXECUTION 时，放款人能不能 withdrawLend？他手里已经有 sp 了，钱却取不出来——卡在哪一步、谁能解开？

（答案：卡在管理员还没调 `finish` 或 `liquidate`。这是第 0 课「没有闹钟」的后果。）

---

## 第 11 课：DEX 内部函数与费用

**目标：** 知道到期时合约怎么跟 swap 说话，以及平台从哪抽成。用时约 25 分钟。

打开 `_getSwapPath` / `_getAmountIn` / `_sellExactAmount` / `_swap` / `redeemFees`。

路径：lendToken 或 borrowToken 为 `address(0)` 时换成 WETH，所以 BNB 也能进池。  
`_swap` 的 `minOut` 传 0，滑点保护只在 `finish` 的 require 里；`liquidate` 没有。  
`_safeApprove(..., uint256(-1))` 是对 router 的无限授权。`setSwapRouterAddress` 也是 `validCall`，管理员换 router 等于换资金出口。

费用两处：

- 换汇盈余（`amountIn - lendAmount`）→ `feeAddress`（稳定币）
- 剩余抵押物 × `borrowFee` → `feeAddress`（抵押物）

**过关：** 用户存款会被收 lendFee 吗？还是只有到期/清算才收费？

---

## 第 12 课：把一张资金流图画完（合上代码）

**目标：** 不看文件，能讲完整条链路。用时约 30 分钟。

用贯穿例子把数字填进箭头：

```
募集
  放款人 50000 U ──depositLend──► 池子
  借款人 1 BTC   ──depositBorrow─► 池子

settle（只记账）
  settleAmountLend   = 20000
  settleAmountBorrow = 1 BTC

claim / refund
  30000 U ──refundLend──► 放款人
  20000 sp ──claimLend mint──► 放款人
  20000 U + 30000 jp ──claimBorrow──► 借款人

finish（假设半年利息 1000，忽略手续费与滑点）
  部分 BTC ──DEX──► 21000 U 留在池里
  剩余 BTC 留在池里

withdraw
  放款人 burn 20000 sp ──► 21000 U
  借款人 burn 30000 jp ──► 剩余 BTC
```

对照检查：池子进出是否守恒？放款人是否先于借款人拿到足额本息（在 finish 成功的前提下）？

---

## 第 13 课：部署关系（知道即可）

打开 `pledgev2/README.md` 的部署顺序，对照构造函数：

```
1. multiSignature(owners, threshold)
2. DebtToken × 2（sp、jp，都传入多签地址）
3. BscPledgeOracle(multiSignature)
4. PledgePool(oracle, swapRouter, feeAddress, multiSignature)
5. sp.addMinter(PledgePool) 和 jp.addMinter(PledgePool)   ← 漏了 claim 会失败
```

`PledgePool` 构造函数还要求 oracle / router / feeAddress 都非零。

**过关：** 为什么 DebtToken 和 PledgePool 都要传入同一个 multiSignature 地址？

---

## 读代码时建议的文件顺序（复习用）

如果以后只想快速过一遍，按这个点：

1. `PledgePool.sol` 顶部状态机注释 + 底部修饰器
2. `createPoolInfo` → `deposit*` → `settle`
3. `claim*` / `refund*`
4. `finish` 与 `liquidate` 对照
5. `withdraw*` / `emergency*`
6. `DebtToken.sol` + `AddressPrivileges.sol`
7. `multiSignatureClient.sol` 的 `validCall`
8. `BscPledgeOracle.sol` 的 `getUnderlyingPrice` 和 `setPrices` bug

---

## 总自检（全部课时结束后，闭卷）

能不看代码答出下面 10 题，合约部分就算过关：

1. `depositLend` / `depositBorrow` 会铸 sp/jp 吗？成交部分在哪个函数铸？
2. settle 转不转钱？成交额两个字段怎么来？
3. 固定利率的「固定」锁在哪几个参数上？
4. 借款人是在哪一个函数里拿到稳定币的？
5. jp 的数量为什么等于 `settleAmountLend × martgageRate`？
6. finish 失败（滑点）时，状态会不会变成 FINISH？
7. liquidate 比 finish 少了哪一句保护？亏的是谁？
8. 清算是按人还是按池？
9. `validCall` 批准的是函数，还是「调用者 × 合约」？
10. 管理员不调用 finish，用户拿着 sp 能不能把本息取走？卡在什么状态？

卡在第 10 题是正常的：这正是本项目「无链上定时器」要记住的产品风险。

---

## 可选：跑测试加深印象

只作为验证，不是必须。在 `pledgev2/`：

```
npm install
npx hardhat test
```

先看 `tests/PledgePool.test.js` 里 `Create Pool info` 和 `deposit lend ... pool state is match` 两个用例，对照第 4、5 课。注意测试部署的是 `MockPledgePool`，和正式 `PledgePool` 的权限、函数参数不完全相同。正式逻辑以 `contracts/pledge/PledgePool.sol` 为准。
