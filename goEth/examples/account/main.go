// 离线演示 Ethereum 账户：生成密钥、推导地址，并完成个人消息的签名与恢复。
package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	address := crypto.PubkeyToAddress(key.PublicKey)
	fmt.Println("address    ", address.Hex())
	fmt.Println("public key ", hexutil.Encode(crypto.FromECDSAPub(&key.PublicKey)))
	fmt.Println("private key", hexutil.Encode(crypto.FromECDSA(key)))
	fmt.Println("这把密钥只在本次进程内存在，没有对应的链上余额。")

	message := []byte("hello, ethereum")
	hash := personalHash(message)
	signature, err := crypto.Sign(hash.Bytes(), key)
	if err != nil {
		log.Fatal(err)
	}

	recovered, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		log.Fatal(err)
	}
	recoveredAddress := crypto.PubkeyToAddress(*recovered)

	fmt.Println("message    ", string(message))
	fmt.Println("msg hash   ", hash.Hex())
	fmt.Println("signature  ", hexutil.Encode(signature))
	fmt.Println("recovered  ", recoveredAddress.Hex())
	fmt.Println("match      ", recoveredAddress == address)
}

// personalHash 实现 eth_sign / personal_sign 使用的前缀哈希。
func personalHash(message []byte) common.Hash {
	prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	return crypto.Keccak256Hash([]byte(prefixed))
}
