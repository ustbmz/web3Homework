# 常用 ERC / EIP 标准

## 一、什么是 EIP 和 ERC？

EIP（Ethereum Improvement Proposal，以太坊改进提案）是提出、讨论和记录以太坊改进方案的文档。ERC（Ethereum Request for Comments）主要描述以太坊应用层的标准，例如代币、NFT 和合约接口。

简单来说，ERC 通常属于应用层标准；EIP 的范围更广，也可能涉及以太坊协议本身。ERC 和 EIP 是文档编号与标准类别，不代表某个项目一定实现了该标准。标准也可能处于 Draft（草案）等不同状态，阅读时应查看官方页面的最新状态。

官方标准目录可查阅 [EIP/ERC 列表](https://eips.ethereum.org/all)、[ERCs 仓库](https://github.com/ethereum/ERCs) 和 [EIPs 仓库](https://github.com/ethereum/EIPs)。

下面是常见的应用场景。场景仅用于帮助理解，不代表表中提到的项目一定实现了所有列出的标准。

## 二、常用标准概览

| 标准 | 名称 / 主题 | 主要作用 |
| --- | --- | --- |
| [ERC-20](https://eips.ethereum.org/EIPS/eip-20) | Token Standard | 同质化代币的转账、授权和余额查询接口 |
| [ERC-2612](https://eips.ethereum.org/EIPS/eip-2612) | Permit Extension for ERC-20 | 通过签名设置 ERC-20 授权额度，减少单独发送 `approve` 交易的需要 |
| [ERC-721](https://eips.ethereum.org/EIPS/eip-721) | Non-Fungible Token Standard | 唯一 NFT 的所有权和转移接口 |
| [ERC-1155](https://eips.ethereum.org/EIPS/eip-1155) | Multi Token Standard | 在一个合约中管理多种同质化或非同质化 token，并支持批量操作 |
| [ERC-165](https://eips.ethereum.org/EIPS/eip-165) | Standard Interface Detection | 让合约声明并查询其支持的接口 |
| [ERC-2981](https://eips.ethereum.org/EIPS/eip-2981) | NFT Royalty Standard | 查询 NFT 交易的版税收款方和金额 |
| [ERC-1271](https://eips.ethereum.org/EIPS/eip-1271) | Standard Signature Validation Method for Contracts | 让智能合约钱包验证签名是否有效 |
| [ERC-4626](https://eips.ethereum.org/EIPS/eip-4626) | Tokenized Vaults | 为单一 ERC-20 资产的代币化金库定义存入、取出和份额接口 |
| [ERC-4337](https://eips.ethereum.org/EIPS/eip-4337) | Account Abstraction Using Alt Mempool | 通过 `UserOperation`、EntryPoint、Bundler 等实现账户抽象，而不需要共识层改动 |
| [EIP-712](https://eips.ethereum.org/EIPS/eip-712) | Typed Structured Data Hashing and Signing | 对结构化数据进行哈希和签名，方便钱包展示签名内容 |
| [EIP-7702](https://eips.ethereum.org/EIPS/eip-7702) | Set Code for EOAs | 允许 EOA 通过授权交易委托执行代码，支持账户体验改进 |
| [ERC-8004](https://eips.ethereum.org/EIPS/eip-8004) | Trustless Agents | 提出 Agent 身份、声誉和验证注册表，帮助发现并评估 Agent |
| [ERC-8183](https://eips.ethereum.org/EIPS/eip-8183) | Agentic Commerce | 提出带托管资金和评估者确认的 Agent 任务交易流程 |
| [x402](https://docs.x402.org/) | Internet-native payment protocol | 基于 HTTP `402 Payment Required` 的开放式网络支付协议，不是 ERC 或 EIP |

## 三、关键调用方法速查

下面列出常见的外部接口或交互入口，方便读合约和看调用流程。接口签名会因 Solidity 版本、标准版本和具体实现略有不同；`view` 方法只读链上数据，其他方法通常需要发送交易并支付 gas。事件也很重要，但不是函数调用。

### ERC-20：同质化代币

```solidity
function totalSupply() external view returns (uint256);
function balanceOf(address account) external view returns (uint256);
function transfer(address to, uint256 value) external returns (bool);
function approve(address spender, uint256 value) external returns (bool);
function allowance(address owner, address spender) external view returns (uint256);
function transferFrom(address from, address to, uint256 value) external returns (bool);
```

常见流程：用户调用 `approve(spender, amount)` 授权；被授权的合约再调用 `transferFrom` 转走 token。`Transfer` 和 `Approval` 是常见事件。`name`、`symbol`、`decimals` 属于可选元数据方法。

### ERC-2612：签名授权

```solidity
function permit(
    address owner,
    address spender,
    uint256 value,
    uint256 deadline,
    uint8 v,
    bytes32 r,
    bytes32 s
) external;

function nonces(address owner) external view returns (uint256);
function DOMAIN_SEPARATOR() external view returns (bytes32);
```

用户离线签名后，由任意账户提交 `permit`；验证成功后更新授权额度。`nonces` 用于防止签名重复使用，`deadline` 限制签名有效时间。常见调用顺序是 `permit` 后再由应用调用 `transferFrom`。

### ERC-721：NFT

```solidity
function balanceOf(address owner) external view returns (uint256);
function ownerOf(uint256 tokenId) external view returns (address);
function approve(address to, uint256 tokenId) external;
function getApproved(uint256 tokenId) external view returns (address);
function setApprovalForAll(address operator, bool approved) external;
function isApprovedForAll(address owner, address operator) external view returns (bool);
function transferFrom(address from, address to, uint256 tokenId) external;
function safeTransferFrom(address from, address to, uint256 tokenId) external;
```

`safeTransferFrom` 还有带 `bytes data` 的重载版本。若接收方是合约，它需要实现 `onERC721Received` 接收回调，否则安全转账会回滚。常见事件为 `Transfer`、`Approval` 和 `ApprovalForAll`。

### ERC-1155：多代币

```solidity
function balanceOf(address account, uint256 id) external view returns (uint256);
function balanceOfBatch(address[] calldata accounts, uint256[] calldata ids)
    external view returns (uint256[] memory);
function setApprovalForAll(address operator, bool approved) external;
function isApprovedForAll(address account, address operator) external view returns (bool);
function safeTransferFrom(
    address from, address to, uint256 id, uint256 amount, bytes calldata data
) external;
function safeBatchTransferFrom(
    address from, address to, uint256[] calldata ids, uint256[] calldata amounts, bytes calldata data
) external;
```

转入合约时会检查 `onERC1155Received` 或 `onERC1155BatchReceived` 回调。常见事件为 `TransferSingle`、`TransferBatch`、`ApprovalForAll` 和 `URI`。

### ERC-165：接口检测

```solidity
function supportsInterface(bytes4 interfaceId) external view returns (bool);
```

调用方传入接口标识符，查询目标合约是否声明支持该接口。它是声明式检测，不是安全审计或行为正确性的证明。

### ERC-2981：版税查询

```solidity
function royaltyInfo(uint256 tokenId, uint256 salePrice)
    external view returns (address receiver, uint256 royaltyAmount);
```

市场可用销售价格查询版税接收方和金额，再由市场按自身规则付款。ERC-2981 本身不执行转账。

### ERC-1271：智能合约钱包签名验证

```solidity
function isValidSignature(bytes32 hash, bytes calldata signature)
    external view returns (bytes4 magicValue);
```

有效签名时，合约钱包返回规定的 magic value `0x1626ba7e`。调用方通常以 `staticcall` 查询验证结果。

### ERC-4626：代币化金库

```solidity
function asset() external view returns (address);
function totalAssets() external view returns (uint256);
function convertToShares(uint256 assets) external view returns (uint256);
function convertToAssets(uint256 shares) external view returns (uint256);
function deposit(uint256 assets, address receiver) external returns (uint256 shares);
function mint(uint256 shares, address receiver) external returns (uint256 assets);
function withdraw(uint256 assets, address receiver, address owner) external returns (uint256 shares);
function redeem(uint256 shares, address receiver, address owner) external returns (uint256 assets);
```

常见流程是先授权金库使用底层 `asset`，再调用 `deposit`；取回资产时可调用 `withdraw` 或 `redeem`。标准还定义 `maxDeposit`、`maxMint`、`maxWithdraw`、`maxRedeem` 和对应的 `previewDeposit`、`previewMint`、`previewWithdraw`、`previewRedeem` 查询方法。

### ERC-4337：账户抽象

ERC-4337 既有链上合约接口，也有 Bundler 提供的 RPC；这里列的是常见入口，具体函数签名取决于 EntryPoint 版本。

```solidity
// 以较新的 EntryPoint 接口为例；旧版本使用的结构体名称可能不同
function handleOps(PackedUserOperation[] calldata ops, address payable beneficiary) external;
function getUserOpHash(PackedUserOperation calldata userOp) external view returns (bytes32);

// 智能账户实现的验证与执行接口（具体签名随版本变化）
function validateUserOp(
    PackedUserOperation calldata userOp,
    bytes32 userOpHash,
    uint256 missingAccountFunds
) external returns (uint256 validationData);
function execute(address target, uint256 value, bytes calldata data) external;
```

常见流程：客户端构造并签署 `UserOperation`，通过 Bundler 的 `eth_sendUserOperation` RPC 提交；Bundler 最终调用 EntryPoint 的 `handleOps`。`execute` 等账户执行方法由具体智能账户定义，不是所有账户都使用相同签名。Paymaster 也有自己的验证和结算接口。

### EIP-712：结构化数据签名

EIP-712 本身不是一组必须由合约实现的 Solidity 函数。常见钱包交互入口是 JSON-RPC：

```text
eth_signTypedData_v4(account, typedData)
```

合约端通常实现 domain separator 和结构化数据哈希，再通过 `ecrecover` 或 ERC-1271 验证签名。应用仍需自行设计 nonce、链 ID、合约地址、截止时间等防重放约束。

### EIP-7702：EOA 委托代码

EIP-7702 授权元组的核心字段为：

```text
[chain_id, delegate_address, nonce, y_parity, r, s]
```

钱包创建并签署授权，支持该交易类型的客户端将授权随 set-code transaction 提交。委托代码的具体执行方法由被委托的实现合约定义。

### ERC-8004：Agent 注册、声誉和验证（草案）

```solidity
// Identity Registry
function register(string agentURI) external returns (uint256 agentId);
function setAgentURI(uint256 agentId, string calldata newURI) external;
function getAgentWallet(uint256 agentId) external view returns (address);
function setAgentWallet(
    uint256 agentId, address newWallet, uint256 deadline, bytes calldata signature
) external;

// Reputation Registry
function giveFeedback(
    uint256 agentId, int128 value, uint8 valueDecimals,
    string calldata tag1, string calldata tag2, string calldata endpoint,
    string calldata feedbackURI, bytes32 feedbackHash
) external;
function revokeFeedback(uint256 agentId, uint64 feedbackIndex) external;
function readFeedback(uint256 agentId, address clientAddress, uint64 feedbackIndex)
    external view returns (int128 value, uint8 valueDecimals, string memory tag1,
        string memory tag2, bool isRevoked);

// Validation Registry
function validationRequest(
    address validatorAddress, uint256 agentId, string calldata requestURI, bytes32 requestHash
) external;
function validationResponse(
    bytes32 requestHash, uint8 response, string calldata responseURI,
    bytes32 responseHash, string calldata tag
) external;
```

Identity Registry 以 ERC-721 风格注册 Agent；Reputation Registry 记录反馈；Validation Registry 记录验证请求和验证者响应。以上是草案中的代表性入口，最终应以官方最新版本及具体部署为准。

### ERC-8183：Agent 商务任务托管（草案）

```solidity
function createJob(
    address provider, address evaluator, uint256 expiredAt,
    string calldata description, address hook
) external returns (uint256 jobId);
function setProvider(uint256 jobId, address provider) external;
function setBudget(uint256 jobId, uint256 amount) external;
function fund(uint256 jobId, uint256 expectedBudget) external;
function submit(uint256 jobId, bytes32 deliverable) external;
function complete(uint256 jobId, bytes32 reason) external;
function reject(uint256 jobId, bytes32 reason) external;
function claimRefund(uint256 jobId) external;
```

基本流程：创建任务、设定预算、客户注资托管、服务方提交结果、评估者确认完成或拒绝；到期后可按规则退款。草案中的可选 hook 参数等实现细节可能不同，且通常需要先授权任务合约转移 ERC-20 资金。

### x402：HTTP 支付流程（非 ERC/EIP）

x402 没有统一的 Solidity 合约方法。典型交互是：

```text
客户端请求资源
← 服务端返回 HTTP 402 和支付要求
客户端携带 PAYMENT-SIGNATURE 重新请求
→ 服务端或 Facilitator 验证并结算
← 支付成功后返回资源（可包含 PAYMENT-RESPONSE）
```

具体支付网络、代币、签名载荷和结算方式由 x402 scheme 决定；它可以使用链上交易，但本身是 HTTP 层的支付协议。

## 四、标准说明

### 1. ERC-20：同质化代币

ERC-20 为同质化代币定义常见接口，包括 `totalSupply`、`balanceOf`、`transfer`、`approve`、`allowance` 和 `transferFrom`。

DEX 可以通过统一接口读取代币余额、转移 token，并在用户授权后代为转账。需要注意，ERC-20 的 `name`、`symbol` 和 `decimals` 是可选元数据；调用方也不能假设所有代币都一定以相同方式处理转账。

### 2. ERC-2612：签名授权

ERC-2612 在 ERC-20 上增加 `permit`，用户可以签署一条授权消息，由任何人提交到链上来更新 `allowance`。这通常可以减少一次单独的 `approve` 交易。

`permit` 使用 EIP-712 格式的结构化签名，并包含 nonce 和截止时间等字段。它不是所有 ERC-20 都支持的功能；集成时应检查 token 是否实现该接口。签名本身也不等于转账，后续仍需要合约调用来使用授权额度。

### 3. ERC-721：非同质化代币（NFT）

ERC-721 通过 `tokenId` 区分每一个 NFT，定义所有权查询、授权和转移等接口。同一合约中的每个 token ID 对应独立资产，因此适合数字藏品、游戏道具等场景。

ERC-721 合约还需要支持 ERC-165 接口检测。将 NFT 安全转入合约时，接收合约需要实现相应的接收回调，否则转账可能失败，以减少 NFT 被转入后无法取出的情况。

### 4. ERC-1155：多代币标准

ERC-1155 允许一个合约管理多个 token ID。不同 ID 可以代表同质化资产、NFT 或介于两者之间的资产，并支持批量查询与批量转移。

与 ERC-721 每个 NFT 都有独立 token ID 的常见用法相比，ERC-1155 更方便在同一合约中管理多种游戏道具或可批量转移的资产。

### 5. ERC-165：接口检测

ERC-165 提供 `supportsInterface(bytes4)`，让合约声明自己支持哪些接口。其他合约或应用可以先查询，再决定如何与目标合约交互。

例如，NFT 市场可以查询某个 NFT 合约是否支持 ERC-721 或 ERC-2981。接口检测只表明合约声明支持某接口，不等同于对合约安全性或行为正确性的保证。

### 6. ERC-2981：NFT 版税信息

ERC-2981 定义 `royaltyInfo(tokenId, salePrice)`，用于查询某次销售对应的版税接收地址和金额。它提供的是版税信息接口，并不会自动收取版税。

是否支付版税取决于市场或交易方是否遵循该信息；普通 NFT 转账也不一定代表销售。因此，不应把 ERC-2981 理解为链上强制版税机制。

### 7. ERC-1271：智能合约钱包签名验证

EOA 可以用私钥签名；智能合约钱包没有对应的 EOA 私钥，不能用普通 EOA 验签流程验证。ERC-1271 定义 `isValidSignature(hash, signature)`，由钱包合约根据自己的规则判断签名是否有效。

这使 DEX、NFT 市场等使用离线签名的应用也可以支持智能合约钱包。ERC-1271 负责“验证签名”，不负责定义签名消息的结构；消息格式可与 EIP-712 等标准配合使用。

### 8. ERC-4626：代币化金库

ERC-4626 为管理单一 ERC-20 底层资产的金库定义通用接口。用户存入资产后取得代表金库份额的 share token；标准接口包括 `deposit`、`mint`、`withdraw`、`redeem`，以及资产和份额的换算查询。

它便于借贷、收益策略和聚合器等应用以统一方式集成金库。ERC-4626 的 share 通常是 ERC-20，但具体收益、费用和风险由金库实现决定，标准本身不保证收益。

### 9. EIP-712：结构化数据签名

EIP-712 规定如何对带有类型信息的结构化数据进行编码、哈希和签名。相较于只签一串难以理解的字节，钱包可以向用户展示更明确的字段，例如授权对象、数量和截止时间。

EIP-712 是签名数据格式，不是代币标准，也不自带通用的重放保护。应用通常还需要在签名内容中加入 nonce、链 ID、合约地址或截止时间等约束。ERC-2612 是使用 EIP-712 的常见例子。

### 10. ERC-4337：账户抽象

ERC-4337 定义一套不需要以太坊共识层修改的账户抽象机制。用户提交 `UserOperation`，Bundler 将操作打包并提交给 EntryPoint 合约；智能账户可以自定义验证和执行逻辑，Paymaster 则可在相应规则下代付 gas。

它可以支持批量操作、赞助 gas 等钱包体验，但这些能力取决于具体账户、Bundler 和 Paymaster 的实现与配置。

### 11. EIP-7702：为 EOA 委托执行代码

EIP-7702 是以太坊核心层的交易类型改进，不是 ERC 代币接口。它允许 EOA 授权某个代码地址，并让账户以委托代码的方式执行，从而支持批量操作、赞助交易等钱包体验改进。

它与 ERC-4337 是不同的机制：ERC-4337 使用 `UserOperation` 和 EntryPoint 等应用层基础设施；EIP-7702 改变了 EOA 参与交易执行的能力。两者可以结合使用，但不是同一个标准。

### 12. ERC-8004：Agent 身份、声誉与验证

ERC-8004 提出三类注册表：

- **Identity Registry**：为 Agent 提供链上身份，并关联其注册信息。
- **Reputation Registry**：发布和读取对 Agent 的反馈。
- **Validation Registry**：记录或请求第三方验证。

它主要关注 Agent 的发现和信任信息，不负责 Agent 之间的支付。官方页面当前标记为 Draft（草案），内容可能继续调整。

### 13. ERC-8183：Agent 任务交易与托管

ERC-8183 提出一种围绕任务的交易流程：客户创建并资助任务，服务提供方提交工作，评估者确认完成或拒绝；根据结果，资金支付给提供方或退还给客户。协议还描述了任务过期后的退款路径。

因此，它更准确地说是带评估流程的 Agent 商务任务托管协议，而不是泛指所有区块链结算。官方页面当前标记为 Draft（草案）。

### 14. x402：基于 HTTP 402 的网络支付

x402 是开放的互联网支付协议，不是 Ethereum ERC 或 EIP。服务端可以对资源返回 HTTP `402 Payment Required` 和支付要求；客户端提交支付信息，服务端或 Facilitator 验证并结算后，再提供资源。

它适用于 API 按次付费、内容付费和机器间支付等场景。x402 关注的是网络请求中的支付流程，可以与链上支付结合，但不等于某个特定区块链代币标准。

## 五、按应用场景理解

### DEX（去中心化交易所）

- **ERC-20**：交换同质化代币所需的通用 token 接口。
- **ERC-2612**：若 token 支持，可用签名授权减少单独的 `approve` 交易。
- **EIP-712**：用于构造可读的结构化签名，例如 permit 或订单签名。
- **ERC-721**：适用于支持 NFT 交易的市场，不是普通 ERC-20 代币兑换的必需标准。

### 借贷与金库

- **ERC-20**：用于底层资产、债务凭证或其他可替代 token。
- **ERC-2612**：可用于支持 permit 的资产授权。
- **ERC-4626**：适用于符合该标准的单资产金库，提供统一存取与份额接口。
- **EIP-712**：可用于签名授权或其他结构化消息。

### NFT 市场

- **ERC-721 / ERC-1155**：NFT 或多类型 token 的资产接口。
- **ERC-165**：检测合约声明支持的接口。
- **ERC-2981**：读取版税信息；是否实际支付由市场或交易方决定。
- **ERC-1271 / EIP-712**：分别用于智能合约钱包签名验证和结构化消息签名。

### 钱包

- **ERC-4337**：通过智能账户和 UserOperation 等基础设施提供账户抽象功能。
- **EIP-7702**：允许 EOA 委托代码执行，支持更灵活的钱包交互。

### AI 与 Web3

- **ERC-8004**：Agent 身份、声誉和验证信息。
- **x402**：通过 HTTP 支付流程实现服务或 API 付费；它不是 ERC/EIP。
- **ERC-8183**：围绕 Agent 任务的资金托管、提交、评估和结算流程。

## 六、容易混淆的地方

1. **ERC-2981 提供版税信息，不保证版税一定到账。**
2. **EIP-712 规定结构化数据签名格式，不负责定义完整的重放保护。**
3. **ERC-1271 是合约钱包的签名验证接口；ERC-2612 是 token 的签名授权扩展。**
4. **ERC-4626 的 share 表示金库份额，不代表固定收益或无风险。**
5. **ERC-4337 和 EIP-7702 都能改善账户体验，但机制不同。**
6. **ERC-8004 和 ERC-8183 当前是草案；集成前要核对最新规范。**
7. **x402 是互联网支付协议，不属于 ERC 或 EIP。**
8. **应用场景和标准不是一一绑定关系。某个项目是否支持某标准，应以其实际合约和官方文档为准。**



![image-20261111195533010](https://files.seeusercontent.com/2026/11/11/jWk8/image-20261111195533010.png)
