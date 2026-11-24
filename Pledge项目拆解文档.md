# Pledge 项目拆解文档

> 学习目标：从**业务**理解项目，而不是背代码。读完应能独立画出架构图、讲清核心流程、说明各模块职责，并快速对照同类型 Web3 借贷项目。
>
> 代码依据：当前仓库 `pledgev2`（合约）、`pledge-backend`（后端）、`pledge-fe`（前端）。  
> 仓库：https://github.com/MetaNodeAcademy/ProjectBreakdown-Pledge

---

## 0. 一句话讲清这个项目

Pledge 做的是 **固定利率、固定期限的抵押借贷**。

管理员先建一个「期限 + 利率 + 抵押率」都锁死的资金池；放款人在结算日前存入稳定币，借款人存入抵押物（如 BTC）。到期后合约按事先锁定的利率卖出抵押物、还本付息；中途抵押物跌破阈值则整池清算。

它要填的市场空白是：Aave / Compound 一类协议是**浮动利率、随时存取**，主要服务短线交易；Pledge 把「期限」和「利率」在建池时写死，服务的是需要确定性融资成本的中长期需求。

---

## 1. 项目解决什么问题

| 对比维度 | Aave / Compound | Pledge |
| --- | --- | --- |
| 利率 | 由资金利用率实时浮动 | 建池时锁定，到期才结算 |
| 期限 | 无固定到期日，随时借还 | 有 `settleTime`（撮合）和 `endTime`（到期） |
| 撮合方式 | 资金池共享，随存随借 | 募集期双边存入，到期一次性撮合 |
| 清算粒度 | 按单个借款人仓位 | **整池一次性清算** |
| 凭证 | aToken / cToken 随存随铸 | settle 之后才 mint `spToken` / `jpToken` |

**业务上的三个关键选择：**

1. **先存后撮、不先发凭证。** 存款阶段只记账，能不能成交要等 `settle()`。先铸再退会把凭证逻辑绕死。
2. **利息在建池时就确定。** `interest = 期限占比 × 年化利率 × 成交本金`。这就是「固定利率」的合约含义。
3. **链上没有 cron。** `settle` / `finish` / `liquidate` 必须由多签持有者主动发交易。合约只校验时间与状态，不会自己跳状态。

---

## 2. 仓库怎么拆

```
ProjectBreakdown-Pledge/
├── Pledge/                 # 需求清单、业务流程说明（学习材料，不是运行代码）
├── pledgev2/               # 智能合约（Hardhat + Solidity 0.6.12）
│   └── contracts/
│       ├── pledge/         # 核心：PledgePool / DebtToken / Oracle / 权限
│       ├── multiSignature/ # 多签治理
│       ├── library/        # SafeTransfer 等
│       └── mock/           # 测试用代币与 Uniswap
├── pledge-backend/         # Go 后端，拆成两个进程
│   ├── api/                # HTTP API（给前端读列表）
│   └── schedule/           # 定时任务（把链上状态镜像进 MySQL）
└── pledge-fe/              # React DApp（钱包写链 + 调后端读列表）
```

读代码时记住：**资金与状态以合约为准；后端是只读索引；前端是操作入口。**

---

## 3. 整体架构图

```mermaid
flowchart TB
    subgraph 用户层
        Lender[放款人 Lender]
        Borrower[借款人 Borrower]
        Admin[多签管理员]
        Wallet[MetaMask]
    end

    subgraph G2["前端 DApp"]
        Market[市场列表页]
        PoolPage[存入页 Coin_pool]
        Portfolio[仓位页 Market_Mode]
        Dex[DEX 兑换]
    end

    subgraph 后端
        API[api 进程 :8080]
        Task[schedule 进程]
        MySQL[(MySQL)]
        Redis[(Redis)]
    end

    subgraph 链上合约
        Pool[PledgePool]
        SP[spToken / DebtToken]
        JP[jpToken / DebtToken]
        Oracle[BscPledgeOracle]
        MultiSig[multiSignature]
        Swap[PancakeSwap Router]
    end

    Lender --> Wallet
    Borrower --> Wallet
    Admin --> Wallet
    Wallet --> PoolPage
    Wallet --> Portfolio
    Market -->|GET 池列表| API
    API --> MySQL
    Task -->|每 2 分钟 view 调用| Pool
    Task -->|每 1 分钟读价格| Oracle
    Task --> MySQL
    Task --> Redis
    PoolPage -->|approve + deposit*| Pool
    Portfolio -->|claim / refund / withdraw| Pool
    Admin -->|validCall: 建池/结算/到期/清算| MultiSig
    MultiSig -->|授权校验| Pool
    Pool -->|mint / burn| SP
    Pool -->|mint / burn| JP
    Pool -->|取价| Oracle
    Pool -->|卖抵押物换稳定币| Swap
```

