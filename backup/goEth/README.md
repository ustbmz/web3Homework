# goeth

用 [go-ethereum](https://github.com/ethereum/go-ethereum)（Geth v1.17）连接 Sepolia 测试网的两个小项目，以及一组本地示例。

## 两个项目

| 目录 | 介绍 |
| --- | --- |
| `src/queryblock` | 查询指定区块的哈希、时间戳和交易数量，并签名发送一笔以太币转账 |
| `src/counter` | 用 Solidity 写计数器，经 `abigen` 生成 Go 绑定后，在 Sepolia 上部署并调用 `increment` |

两个项目都从环境变量读取私钥和节点地址，说明分别在各自目录的 `README.md`。

```bash
go run ./src/queryblock -block 5000000
go run ./src/queryblock/sendtx
go run ./src/counter
```

## 运行

```bash
go run ./examples/account
go run ./examples/units
go run ./examples/tx
go run ./examples/abi
go run ./examples/client
```

`examples/client` 通过 JSON-RPC 读取链上数据。默认使用公共节点 `https://ethereum.publicnode.com`，也可以换成自己的节点：

```bash
ETH_RPC_URL=http://127.0.0.1:8545 go run ./examples/client
```

## 示例

| 目录 | 内容 |
| --- | --- |
| `examples/account` | 生成 secp256k1 密钥、推导地址、按 Ethereum 个人消息规则签名并恢复签名者 |
| `examples/units` | 用 `params` 在 wei、gwei、ether 之间换算 |
| `examples/tx` | 构造 EIP-1559 交易，本地签名后从签名恢复发送者 |
| `examples/abi` | 计算函数选择器，并编码 / 解码 ERC-20 `balanceOf` |
| `examples/client` | 查询 chain ID、最新区块头、账户余额和合约字节码 |
