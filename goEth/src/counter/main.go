// 使用 abigen 生成的绑定，在 Sepolia 上部署 Counter 并调用 increment。
//
//	export SEPOLIA_PRIVATE_KEY=发送方私钥
//	go run ./src/counter
//
// 已有合约时设置 CONTRACT_ADDRESS，程序会跳过部署，只调用 increment。
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

	"goeth/src/counter/binding"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const sepoliaChainID int64 = 11155111

func main() {
	privateKey, from, err := loadKey(os.Getenv("SEPOLIA_PRIVATE_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	client, err := ethclient.DialContext(ctx, sepoliaRPC())
	if err != nil {
		log.Fatalf("dial Sepolia: %v", err)
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if chainID.Cmp(big.NewInt(sepoliaChainID)) != 0 {
		log.Fatalf("connected to chain id %s, not Sepolia (%d)", chainID, sepoliaChainID)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal(err)
	}
	auth.Context = ctx
	tip, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatal(err)
	}
	head, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if head.BaseFee == nil {
		log.Fatal("node did not return a base fee")
	}
	auth.GasTipCap = tip
	auth.GasFeeCap = new(big.Int).Add(new(big.Int).Mul(head.BaseFee, big.NewInt(2)), tip)

	instance, address, err := connect(ctx, client, auth)
	if err != nil {
		log.Fatal(err)
	}

	before, err := instance.Count(&bind.CallOpts{Context: ctx})
	if err != nil {
		log.Fatal(err)
	}
	tx, err := instance.Increment(auth)
	if err != nil {
		log.Fatal(err)
	}
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		log.Fatal(err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Fatalf("increment transaction failed: %s", tx.Hash().Hex())
	}
	after, err := instance.Count(&bind.CallOpts{Context: ctx, BlockNumber: receipt.BlockNumber})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("account   ", from.Hex())
	fmt.Println("contract  ", address.Hex())
	fmt.Println("tx hash   ", tx.Hash().Hex())
	fmt.Println("block     ", receipt.BlockNumber)
	fmt.Println("before    ", before)
	fmt.Println("after     ", after)
}

func connect(ctx context.Context, client *ethclient.Client, auth *bind.TransactOpts) (*counter.Counter, common.Address, error) {
	raw := strings.TrimSpace(os.Getenv("CONTRACT_ADDRESS"))
	if raw != "" {
		if !common.IsHexAddress(raw) {
			return nil, common.Address{}, fmt.Errorf("CONTRACT_ADDRESS is not a valid hex address")
		}
		address := common.HexToAddress(raw)
		instance, err := counter.NewCounter(address, client)
		return instance, address, err
	}

	address, tx, instance, err := counter.DeployCounter(auth, client)
	if err != nil {
		return nil, common.Address{}, err
	}
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, common.Address{}, err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, common.Address{}, fmt.Errorf("deploy transaction failed: %s", tx.Hash().Hex())
	}
	fmt.Println("deploy tx ", tx.Hash().Hex())
	return instance, address, nil
}

func sepoliaRPC() string {
	if explicit := os.Getenv("SEPOLIA_RPC_URL"); explicit != "" {
		return explicit
	}
	if key := os.Getenv("INFURA_API_KEY"); key != "" {
		return "https://sepolia.infura.io/v3/" + key
	}
	return "https://ethereum-sepolia-rpc.publicnode.com"
}

func loadKey(raw string) (*ecdsa.PrivateKey, common.Address, error) {
	raw = strings.TrimSpace(strings.TrimPrefix(raw, "0x"))
	if raw == "" {
		return nil, common.Address{}, fmt.Errorf("SEPOLIA_PRIVATE_KEY is required")
	}
	key, err := crypto.HexToECDSA(raw)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("parse SEPOLIA_PRIVATE_KEY: %w", err)
	}
	return key, crypto.PubkeyToAddress(key.PublicKey), nil
}
