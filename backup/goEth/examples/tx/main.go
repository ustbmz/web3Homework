// 离线构造一笔 EIP-1559 转账，签名后从签名恢复发送者。交易不会被广播。
package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

func main() {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	from := crypto.PubkeyToAddress(key.PublicKey)

	// 主网 chain ID。这里只做本地签名，不连接节点，也不调用 SendTransaction。
	chainID := big.NewInt(1)
	to := common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
	value := new(big.Int).Mul(big.NewInt(params.GWei), big.NewInt(10)) // 10 gwei

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     0,
		GasTipCap: big.NewInt(params.GWei),
		GasFeeCap: big.NewInt(30 * params.GWei),
		Gas:       21_000,
		To:        &to,
		Value:     value,
	})

	signer := types.LatestSignerForChainID(chainID)
	signed, err := types.SignTx(tx, signer, key)
	if err != nil {
		log.Fatal(err)
	}

	sender, err := types.Sender(signer, signed)
	if err != nil {
		log.Fatal(err)
	}
	raw, err := signed.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("from      ", from.Hex())
	fmt.Println("to        ", signed.To().Hex())
	fmt.Println("value wei ", signed.Value())
	fmt.Println("nonce     ", signed.Nonce())
	fmt.Println("gas       ", signed.Gas())
	fmt.Println("tip cap   ", signed.GasTipCap())
	fmt.Println("fee cap   ", signed.GasFeeCap())
	fmt.Println("tx type   ", signed.Type())
	fmt.Println("tx hash   ", signed.Hash().Hex())
	fmt.Println("sender    ", sender.Hex())
	fmt.Println("match     ", sender == from)
	fmt.Println("raw tx    ", hexutil.Encode(raw))
}
