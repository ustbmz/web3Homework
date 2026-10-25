# Counter：abigen 绑定

用 Solidity 写一个计数器，编译出 ABI 和字节码，再用 `abigen` 生成 Go 绑定，在 Sepolia 上部署并调用 `increment`。

## 目录

| 路径 | 作用 |
| --- | --- |
| `contracts/Counter.sol` | 计数器合约，`count` 可读，`increment` 加一 |
| `build/Counter.abi` | 编译得到的 ABI |
| `build/Counter.bin` | 编译得到的字节码 |
| `binding/counter.go` | `abigen` 生成的 Go 绑定，不要手改 |
| `main.go` | 部署合约、调用 `increment`、打印调用前后的计数值 |

合约用 solc 0.8.30 编译。`abigen` 版本与项目依赖一致，是 go-ethereum v1.17.10。

## 重新生成绑定

安装工具：

```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@v1.17.10
```

已安装 solc 时，在项目根目录执行：

```bash
solc --abi --bin src/counter/contracts/Counter.sol -o src/counter/build --overwrite
abigen \
  --abi src/counter/build/Counter.abi \
  --bin src/counter/build/Counter.bin \
  --pkg counter \
  --type Counter \
  --out src/counter/binding/counter.go
```

`solc --bin` 生成的文件名可能带上目录前缀，需要把 ABI 和字节码整理成 `Counter.abi`、`Counter.bin` 再交给 `abigen`。

## 在 Sepolia 上调用

发送方账户需要有测试币。私钥只放在环境变量里。

```bash
export SEPOLIA_PRIVATE_KEY=发送方私钥
go run ./src/counter
```

程序会先部署合约，再调用 `increment`。成功时打印合约地址、交易哈希，以及调用前、调用后的 `count`。

同一份合约上继续加一：

```bash
export CONTRACT_ADDRESS=0x上一次打印的合约地址
go run ./src/counter
```
