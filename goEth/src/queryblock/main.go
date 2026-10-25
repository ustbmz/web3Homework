// 查询 Sepolia 测试网指定区块的哈希、时间戳和交易数量。
//
//	go run ./src/queryblock -block 5000000
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"time"

	"goeth/src/queryblock/rpc"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	blockFlag := flag.Int64("block", 11988118, "Sepolia block number to query")
	flag.Parse()

	rpcURL, _, err := rpc.RPC()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		log.Fatalf("dial Sepolia: %v", err)
	}
	defer client.Close()

	header, err := client.HeaderByNumber(ctx, big.NewInt(*blockFlag))
	if err != nil {
		log.Fatal(err)
	}
	txCount, err := client.TransactionCount(ctx, header.Hash())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("block     ", header.Number)
	fmt.Println("hash      ", header.Hash().Hex())
	fmt.Println("timestamp ", header.Time)
	fmt.Println("time      ", time.Unix(int64(header.Time), 0).UTC().Format(time.RFC3339))
	fmt.Println("tx count  ", txCount)
}
