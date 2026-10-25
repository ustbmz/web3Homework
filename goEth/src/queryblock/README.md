# Sepolia 区块查询与转账

在 Sepolia 测试网（chain ID `11155111`）上查询区块，并发送一笔以太币转账。

## 目录

| 路径 | 作用 |
| --- | --- |
| `main.go` | 查询指定区块的哈希、时间戳、交易数量 |
| `rpc/rpc.go` | 根据 `INFURA_API_KEY` 或 `SEPOLIA_RPC_URL` 得到节点地址 |
| `sendtx/main.go` | 签名并广播一笔转账，打印交易哈希 |

## 环境

- Go 1.25 及以上
- 依赖 `github.com/ethereum/go-ethereum`，在项目根目录执行 `go mod download`
- [Infura](https://infura.io) 的 Sepolia API Key。没有 Key 时，查询会改用公共节点

发送交易还需要：

- 发送方私钥，只放在环境变量 `SEPOLIA_PRIVATE_KEY` 里
- 接收方完整的 `0x` 地址
- 发送方地址上要有 Sepolia 测试币，可从 [sepoliafaucet.com](https://sepoliafaucet.com) 领取

这把私钥同样能控制该地址在以太坊主网上的资产。请使用空的测试账户，不要把私钥写进代码或提交到仓库。

## 查询区块

在项目根目录执行：

```bash
go run ./src/queryblock -block 5000000
```

`-block` 的默认值是 `5000000`。在 `src/queryblock` 目录里运行时，把路径改成 `.`。

## 发送交易

```bash
export INFURA_API_KEY=你的key
export SEPOLIA_PRIVATE_KEY=发送方私钥
export TO_ADDRESS=0x接收方完整地址
export AMOUNT_ETH=0.001
go run ./src/queryblock/sendtx
```

`AMOUNT_ETH` 可以省略，默认转 `0.001` ETH。程序会核对 chain ID，余额不够时不会发送。成功后控制台输出 `tx hash`。

![微信图片_20261025114437_301_8](https://files.seeusercontent.com/2026/10/25/q8Ao/_20261025114437_301_8.png)
