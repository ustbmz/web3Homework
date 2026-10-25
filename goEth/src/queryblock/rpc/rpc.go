// Package rpc 提供 Sepolia 测试网的 RPC 地址。
// 优先使用 INFURA_API_KEY，对应地址为 https://sepolia.infura.io/v3/<key>。
// 没有 Key 时使用公共节点，方便先把查询跑通。
package rpc

import (
	"fmt"
	"os"
)

const (
	// ChainID 是 Sepolia 的链 ID。发送交易前必须核对，避免打到主网。
	ChainID int64 = 11155111

	publicRPC = "https://ethereum-sepolia-rpc.publicnode.com"
	infuraURL = "https://sepolia.infura.io/v3/%s"
)

// RPC 返回要连接的 Sepolia 节点地址，以及是否包含 Infura Key。
// 含 Key 的完整地址不要打印到日志里。
func RPC() (url string, viaInfura bool, err error) {
	if explicit := os.Getenv("SEPOLIA_RPC_URL"); explicit != "" {
		return explicit, false, nil
	}
	if key := os.Getenv("INFURA_API_KEY"); key != "" {
		return fmt.Sprintf(infuraURL, key), true, nil
	}
	return publicRPC, false, nil
}