**分工一句话：**

- **合约**：管钱、管状态机、管清算。
- **后端 schedule**：定时把合约 `view` 数据抄进数据库。
- **后端 api**：给前端提供池列表、代币信息、价格。
- **前端**：展示列表走 HTTP；存取款、领凭证走钱包直连合约。



---

## 4. 角色、资产与凭证

### 4.1 角色（放款人、借款人、多签管理员、平台）

存款只把资产打进池子，不铸 sp/jp。`settle` 只记账、不转钱。成交的凭证在 `claim*` 里铸，没成交的资产在 `refund*` 里退。

**放款人 Lender**

存稳定币，到期拿本金加固定利息：`depositLend` → `claimLend` → `withdrawLend`。

| 动作 | 注解 |
| --- | --- |
| `depositLend` | 募集期、状态是 MATCH 才能存。稳定币进 PledgePool，只记 `stakeAmount`。这一笔不铸 spToken |
| `refundLend` | settle 之后，没成交的稳定币按份额退回。成交的那部分不在这里拿 |
| `claimLend` | 按成交额 `settleAmountLend` 的份额铸 spToken。只拿凭证，不拿稳定币 |
| `withdrawLend` | 烧掉 sp。到期走 `finishAmountLend` 里自己的份额（本金 + 利息）。若已清算，走 `liquidationAmounLend`。不用原来的 `stakeAmount` |
| `emergencyLendWithdrawal` | 一边存成 0、池子变成 UNDONE 时，按原存入额退回。没有 sp 可烧 |

**借款人 Borrower**

押加密资产，settle 后提走稳定币：`depositBorrow` → `claimBorrow` → `withdrawBorrow`。

| 动作 | 注解 |
| --- | --- |
| `depositBorrow` | 和放款人同一窗口存抵押物。只记 `stakeAmount`，不铸 jpToken |
| `refundBorrow` | 没用来覆盖借款的抵押物退回 |
| `claimBorrow` | 铸 jpToken，同时把成交的稳定币打给借款人。jp 总量 = `settleAmountLend × martgageRate / 1e8`，表示这笔债对应的抵押物份额，不是借到的 U |
| `withdrawBorrow` | 烧掉 jp，按份额拿回剩余抵押物。到期用 `finishAmountBorrow`，清算用 `liquidationAmounBorrow` |
| `emergencyBorrowWithdrawal` | 池子 UNDONE 时，抵押物原路退回 |

**多签管理员**

建池、撮合、到期结算、清算。这些函数都过 `validCall`：批准的是「这个调用者 × 这份合约」，不绑定某一个函数。链上没有定时器，不调用状态就不会自己跳。

| 动作 | 注解 |
| --- | --- |
| `createPoolInfo` | 写入 `settleTime`、`endTime`、`interestRate`。利率在这里锁定，池子存续期间不改 |
| `settle` | `settleTime` 之后、状态仍是 MATCH。按预言机价格和抵押率算出成交额，取两边较小的一边。不转币、不铸 sp/jp。一边供给是 0 则池子变成 UNDONE |
| `finish` | 到期、状态是 EXECUTION。卖掉抵押物覆盖本金加利息，写下 `finishAmountLend` / `finishAmountBorrow`。有滑点保护：换回来的稳定币必须不少于应还额 |
| `liquidate` | 还在 EXECUTION、已过 `settleTime` 就可以调。利息公式和 `finish` 相同，但没有滑点保护，换不够的部分由放款人承担 |
| `setPause` | 暂停用户存取。已经在池子里的资产要等恢复后才能 claim / withdraw |

