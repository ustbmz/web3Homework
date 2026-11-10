// 演示合约调用数据：计算函数选择器，并编码、解码 ERC-20 balanceOf。
package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const erc20ABI = `[
  {
    "name": "balanceOf",
    "type": "function",
    "stateMutability": "view",
    "inputs": [{"name": "account", "type": "address"}],
    "outputs": [{"name": "balance", "type": "uint256"}]
  }
]`

func main() {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatal(err)
	}

	account := common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
	calldata, err := parsed.Pack("balanceOf", account)
	if err != nil {
		log.Fatal(err)
	}

	selector := crypto.Keccak256([]byte("balanceOf(address)"))[:4]
	fmt.Println("selector  ", hexutil.Encode(selector))
	fmt.Println("calldata  ", hexutil.Encode(calldata))
	fmt.Println("selector matches keccak prefix:", hexutil.Encode(calldata[:4]) == hexutil.Encode(selector))

	// 模拟节点返回的 ABI 编码结果：一个左填充的 uint256。
	returned := new(big.Int)
	returned.SetString("1500000000000000000", 10)
	word := common.LeftPadBytes(returned.Bytes(), 32)

	values, err := parsed.Unpack("balanceOf", word)
	if err != nil {
		log.Fatal(err)
	}
	balance := values[0].(*big.Int)
	fmt.Println("decoded   ", balance)
	fmt.Println("roundtrip ", balance.Cmp(returned) == 0)
}
