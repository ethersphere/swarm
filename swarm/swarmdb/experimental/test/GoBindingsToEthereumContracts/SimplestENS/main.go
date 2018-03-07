package main

import (
	"fmt"
	"log"
	"strings"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func main() {
	// Create an IPC based RPC connection to a remote node
    //y	conn, err := ethclient.Dial("/home/karalabe/.ethereum/testnet/geth.ipc")
  
    //conn, err := ethclient.Dial("/var/www/vhosts/data/geth.ipc")  // this is working OK
    conn, err := ethclient.Dial("http://127.0.0.1:8545")            // this is working OK  //  JSON-RPC Endpoint   https://github.com/ethereum/wiki/wiki/JSON-RPC see comment below
	//conn, err := ethclient.Dial("http://35.224.194.195:8545")
	//conn, err := ethclient.Dial("http://ens.wolk.com:8545")
	
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	var key = `{"address":"90fb0de606507e989247797c6a30952cae4d5cbe","crypto":{"cipher":"aes-128-ctr","ciphertext":"54396d6ed0335e4b4874cd4440d24eabeca895fcbafb15d310c25c6b1e4bb306","cipherparams":{"iv":"e3a2457cf8420d3072e5adf118d31df8"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"d25987f2f2429e53f51d87eb6474e3f12a67c63603fd860b558657cee19a6ea9"},"mac":"023fc8a29a6e323db43e0c7795d2d59d0c1f295a62cbb9bc625951fca9c385dd"},"id":"dc849ada-c6be-4f12-bfa2-5200ec560c2e","version":3}`
	auth, err := bind.NewTransactor(strings.NewReader(key), "mdotm")
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	// Instantiate the contract and display its name
	sens, err := NewSimplestens(common.HexToAddress("0x6120c3f1fdcd20c384b82eb20d93eef7838e0363"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Simplestens contract: %v", err)
	}
	
	b, err := hex.DecodeString("9f5cd92e2589fadd191e7e7917b9328d03dc84b7a67773db26efb7d0a4635677")
	if err != nil {
		log.Fatalf("Failed to hexify %v", err)
	}
	var b2 [32]byte;
	copy(b2[0:], b)
	//s, err := sens.Content(b)
	s, err := sens.Content(nil, b2)
	if err != nil {
		log.Fatalf("Failed to retrieve Greet: %v", err)
	}
	fmt.Printf("b:%x l:%d b2: %x => %x\n", b, len(b), b2, s);
	
	
	/*node:= ensNode("yaron.eth")
	h: = common.NewHashFromHex("b067fdca3d36f81af079485d443e8db9b2ac561dc6be5faf4a650f193f6a3004")
	h2 := common.SetBytes(h)
	s1, err2 := sens.SetContent(auth, node, h2 )
	*/
	s1, err2 := sens.SetContent(auth, b2, b2)
	if err2 != nil {
		log.Fatalf("Failed to set Content: %v", err2)
	}
	fmt.Printf("Transfer pending: 0x%x\n", s1.Hash())	
	//fmt.Printf("txn hash: %x", s1)
}

/*
for 
      conn, err := ethclient.Dial("http://127.0.0.1:8545")  
	//conn, err := ethclient.Dial("http://35.224.194.195:8545")
	//conn, err := ethclient.Dial("http://ens.wolk.com:8545")
to work

1) geth rpc must be enabled
--rpc \
--rpcaddr 0.0.0.0 \

2) port 8545  must be enabled   


3) and geth call ->

nohup geth --bootnodes enode://f5e184262f11afe7f2fdb636afb6980cbbc8426a2227199e640fb6d1de0c7856f00e062088618d77b7dc87bfbc3ad3649751aa53b821e4234691cc7ad4d184e7@10.128.0.3:30301 \
--identity WolkMainNode \
--datadir /var/www/vhosts/data \
--mine \
--unlock 0 \
--password <(echo -n "mdotm") \
--verbosity 6 \
--networkid 55501 \
--rpc \
--rpcaddr 0.0.0.0 \
2>> /var/www/vhosts/data/geth.log &




*/