**平台**

不单独调函数。`finish` / `liquidate` 把 `lendFee`、`borrowFee` 打到 `feeAddress`。

### 4.2 一个池子里的两类资产

- **lendToken**：放款资产，通常是稳定币（BUSD / DAI）。`address(0)` 表示原生币（BNB）。
- **borrowToken**：抵押资产（BTC / BNB）。名字容易误导：这里的「borrow」指借款人拿来抵押的币，不是借出去的币。

### 4.3 两种凭证（都是 `DebtToken`）

| 凭证 | 给谁 | 何时铸造 | 含义 | 何时销毁 |
| --- | --- | --- | --- | --- |
| **spToken**（Share / 放款凭证） | 放款人 | `claimLend` | 对池内「成交本金」的份额 | `withdrawLend` 时 burn，按份额分本息 |
| **jpToken**（Debt / 借款欠条） | 借款人 | `claimBorrow` | 到期要对应赎回的抵押物份额 | `withdrawBorrow` 时 burn，按份额拿剩余抵押物 |

数量关系（settle 之后）：

```
sp 总量 ≈ settleAmountLend
jp 总量 = settleAmountLend × martgageRate / 1e8
```

例如：成交借款 20000 U，抵押率 150% → jp 总量 = 30000。jp 不是「借到了多少 U」，而是「这笔债对应多少抵押物份额」。

---

## 5. 核心业务流程：池状态机

整个项目围着一张状态机转。**先背这张图，再看函数。**

```mermaid
stateDiagram-v2
    [*] --> MATCH: createPoolInfo 建池
    MATCH --> EXECUTION: settle 双边都有资金
    MATCH --> UNDONE: settle 任一方为 0
    EXECUTION --> FINISH: finish 到期卖抵押物还本息
    EXECUTION --> LIQUIDATION: liquidate 抵押物跌破阈值
    FINISH --> [*]
    LIQUIDATION --> [*]
    UNDONE --> [*]

    note right of MATCH
        settleTime 之前
        双方只能存款
    end note
    note right of EXECUTION
        可 refund 超额、claim 凭证
        借款人领 jp 并提走稳定币
    end note
    note right of FINISH
        endTime 之后
        burn 凭证分钱
    end note
    note right of LIQUIDATION
        整池出局
        放款人优先分清算所得
    end note
    note right of UNDONE
        募集失败
        原路紧急提取
    end note
```

时间轴：

```
建池 ──募集期── settleTime ──执行期── endTime ──完成
         MATCH              EXECUTION          FINISH
                              │
                              └── 抵押物贬值 → LIQUIDATION
```

每个用户函数都挂了三层门禁：`notPause` + 时间修饰器 + 状态修饰器。读任意函数，先看修饰器，就能知道它属于哪一段生命周期。

---

## 6. 完整用户旅程

