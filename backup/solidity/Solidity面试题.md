# Solidity 高频面试题与参考答案

这是一份精简版题库，保留 50 道高频题，适合面试前集中复习。回答时建议按照“结论、原因、风险、验证方式”的顺序组织。

## 一、Solidity 与 EVM 基础

### 1. 什么是智能合约？

智能合约是部署在区块链上的确定性程序。Solidity 源码经过编译生成 ABI 和字节码，部署交易把字节码写入链上，之后由 EVM 执行。

### 2. EVM 为什么要求节点执行相同交易？

EVM 是确定性虚拟机。不同节点执行同一交易后必须得到相同状态，网络才能在没有中心服务器的情况下达成共识。

### 3. 合约部署后能否直接修改？

普通合约代码部署后不能直接修改。常见升级方式是部署新合约迁移数据，或使用代理合约通过 `delegatecall` 更换逻辑合约。代理需要重点防范升级权限和 storage layout 风险。

### 4. 什么是 Gas？哪些操作比较贵？

Gas 是 EVM 计算资源的计量单位。写入 storage、创建合约、复杂循环和外部调用通常比较贵。Gas limit 是本次交易允许消耗的上限，Gas price 是单位 Gas 价格。

### 5. `constructor` 和 `public` 状态变量有什么特点？

`constructor` 只在部署时执行一次，适合初始化 owner 等数据。`public` 状态变量会自动生成 getter，但不会自动生成 setter。

### 6. `storage`、`memory` 和 `calldata` 有什么区别？

`storage` 是跨交易永久保存的数据；`memory` 只在当前调用期间存在且可修改；`calldata` 是外部函数的只读参数，不能修改，通常比复制到 memory 更省 Gas。

### 7. `constant` 和 `immutable` 有什么区别？

`constant` 必须在编译期确定，`immutable` 可以在构造函数中确定，部署后都不能修改。它们通常比普通可修改状态变量更节省存储成本。

### 8. Solidity 0.8 如何处理整数溢出？

Solidity 0.8 及以上默认检查溢出和下溢，发生时交易回滚。`unchecked` 可以省 Gas，但只有在已经证明不会溢出时才能使用。

### 9. 值类型和引用类型有什么区别？

值类型赋值时复制数据，例如 `uint`、`bool`、`address`。数组、struct、bytes 和 string 属于引用类型，赋值时要注意数据位置以及复制还是引用。

### 10. `mapping` 为什么不能遍历？如何解决？

mapping 只保存从 key 到 value 的映射，没有连续的 key 列表、length 或遍历接口。若需要枚举，必须额外维护数组，并防止同一个 key 重复加入。

## 二、函数、类型与资金处理

### 11. `public`、`external`、`internal`、`private` 的区别是什么？

`public` 可被内部和外部调用；`external` 主要作为外部入口；`internal` 供当前合约和子合约使用；`private` 只限当前合约内部。`private` 不代表链上隐私。

### 12. `view`、`pure` 和 `payable` 分别是什么？

`view` 可以读状态但不能修改；`pure` 既不能读状态也不能修改；`payable` 允许函数接收 ETH。函数缺少 `payable` 时，调用不能附带 ETH。

### 13. `msg.sender` 和 `tx.origin` 有什么区别？为什么不用 `tx.origin` 做权限？

`msg.sender` 是当前调用者，合约调用合约时会变化；`tx.origin` 是整笔交易最初的 EOA。恶意合约可以诱导 owner 调用它，再利用 `tx.origin` 通过权限检查，因此权限应使用 `msg.sender`。

### 14. `msg.value` 和 `address(this).balance` 表示什么？

`msg.value` 是当前调用附带的 wei；`address(this).balance` 是合约当前持有的 ETH 总余额。

### 15. `transfer`、`send` 和 `call` 转账有什么区别？

`transfer` 固定提供 2300 Gas，失败自动回滚；`send` 失败返回 `false`，必须检查；`call` 更灵活，返回成功标志和数据，也必须检查，并且要防范重入。

### 16. 什么是 CEI 模式？

CEI 是 Checks-Effects-Interactions：先检查权限和参数，再更新自己的状态，最后进行外部调用。提款函数应先清零可提款金额，再向用户转账。

### 17. 如何安全实现提款？

