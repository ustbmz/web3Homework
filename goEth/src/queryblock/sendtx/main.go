// 在 Sepolia 上签名并发送一笔以太币转账，然后打印交易哈希。
//
//	export INFURA_API_KEY=你的 Infura Sepolia API Key
//	export SEPOLIA_PRIVATE_KEY=发送方私钥（可带 0x 前缀）
//	export TO_ADDRESS=0x接收方地址
//	export AMOUNT_ETH=0.001
//	go run ./src/queryblock/sendtx
//
// 私钥只从环境变量读取，不会写入代码或打印出来。
// 账户需要有 Sepolia ETH，可从 https://sepoliafaucet.com 领取。
package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"goeth/src/queryblock/rpc"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

func main() {
	privateKey, from, err := loadKey(os.Getenv("SEPOLIA_PRIVATE_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	to, err := parseAddress(os.Getenv("TO_ADDRESS"))
	if err != nil {
		log.Fatal(err)
	}
	amount := os.Getenv("AMOUNT_ETH")
	if amount == "" {
		amount = "0.001"
	}
	value, err := etherToWei(amount)
	if err != nil {
		log.Fatal(err)
	}

	rpcURL, viaInfura, err := rpc.RPC()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		log.Fatalf("dial Sepolia: %v", err)
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if chainID.Cmp(big.NewInt(rpc.ChainID)) != 0 {
		log.Fatalf("connected to chain id %s, not Sepolia (%d); transaction not sent", chainID, rpc.ChainID)
	}

	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		log.Fatal(err)
	}
	tip, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatal(err)
	}
	head, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if head.BaseFee == nil {
		log.Fatal("node did not return a base fee; cannot build an EIP-1559 transaction")
	}
	feeCap := new(big.Int).Add(new(big.Int).Mul(head.BaseFee, big.NewInt(2)), tip)

	balance, err := client.BalanceAt(ctx, from, nil)
	if err != nil {
		log.Fatal(err)
	}
	cost := new(big.Int).Add(value, new(big.Int).Mul(feeCap, big.NewInt(21_000)))
	if balance.Cmp(cost) < 0 {
		log.Fatalf("insufficient funds: %s has %s wei, this transfer needs up to %s wei including gas", from.Hex(), balance, cost)
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tip,
		GasFeeCap: feeCap,
		Gas:       21_000,
		To:        &to,
		Value:     value,
	})
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	if err := client.SendTransaction(ctx, signed); err != nil {
		log.Fatal(err)
	}

	endpoint := rpcURL
	if viaInfura {
		endpoint = "https://sepolia.infura.io/v3/<INFURA_API_KEY>"
	}
	fmt.Println("network   ", "Sepolia")
	fmt.Println("rpc       ", endpoint)
	fmt.Println("from      ", from.Hex())
	fmt.Println("to        ", to.Hex())
	fmt.Println("amount    ", amount, "ETH")
	fmt.Println("nonce     ", nonce)
	fmt.Println("tx hash   ", signed.Hash().Hex())
}

func loadKey(raw string) (*ecdsa.PrivateKey, common.Address, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "0x")
	if raw == "" {
		return nil, common.Address{}, fmt.Errorf("SEPOLIA_PRIVATE_KEY is required")
	}
	key, err := crypto.HexToECDSA(raw)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("parse SEPOLIA_PRIVATE_KEY: %w", err)
	}
	return key, crypto.PubkeyToAddress(key.PublicKey), nil
}

func parseAddress(raw string) (common.Address, error) {
	raw = strings.TrimSpace(raw)
	if !common.IsHexAddress(raw) {
		return common.Address{}, fmt.Errorf("TO_ADDRESS must be a valid hex address")
	}
	return common.HexToAddress(raw), nil
}

// etherToWei converts a decimal ETH amount to wei with integer arithmetic.
// big.Float's default 64-bit mantissa cannot represent values such as 0.001
// exactly; multiplying by 1e18 and truncating with Int drops 1 wei on chain.
func etherToWei(amount string) (*big.Int, error) {
	whole, frac, ok := splitEther(strings.TrimSpace(amount))
	if !ok {
		return nil, fmt.Errorf("AMOUNT_ETH must be a number greater than 0")
	}
	if len(frac) > 18 {
		if strings.TrimRight(frac[18:], "0") != "" {
			return nil, fmt.Errorf("AMOUNT_ETH must be an integer number of wei (at most 18 decimal places)")
		}
		frac = frac[:18]
	}
	frac += strings.Repeat("0", 18-len(frac))

	w, ok := new(big.Int).SetString(whole, 10)
	if !ok {
		return nil, fmt.Errorf("AMOUNT_ETH must be a number greater than 0")
	}
	f, ok := new(big.Int).SetString(frac, 10)
	if !ok {
		return nil, fmt.Errorf("AMOUNT_ETH must be a number greater than 0")
	}
	wei := new(big.Int).Add(new(big.Int).Mul(w, big.NewInt(params.Ether)), f)
	if wei.Sign() <= 0 {
		return nil, fmt.Errorf("AMOUNT_ETH must be a number greater than 0")
	}
	return wei, nil
}

func splitEther(amount string) (whole, frac string, ok bool) {
	if amount == "" || amount[0] == '-' || amount[0] == '+' {
		return "", "", false
	}
	whole, frac, hasFrac := strings.Cut(amount, ".")
	if hasFrac && strings.Contains(frac, ".") {
		return "", "", false
	}
	if whole == "" {
		whole = "0"
	}
	if !digitsOnly(whole) || !digitsOnly(frac) {
		return "", "", false
	}
	return whole, frac, true
}

func digitsOnly(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