![](https://files.seeusercontent.com/2026/11/16/3ugX/Pledge.png)

### 6.1 放款人

```mermaid
sequenceDiagram
    participant U as 放款人
    participant ERC as lendToken
    participant P as PledgePool
    participant SP as spToken

    U->>ERC: approve(PledgePool, amount)
    U->>P: depositLend(pid, amount)
    Note over P: MATCH 且 now < settleTime<br/>只记账，不铸币
    Note over P: 管理员调用 settle()
    U->>P: refundLend(pid)
    Note over P: 按份额退回未成交部分
    U->>P: claimLend(pid)
    P->>SP: mint(user, 成交份额)
    Note over P: 管理员调用 finish() 或 liquidate()
    U->>P: withdrawLend(pid, spAmount)
    P->>SP: burn(user, spAmount)
    P->>U: 转回稳定币本金+利息（或清算残值）
```

### 6.2 借款人

和放款人对称，但 `claimBorrow` 会**当场把稳定币打给借款人**，同时 mint jpToken：

1. `approve(borrowToken)` → `depositBorrow`
2. settle 后 `refundBorrow`（退未成交抵押物）
3. `claimBorrow`：mint jp + `_redeem(lendToken)` 提走借款
4. 到期 `withdrawBorrow`：burn jp，拿回卖剩的抵押物

### 6.3 管理员（链下人工触发）

合约没有定时器。到点后必须有人发交易：

1. 多签：`createApplication(PledgePool地址)` → 达到阈值的 owner `signApplication`
2. 被授权的调用者执行 `settle` / `finish` / `liquidate`
3. `validCall` 只校验「这个调用者是否被批准操作这个合约」，**不绑定具体函数**

---

## 7. 合约层

### 7.1 合约如何协作

```mermaid
flowchart LR
    MS[multiSignature]
    MSC[multiSignatureClient.validCall]
    AP[AddressPrivileges.onlyMinter]
    PP[PledgePool]
    DT[DebtToken sp/jp]
    OR[BscPledgeOracle]
    SW[UniswapV2Router]

    MS -->|getValidSignature| MSC
    MSC --> PP
    MSC --> OR
    MSC --> AP
    AP --> DT
    PP -->|mint/burn| DT
    PP -->|getPrices| OR
    PP -->|swapExactTokensForTokens| SW
    OR -->|可选| CL[Chainlink Aggregator]
```

| 合约 | 职责 |
| --- | --- |
| `PledgePool.sol` | 唯一核心。建池、存取、撮合、到期、清算、收费 |
| `DebtToken.sol` | 极简 ERC20。只允许 minter mint/burn |
| `AddressPrivileges.sol` | minter 白名单。部署后把 PledgePool 加进去 |
| `BscPledgeOracle.sol` | 价格源：优先 Chainlink，否则管理员手写 `priceMap` |
| `multiSignature.sol` | 多签：创建申请、签名、达阈值 |
| `multiSignatureClient.sol` | `validCall`：执行前询问多签合约 |
| `SafeTransfer.sol` | 兼容 ERC20 与原生币的转入转出 |

部署顺序（见 `pledgev2/README.md`）：

1. `multiSignature`（先定管理员和阈值）
2. `DebtToken`（sp / jp，构造函数传入多签地址）
3. Oracle、SwapRouter
4. `PledgePool(oracle, swapRouter, feeAddress, multiSignature)`
5. `addMinter(PledgePool)`，否则 claim 时铸币会失败

### 7.2 用户操作 → 合约方法

| 业务动作 | 方法 | 谁能调 | 准入条件 | 链上资金变化 |
| --- | --- | --- | --- | --- |
| 建池 | `createPoolInfo` | 多签 | `validCall` | 无，只 push 两个 struct |
| 放款存入 | `depositLend` | 用户 | MATCH 且 `now < settleTime` | lendToken 转入池 |
| 借款存押 | `depositBorrow` | 用户 | 同上 | borrowToken 转入池 |
| 撮合锁定成交额 | `settle` | 多签 | MATCH 且 `now > settleTime` | **不转钱、不铸币，只记账** |
| 退未成交放款 | `refundLend` | 用户 | EXECUTION/FINISH/LIQUIDATION | 退 lendToken |
| 退未成交抵押 | `refundBorrow` | 用户 | 同上 | 退 borrowToken |
| 领放款凭证 | `claimLend` | 用户 | 同上 | mint spToken |
| 领借款 + 提稳定币 | `claimBorrow` | 用户 | 同上 | mint jpToken + 转出 lendToken |
| 到期结算 | `finish` | 多签 | EXECUTION 且 `now > endTime` | 卖抵押物 → 稳定币，扣费 |
| 爆仓清算 | `liquidate` | 多签 | EXECUTION 且 `now > settleTime` | 同上，无滑点保护 |
| 放款赎回 | `withdrawLend` | 用户 | FINISH 或 LIQUIDATION | burn sp，按份额转 lendToken |
| 借款赎回抵押 | `withdrawBorrow` | 用户 | 同上 | burn jp，按份额转剩余 borrowToken |
| 募集失败提款 | `emergencyLend/BorrowWithdrawal` | 用户 | UNDONE | 原路退回 stakeAmount |
| 暂停 | `setPause` | 多签 | `validCall` | 无 |

存款阶段**不 mint**。凭证发生在 claim。这是读代码时最容易和 Aave 搞混的一点。

### 7.3 settle：一次性撮合（最核心）

`settle()` 只做三件事：取价、按抵押率算能借多少、两边取小。

设抵押率 `martgageRate = 150% = 1.5e8`，借款人押了 1 BTC，预言机 BTC = 30000、稳定币 = 1：

```
抵押物价值 totalValue = 1 × (30000 / 1) = 30000 U
可支持借款 actualValue = 30000 × 1e8 / 1.5e8 = 20000 U
```

然后：

- 放款人存了 50000 U（供过于求）→ 成交 20000 U，多出的 30000 之后 `refundLend`
- 放款人只存了 15000 U（供不应求）→ 成交 15000 U，多余抵押物 `refundBorrow`

任一方 supply 为 0 → 状态变 `UNDONE`，走紧急提取。

**settle 不转钱。** 真正给借款人打稳定币的是之后的 `claimBorrow`。

### 7.4 固定利率怎么算

`finish()` 与 `liquidate()` 用同一套公式（`baseDecimal = 1e8`，`baseYear = 365 days`）：

```
timeRatio = (endTime - settleTime) × 1e8 / 365天
interest  = timeRatio × interestRate × settleAmountLend / 1e16
          = (期限/一年) × (interestRate/1e8) × 本金
lendAmount = 本金 + 利息
```

若年化 10%，`interestRate = 1e7`，期限半年，成交 20000 U → 利息 1000 U，到期应付 21000 U。

建池时利率和期限已锁死，所以到期利息在建池那一刻就可预知。这就是产品上的「固定利率」。

### 7.5 资金如何流转

```mermaid
flowchart TD
    subgraph G1["募集 MATCH"]
        L1[放款人稳定币] -->|depositLend| Pool[PledgePool 托管]
        B1[借款人抵押物] -->|depositBorrow| Pool
    end

    subgraph G2["执行 EXECUTION"]
        Pool -->|refundLend 未成交| L2[放款人]
        Pool -->|refundBorrow 未成交| B2[借款人]
        Pool -->|claimBorrow 成交稳定币| B3[借款人拿走 U]
        Pool -->|claimLend mint| SP[spToken]
        Pool -->|claimBorrow mint| JP[jpToken]
    end

    subgraph G3["到期 FINISH 或清算 LIQUIDATION"]
        Pool -->|DEX 卖出部分抵押物| DEX[PancakeSwap]
        DEX -->|换回稳定币| Pool
        Pool -->|超出本息部分| Fee[feeAddress]
        Pool -->|剩余抵押物扣 borrowFee| Fee
        L3[放款人 burn sp] -->|withdrawLend| U1[本金+利息或清算残值]
        B4[借款人 burn jp] -->|withdrawBorrow| U2[剩余抵押物]
    end
```

资金安全设计：

- 放款人优先：卖抵押物先覆盖本息，剩余才还给借款人。
- 清算时没有 `require(amountIn >= lendAmount)`。滑点亏损记在 `liquidationAmounLend` 上，放款人可能收不回全额。`finish()` 有这道保护，清算没有——这是协议对「爆仓」的取舍。
- 清算是**整池**，不是 Aave 那种逐仓。一个池子爆仓，该池所有借款人一起出局。

### 7.6 清算条件

`checkoutLiquidate`：

```
抵押物现值 = settleAmountBorrow × (抵押物价 / 稳定币价)
清算线     = settleAmountLend × (1 + autoLiquidateThreshold)
现值 < 清算线 → 可清算
```

注意：判定用的是结算时锁定的 `settleAmount*`，不是用户事后补仓。settle 之后抵押物整池托管，没有「单独加保证金」这条路径。

另外：清算线用的是本金，**不含已计利息**。这和 `finish` 里「本金 + 利息」不是同一口径。

### 7.7 多签实际在保什么

`checkMultiSignature` 计算：

```
msgHash = keccak256(msg.sender, address(this))
```

含义：批准的是「某调用者 × 某目标合约」，不是「某一笔具体函数调用」。一旦签过，该调用者可以对这个合约的所有 `validCall` 方法生效。学习时把它理解成「管理员钥匙」，而不是链上 DAO 提案机。

敏感操作全部 `validCall`：建池、settle、finish、liquidate、改费率、改路由、暂停、给 DebtToken 加 minter、给 Oracle 写价格。

### 7.8 预言机

`BscPledgeOracle.getUnderlyingPrice`：

1. 若该资产配了 Chainlink Aggregator → `latestRoundData()`
2. 否则读管理员写入的 `priceMap`

生产安全性完全取决于价格源。settle 成交比例和清算触发都靠它。测试网常见路径是 `setPrice` 手写价格。

`setPrices` 循环里写的是 `priceMap[i]`（数组下标）而不是 `priceMap[assets[i]]`，和读取路径对不上。批量设价这条路实际无效，单笔 `setPrice(address, price)` 才是可用接口。

---

## 8. 链下：只做列表

合约是资产和凭证的账本。用户存、取、领凭证都直接调 `PledgePool`，不经过后端。后端也不替池子调用 `settle` / `finish` / `liquidate`。

Schedule 每隔几分钟用 `eth_call` 读 `PoolLength`、`PoolBaseInfo`、`PoolDataInfo` 和预言机价格，有变化才写入数据库。这是轮询 view，不是扫事件日志。API 只读这些表，给页面提供池子列表、利率、期限和 logo。利率、期限来自链上字段；logo 是后端拼上去的。列表会比链上晚大约一个同步周期。个人的 `stakeAmount` 没有入库，页面要看自己的仓位就直接读合约。

---

## 9. 从业务角度怎么读代码

合约按学习笔记的课时读。拆解文档看到这里，函数级细节以笔记为准。

1. `createPoolInfo`：期限和利率写在哪。
2. `depositLend` / `depositBorrow`：只记账，不铸 sp/jp。
3. `settle`：成交额怎么取两边较小的一边，这一步不转钱。
4. `refund*` / `claim*`：没成交的怎么退，成交的凭证在哪个函数铸。
5. `finish` / `liquidate`：利息公式一样，清算没有滑点保护。
6. `withdrawLend` / `withdrawBorrow`：到期拿的是 `finishAmount*`，不是原来的 `stakeAmount`。

---

## 10. 同类型 Web3 项目怎么快速上手

Pledge 是「有期限的抵押借贷 + 链下索引」这一类。换到别的项目，用同一套问题清单：

| 问题 | 在 Pledge 里的答案 | 换项目时看哪里 |
| --- | --- | --- |
| 资金托管在哪 | PledgePool | 找持有 ERC20 的核心合约 |
| 利率谁决定 | 建池参数，固定 | 搜 `interestRate` / utilization |
| 凭证何时铸 | claim，不是 deposit | 搜 `mint(` 的调用点 |
| 清算粒度 | 整池 | 搜 `liquidate` 参数是 pid 还是 user |
| 价格从哪来 | 自建 Oracle，可接 Chainlink | 搜 `oracle` / `Aggregator` |
| 管理员权限 | 多签 validCall | 搜 `onlyOwner` / `AccessControl` / 多签 |
| 后端是否管钱 | 否，只做索引 | 看有没有 `sendTransaction` / 私钥签名用户操作 |
| 索引方式 | 轮询 view | 搜 `FilterLogs` / subgraph / 轮询 |

先画「用户 → 前端 → 合约 / 后端」三条线，再补状态机，同类型项目可以在一天内讲清楚骨架。

---

## 11. 讲解提纲（答辩 / 面试按这个说）

**1. 这是什么**  
固定利率、固定期限的抵押借贷协议。募集期内双边存入，settle 一次性撮合，到期按锁定利率还本付息。

**2. 和 Aave 的差别**  
Aave 是共享资金池、浮动利率、随存随借、逐仓清算。Pledge 是订单式资金池：期限和利率事先锁死，凭证延后到撮合成功才发，清算按池而不是按人。

**3. 钱怎么走**  
用户把资产打进 PledgePool；settle 只记账；claim 时借款人提走稳定币、双方拿到 sp/jp；finish 时合约经 DEX 卖掉部分抵押物覆盖本息，剩余抵押物给借款人；平台从价差和费率抽成。

**4. 合约怎么分工**  
PledgePool 管状态机；DebtToken 只是可 mint/burn 的 ERC20；Oracle 给 settle 和清算提供价格；多签卡住所有管理操作；DEX 只在到期和清算时被调用。

**5. 链下干什么**  
Schedule 轮询链上 view，把池子抄进数据库给列表用。用户交易不经过后端，个人仓位直接读合约。

**6. 风险点**  
- settle/finish 依赖管理员在线，链上无 cron  
- 清算无滑点保护，薄流动性交易对可能让放款人亏本  
- Oracle 可手写价格，多签作恶能操纵撮合与清算  
- 多签绑定的是「调用者 + 合约」，不是单笔函数  
- 整池清算，一个借款人出问题会连累同池所有人

---

## 12. 关键结构体

合约 `PoolBaseInfo` ↔ 表 `poolbases`：

| 字段 | 业务含义 |
| --- | --- |
| settleTime / endTime | 撮合时刻 / 到期时刻 |
| interestRate | 固定年化，1e8 = 100% |
| maxSupply | 放款募集上限 |
| lendSupply / borrowSupply | 当前已存入量（含未成交） |
| martgageRate | 抵押率，1.5e8 = 150% |
| lendToken / borrowToken | 稳定币 / 抵押物地址 |
| state | 0 MATCH … 4 UNDONE |
| spCoin / jpCoin | 该池专属凭证合约 |
| autoLiquidateThreshold | 清算缓冲，1e7 = 10% |

合约 `PoolDataInfo` ↔ 表 `pooldata`：

| 字段 | 何时写入 | 给谁分 |
| --- | --- | --- |
| settleAmountLend / Borrow | `settle` | 成交本金、成交抵押 |
| finishAmountLend / Borrow | `finish` | 到期后可赎回的稳定币 / 剩余抵押 |
| liquidationAmounLend / Borrow | `liquidate` | 清算后可赎回部分 |

用户仓位只在链上：

```
userLendInfo[user][pid]   = { stakeAmount, refundAmount, hasNoRefund, hasNoClaim }
userBorrowInfo[user][pid] = 同上
```

后端没有用户仓位表。这再次证明：**账户资产以合约为唯一账本。**

---

## 13. 拆解自检

能不看代码回答下面这些问题，就算达到拆解阶段的要求：

1. `depositLend` 这一笔里铸不铸 spToken？成交的那部分在哪个函数铸，没成交的怎么退？
2. `settle` 转不转钱？成交额两个数字怎么来的？
3. 固定利率的「固定」体现在哪几个参数上？
4. 放款人到期拿钱，走的是哪笔余额（`finishAmountLend` 还是原来的 `stakeAmount`）？
5. 后端 schedule 会不会帮池子调用 `settle`？
6. 前端市场列表的 logo 和利率，分别来自哪里？
7. 如果管理员一直不调用 `finish`，用户的钱在哪、能不能自己取？
8. 换一个「固定期限债券」项目，你准备先打开哪个合约函数？

答：

1. `depositLend` 这一笔不铸 sp。成交的部分在 `claimLend` 按 `settleAmountLend` 铸，没成交的由 `refundLend` 退回稳定币。
2. `settle` 不转钱。
3. 固定利率锁在 `_settleTime`、`_endTime`、`_interestRate`。
4. 到期拿的是 `finishAmountLend`，不是 `stakeAmount`。
5. 换一个固定期限项目，先打开 `createPoolInfo`。
6. 不答
7. 执行期内用户只能 refund / claim，**不能** `withdrawLend`。withdraw 必须等到 FINISH 或 LIQUIDATION。管理员不点 `finish`，资金会一直锁在池里——这是「无 cron + 多签推进状态机」的最大产品风险，也是讲架构时最有说服力的一句。