检查权限和可提款余额，先把用户额度清零，再使用 `call` 转账并检查返回值。可配合 OpenZeppelin 的 `ReentrancyGuard`，并测试恶意接收合约回调。

### 18. `require`、`assert`、`revert` 和自定义 error 怎么选？

`require` 用于用户输入和业务前提；`assert` 用于理论上不应失败的内部不变量；`revert` 用于复杂分支主动回滚；自定义 error 比长字符串更省 Gas，并能携带参数。

### 19. `modifier` 有什么作用？

modifier 用来复用权限、状态和前置检查逻辑，`_` 表示插入被修饰函数体。常见例子是 `onlyOwner` 和 `whenNotPaused`。

### 20. 为什么链上循环必须有上限？

交易受 Gas limit 限制。无界数组循环可能因为 Gas 不够而永久失败，导致功能不可用。应使用分页、单用户操作、mapping 或拉取模式。

## 三、ERC20、NFT 与事件

### 21. ERC20 的核心接口有哪些？

主要包括 `totalSupply`、`balanceOf`、`transfer`、`approve`、`allowance` 和 `transferFrom`，核心事件是 `Transfer` 和 `Approval`。

### 22. `transfer`、`approve` 和 `transferFrom` 如何配合？

`transfer` 从调用者自己的余额转账；`approve` 由 owner 设置 spender 的额度；`transferFrom` 由 spender 代扣 owner 的余额，并消耗授权额度。

### 23. ERC20 的 `approve` 有什么风险？

标准 `approve` 通常是覆盖额度。直接从一个非零额度改成另一个额度可能被抢跑，常见缓解方式是先设为 0，再设置新额度，或使用增加/减少授权函数。

### 24. `mint` 和 `burn` 要注意什么？

`mint` 同时增加接收者余额和 totalSupply；`burn` 同时减少持有者余额和 totalSupply。两者必须限制权限，并检查零地址、余额和数量边界。

### 25. ERC20 和 ERC721 的区别是什么？

ERC20 是可互换的同质化资产，用余额表示；ERC721 的每个 `tokenId` 都是独立资产，通过 `ownerOf(tokenId)` 查询所有者。

### 26. ERC721 的安全转账有什么区别？

安全转账在接收方是合约时，会检查它是否实现 `IERC721Receiver`，避免 NFT 被转入无法管理的合约。

### 27. NFT 市场购买流程应如何设计？

检查 listing 有效、价格正确、卖家仍是 owner 且授权有效；先关闭 listing，再转移 NFT 和结算资金；外部调用必须遵守 CEI 并使用重入防护。

### 28. 事件有什么作用？`indexed` 是什么？

事件写入交易日志，供前端和索引服务查询，合约本身不能读取历史事件。`indexed` 参数进入 topics，便于按地址或 tokenId 过滤。事件不能代替合约需要读取的 storage。

## 四、继承、调用与代理

### 29. `virtual` 和 `override` 是什么？

`virtual` 表示父合约函数允许重写，`override` 表示子合约正在重写父函数。重写函数的参数和返回类型要兼容，不能降低可见性。

### 30. interface 和 abstract contract 有什么区别？

interface 主要包含函数、事件和错误的声明，适合描述外部合约；abstract contract 可以有状态变量和部分实现，但因存在未实现函数不能直接部署。

### 31. `call`、`delegatecall` 和 `staticcall` 的区别是什么？

`call` 在目标合约上下文中执行；`delegatecall` 使用目标代码但保留当前合约的 storage、地址和调用上下文；`staticcall` 禁止状态修改，适合只读调用。

### 32. 为什么 `delegatecall` 有很大风险？

目标代码会在当前合约的 storage 和权限上下文中执行。不可信代码可能修改 owner、余额或实现地址，storage layout 不一致还会造成槽位错写。

### 33. 代理合约为什么能升级？

代理保存数据和稳定地址，使用 `delegatecall` 执行逻辑合约代码。升级时只更换逻辑合约地址，因此用户仍然调用同一个代理地址。

### 34. 代理升级要注意哪些问题？

必须保护升级权限，防止重复初始化，保持 storage layout 兼容，通常只能在末尾追加变量，并在升级前后进行 layout 检查和完整测试。

## 五、安全与 Gas

### 35. 什么是重入攻击？

