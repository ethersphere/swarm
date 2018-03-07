package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Create an IPC based RPC connection to a remote node
    //y	conn, err := ethclient.Dial("/home/karalabe/.ethereum/testnet/geth.ipc")
    conn, err := ethclient.Dial("/var/www/vhosts/data/geth.ipc")  // this is working OK
    // conn, err := ethclient.Dial("http://127.0.0.1:8545")       // this is working OK	 //  JSON-RPC Endpoint   https://github.com/ethereum/wiki/wiki/JSON-RPC   see comment below
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	// Instantiate the contract and display its name
	greeter, err := NewGreeter(common.HexToAddress("0x159e7021a334ab47b20871e7f01e028bd70ad15b"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a greeter contract: %v", err)
	}
	Greet, err := greeter.Greet(nil)
	if err != nil {
		log.Fatalf("Failed to retrieve Greet: %v", err)
	}
	fmt.Println("Greet:", Greet)
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