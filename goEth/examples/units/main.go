// 用 go-ethereum 的面额常量在 wei、gwei、ether 之间换算。
package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

func main() {
	oneEther := big.NewInt(params.Ether)
	oneGwei := big.NewInt(params.GWei)

	fmt.Println("1 ether =", oneEther, "wei")
	fmt.Println("1 gwei  =", oneGwei, "wei")
	fmt.Println("1 ether =", new(big.Int).Quo(oneEther, oneGwei), "gwei")

	fee := new(big.Int).Mul(big.NewInt(21_000), big.NewInt(30*params.GWei))
	fmt.Println("21000 gas * 30 gwei =", formatEther(fee), "ether")

	value, ok := new(big.Float).SetString("1.5")
	if !ok {
		panic("invalid ether amount")
	}
	fmt.Println("1.5 ether =", etherToWei(value), "wei")
}

func etherToWei(ether *big.Float) *big.Int {
	wei := new(big.Float).Mul(ether, new(big.Float).SetInt64(params.Ether))
	result, _ := wei.Int(nil)
	return result
}

func formatEther(wei *big.Int) string {
	ether := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt64(params.Ether))
	return ether.Text('f', 6)
}
