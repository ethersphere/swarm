package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
//y	"github.com/ethereum/go-ethereum/rpc" //y add this
)

//y const key = `paste the contents of your *testnet* key json here`
const key = `{"address":"c157ed050bdd743d8ea63ddc1dc1cd834781155b","crypto":{"cipher":"aes-128-ctr","ciphertext":"177c223b933700c75824d0cdef8c7ffa38eada6d73aa174ba71346c8280dce9e","cipherparams":{"iv":"8dbbf390616cf60ccc52da2f87e25199"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"08332d1eeac3d132ada9e2f765bf746e36351719f3e0f52b41db223ec6fa98fc"},"mac":"821e377671a495788785ece6a1a20a048657cb257702c99ecbf4ba0049825767"},"id":"2187606c-cd09-4879-9e8a-23bd8d0049f0","version":3}`

func main() {
	// Create an IPC based RPC connection to a remote node and an authorized transactor
//y	conn, err := rpc.NewIPCClient("/home/karalabe/.ethereum/testnet/geth.ipc")
	
//y	conn, err := rpc.NewIPCClient("/var/www/vhosts/data/geth.ipc")
	conn, err := ethclient.Dial("http://10.128.0.13:30303")
	
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	auth, err := bind.NewTransactor(strings.NewReader(key), "mdotm")
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	
	//fmt.Printf("auth :%v\n\n", auth)
	
	// Deploy a new awesome contract for the binding demo
	address, tx, token, err := DeployToken(auth, conn, new(big.Int), "Contracts in Go!!!", 0, "Go!")
	if err != nil {
		log.Fatalf("Failed to deploy new token contract: %v", err)
	}
	fmt.Printf("Contract pending deploy: 0x%x\n", address)
	fmt.Printf("Transaction waiting to be mined: 0x%x\n\n", tx.Hash())

	// Don't even wait, check its presence in the local pending state
	time.Sleep(250 * time.Millisecond) // Allow it to be processed by the local node :P

	name, err := token.Name(&bind.CallOpts{Pending: true})
	if err != nil {
		log.Fatalf("Failed to retrieve pending name: %v", err)
	}
	fmt.Println("Pending name:", name)
}