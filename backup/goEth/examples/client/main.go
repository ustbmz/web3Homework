// 通过 ethclient 读取 Ethereum 链上的公开数据。
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

const (
	defaultRPC = "https://ethereum.publicnode.com"
	// 公开地址，仅用于演示余额查询。
	sampleAccount = "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
	// WETH 合约，用于演示读取合约字节码。
	weth = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
)

func main() {
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		rpcURL = defaultRPC
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		log.Fatalf("连接 %s 失败: %v", rpcURL, err)
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	account := common.HexToAddress(sampleAccount)
	balance, err := client.BalanceAt(ctx, account, nil)
	if err != nil {
		log.Fatal(err)
	}
	code, err := client.CodeAt(ctx, common.HexToAddress(weth), nil)
	if err != nil {
		log.Fatal(err)
	}
	tip, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("rpc        ", rpcURL)
	fmt.Println("chain id   ", chainID)
	fmt.Println("block      ", header.Number)
	fmt.Println("hash       ", header.Hash().Hex())
	fmt.Println("time       ", time.Unix(int64(header.Time), 0).UTC().Format(time.RFC3339))
	fmt.Println("gas limit  ", header.GasLimit)
	fmt.Println("gas used   ", header.GasUsed)
	if header.BaseFee != nil {
		fmt.Println("base fee   ", header.BaseFee, "wei")
	}
	fmt.Println("tip cap    ", tip, "wei")
	fmt.Println("account    ", account.Hex())
	fmt.Println("balance    ", balance, "wei")
	fmt.Println("balance    ", formatEther(balance), "ether")
	fmt.Println("weth code  ", len(code), "bytes")
}

func formatEther(wei *big.Int) string {
	ether := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt64(params.Ether))
	return ether.Text('f', 6)
}
