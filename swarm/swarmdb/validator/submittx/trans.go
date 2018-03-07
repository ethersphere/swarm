package main

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"

	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	"strconv"
	// "log"
	// "strings"
	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"os"
)

type Block struct {
	Number *big.Int
}

func main() {
	// Create an IPC based RPC connection to a remote node
	conn, err := ethclient.Dial("http://10.128.0.10:8545")
	if err != nil {
		panic("Failed to connect to the Ethereum client")
	}

	client, err := rpc.Dial("http://10.128.0.10:8545")
	if err != nil {
		panic("Failed to connect to the Ethereum client")
	}
	// auth, err := bind.NewTransactor(strings.NewReader(key), "mdotm")

	// the nonce MUST be incremental (start with 1, then 2, etc.)
	var tcount string
	err = client.Call(&tcount, "eth_getTransactionCount", "0xce9510bb0d6cce1050caac4018fd3355a212ec83", "latest")
	if err != nil {
		fmt.Println("can't get latest block:", err)
		return
	}
	nonce, _ := strconv.ParseUint(tcount, 0, 64)

	/*
		var lastBlock Block
		err = client.Call(&lastBlock, "eth_getBlockByNumber", "latest", true)
		if err != nil {
			fmt.Println("can't get latest block:", err)
			return
		} else {
		  	fmt.Printf("latest block: %v\n", lastBlock.Number)
		 }
	*/
	// goal: send 3 ETH from above to Account #1: {f27c2737f8e994741c910295399c321281d0899c}
	tx1 := types.NewTransaction(nonce, common.HexToAddress("f27c2737f8e994741c910295399c321281d0899c"), big.NewInt(3000000000000000000), big.NewInt(50000), big.NewInt(4000000000), nil)
	fmt.Printf("TX1: %v", tx1)

	chainId := big.NewInt(66)
	privkey, err := crypto.HexToECDSA("cbb555249c754e8ec0488d45e1d9fa794be2953de7ec3eec89e42422684cf88d")
	if err != nil {
		fmt.Printf("key failure %s", err)
		os.Exit(0)
	}
	signer := types.NewEIP155Signer(chainId)
	tx, err := types.SignTx(tx1, signer, privkey)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	fmt.Printf("TX: %v\n", tx)
	ctx := context.Background()
	err = conn.SendTransaction(ctx, tx)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	// 	fmt.Printf("HASH: %v\n", hash)
}