合约进行外部调用时，攻击合约在原函数状态更新前回调进入该函数，重复执行提款等逻辑。主要防护是 CEI、重入锁和正确的余额记账。

### 36. 什么是访问控制漏洞？

敏感函数如 mint、withdraw、pause、upgrade 或修改管理员没有正确限制调用者，导致普通用户可以执行。应使用 `Ownable`、`AccessControl` 或多签，并逐个测试权限路径。

### 37. 什么是拒绝服务风险？

例如合约循环给所有用户转账，只要一个接收者 revert 或数组过大，整笔交易就失败。应改成记录每个人可领取金额，由用户自行提现。

### 38. 什么是抢跑？如何缓解？

攻击者观察内存池中的待处理交易并提高费用插队。竞价、投票等场景可采用 commit-reveal，敏感交易也可使用私有交易通道。

### 39. 链上随机数为什么不安全？

区块时间、区块哈希和公开状态通常可预测或存在有限操纵空间，不能直接用于公平抽奖。应使用可验证随机数，例如 VRF，或设计严谨的承诺-揭示流程。

### 40. 常见 Gas 优化有哪些？

减少 storage 写入，合理进行 storage packing，外部只读动态参数使用 calldata，循环中缓存 length 和 storage 值，用事件保存链下历史，并通过工具测量而不是凭感觉优化。

### 41. 什么是 storage packing？

把连续的小类型字段放进同一个 32 字节 slot。只有字段排列确实减少 slot 且不会带来大量类型转换时才有收益，最终应以编译器布局和 Gas 报告为准。

### 42. OpenZeppelin 为什么常用？

它提供经过广泛使用、测试和审计的 Ownable、ERC20、ERC721、Pausable 和 ReentrancyGuard 等组件，能减少重复实现错误，但仍需正确配置版本和权限。

## 六、测试、前端与综合设计

### 43. Hardhat 和 Foundry 有什么区别？

Hardhat 生态偏 JavaScript/TypeScript、脚本和插件；Foundry 使用 Rust 工具链，测试通常用 Solidity 编写，速度快并支持 fuzz、invariant 和 cheatcode。两者都可用于编译、测试和部署。

### 44. 合约测试应覆盖哪些内容？

覆盖正常流程、权限、revert、自定义 error、事件、零值、最大值、重复调用、时间边界、余额不足、外部调用失败和重入攻击。

### 45. `view` 调用和发送交易有什么区别？

`view` 通常通过 `eth_call` 查询，不上链、不花 Gas；写操作需要 signer 签名，支付 Gas，等待交易被打包，并处理 pending、成功和失败状态。

### 46. 前端如何通过 ABI 调用合约？

ABI 描述函数、参数、返回值和事件。前端用 provider 读取，用 signer 发送交易，等待 receipt 后解析 logs，再刷新页面状态。

### 47. 为什么前端不能只依赖交易返回值显示历史？

交易返回值只代表一次调用，刷新页面后会丢失内存状态。历史应通过 view、事件日志、区块查询或索引服务恢复。

### 48. 什么是预言机？使用价格预言机要注意什么？

预言机把链下数据提交到链上，因为合约不能直接访问互联网。使用时要检查数据新鲜度、更新时间、decimals、异常值、精度和价格跳变。

### 49. 跨链消息和跨链资产转移有什么区别？

跨链消息传递任意数据；跨链资产转移还需要锁定/销毁源链资产，再在目标链释放/铸造。必须考虑验证者信任、最终性、重放攻击和异步失败。

### 50. 如何设计一个安全的众筹或支付合约？

先确定状态机、权限和资金流，再设计 storage、事件和失败恢复。捐款检查阶段和截止时间，成功后授权提款，失败后用户自行退款；所有资金操作遵守 CEI，加重入和暂停机制，并用测试覆盖边界。

## 高频复习顺序

1. 先掌握第 1 至 20 题：Solidity 基础、函数和资金处理。
2. 再掌握第 21 至 34 题：ERC 标准、事件、合约调用和升级。
3. 重点背熟第 35 至 42 题：重入、权限、DoS、抢跑和 Gas。
4. 最后练习第 43 至 50 题：测试、前端、预言机、跨链和项目设计。

面试回答不要只背定义，要主动补充：谁能调用、修改了哪些状态、外部调用在哪里、失败是否回滚、如何写测试验证。