## **合约实战学习**

### **1. MetaNodeStake**

项目仓库：

https://github.com/MetaNodeAcademy/Advanced2-contract-stake/tree/main/stake-contract

项目体验地址：  
https://meta-node-stake-fe.vercel.app

### **任务要求**

#### **第一阶段：项目分析**

1. 阅读并理解 MetaNodeStake 项目的整体业务逻辑。
2. 梳理项目中的核心合约及其职责。
3. 重点理解以下内容：
    - 质押（Stake）流程
    - 提现（Withdraw）流程
    - 奖励计算逻辑
    - 权限控制机制
    - 数据存储结构

#### **第二阶段：Foundry 迁移**

1. 将项目从 Hardhat 2 升级至 **Foundry**。
2. 完成项目依赖升级及配置调整。
3. 确保项目能够正常完成以下操作：
    - 编译（Compile）
    - 测试（Test）
    - 部署（Deploy）
4. 记录迁移过程中遇到的问题及解决方案。

#### **第三阶段：补充测试用例(Foundry)**

1. 使用测试覆盖率工具统计项目测试覆盖情况。
2. 尽可能提升整体测试覆盖率。
3. 目标覆盖率≥ 80%

---

### **2. MetaNodeSwap**

学完Uniswap V2跟Uniswap V3再学MetaNodeSwap

项目仓库：
 https://github.com/MetaNodeAcademy/Base2_Solidity_Dex/tree/main/swap-contract

项目体验地址：
 https://swap.metanode.tech

任务要求：

1. 阅读并理解 MetaNodeSwap 项目的测试用例。
2. 梳理合约的核心业务逻辑。
3. 结合 Uniswap V2 / V3 的基础知识，理解该项目作为简化版 V3 的设计思路。

---

#### **Uniswap V2 学习要求**

推荐课程：
 https://space.bilibili.com/220951871/lists/5076602?type=season

参考资料：

- https://github.com/MetaNodeAcademy/Base2_Solidity_Dex/blob/main/%E7%B2%BE%E9%80%9AUniswap.md
- https://github.com/MetaNodeAcademy/Base2_Solidity_Dex/blob/main/Uniswap%20V2%20%E6%95%B0%E5%AD%A6%E4%B8%8E%E6%BA%90%E7%A0%81%E8%A7%A3%E8%AF%BB.md
- https://github.com/Uniswap/v2-core
- https://github.com/Uniswap/v2-periphery

需要掌握的问题：

1. 什么是 AMM？它和订单簿模式有什么区别？
2. AMM 的核心算法是什么？
3. 什么是滑点？用户交易时如何避免滑点过高？
4. 什么是三明治攻击？如何降低被三明治攻击的风险？
5. LP 的作用是什么？
6. 什么是无常损失？如何计算？
7. 阅读 Uniswap V2 Core 和 Periphery 合约，整理一份核心接口文档。

---

#### **Uniswap V3 学习要求**

参考资料：
 https://github.com/Dapp-Learning-DAO/Dapp-Learning/blob/main/defi/Uniswap-V3/whitepaperGuide/understandV3Witepaper.md

V3 相比 V2 有较大的升级，重点在于集中流动性、灵活费率和 tick 机制。

需要掌握的问题：

1. 集中流动性的优势是什么？
2. 集中流动性在合约层面大致是如何实现的？
3. tick 和价格之间是什么关系？
4. 为什么 V3 需要 tick？
5. tick spacing 和手续费费率之间有什么关系？
6. 为什么不同费率池会对应不同的 tick spacing？

### **3、熟悉常用的ERC（EIP）标准，整理文档**

DEX（Uniswap）：ERC-20, ERC-2612, ERC-721, EIP-712；
借贷（Aave）：ERC-20, ERC-2612, ERC-4626, EIP-712；
NFT（OpenSea）：ERC-721, ERC-1155, ERC-165, ERC-2981, ERC-1271, EIP-712；
钱包（MetaMask）：ERC-4337，EIP-7702；
AI+Web3：ERC-8004（Agent），X402（支付）,ERC-8183（结算）；

官方查询链接：

https://eips.ethereum.org/all

https://github.com/ethereum/ERCs

https://github.com/ethereum/EIPs

#### **提交要求**

将EIP含义整理成 Markdown 文档发到群里。