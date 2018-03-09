package main

import (
  "fmt"
  "log"
  "strings"
  "time"

  "github.com/ethereum/go-ethereum/accounts/abi/bind"
  "github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
  "github.com/ethereum/go-ethereum/rpc"
)

func main() {
  // Create an IPC based RPC connection to a remote node
//y  conn, err := rpc.NewHTTPClient("http://localhost:9012")
  conn, err := rpc.NewHTTPClient("http://localhost:30303")  
  if err != nil {
    log.Fatalf("Failed to connect to the Ethereum client: %v", err)
  }

  // IF YOU WANT TO DEPLOY YOURSELF
  // this is the json found in your geth chain/keystore folder
//y  key := `{"address":"f2759b4a699dae4fdc3383a0d7a92cfc246315cd","crypto":{"cipher":"aes-128-ctr","ciphertext":"a96fe235356c7ebe6520d2fa1dcc0fd67199cb490fb18c39ffabbb6880a6b3d6","cipherparams":{"iv":"47182104a4811f8da09c0bafc3743e2a"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"81c82f97edb0ee1036e63d1de57b7851271273971803e60a5cbb011e85baa251"},"mac":"09f107c9af8efcb932354d939beb7b2c0cebcfd70362d68905de554304a7cfff"},"id":"eb7ed04f-e996-4bda-893b-28dc6ac24626","version":3}`
//y  auth, err := bind.NewTransactor(strings.NewReader(key), "1234567890")

  key := `{"address":"c157ed050bdd743d8ea63ddc1dc1cd834781155b","crypto":{"cipher":"aes-128-ctr","ciphertext":"177c223b933700c75824d0cdef8c7ffa38eada6d73aa174ba71346c8280dce9e","cipherparams":{"iv":"8dbbf390616cf60ccc52da2f87e25199"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"08332d1eeac3d132ada9e2f765bf746e36351719f3e0f52b41db223ec6fa98fc"},"mac":"821e377671a495788785ece6a1a20a048657cb257702c99ecbf4ba0049825767"},"id":"2187606c-cd09-4879-9e8a-23bd8d0049f0","version":3}`
  auth, err := bind.NewTransactor(strings.NewReader(key), "mdotm")
  if err != nil {
    log.Fatalf("Failed to create authorized transactor: %v", err)
  }
  // Deploy a new awesome contract for the binding demo
  triggerAddr, _, trigger, err := DeployTrigger(auth, backends.NewRPCBackend(conn))
  if err != nil {
    log.Fatalf("Failed to deploy new trigger contract: %v", err)
  }
  // Don't even wait, check its presence in the local pending state
  time.Sleep(5 * time.Second) // Allow it to be processed by the local node :P
  // END IF YOU WANT TO DEPLOY YOURSELF

  // IF YOU HAVE ALREADY DEPLOYED IT
  // deployedTriggerAddr := "0xe2359b4a699dae4fdc3383a0d7a92cfc246315ce"
  deployedTriggerAddr := triggerAddr
  trigger, err = NewTrigger(deployedTriggerAddr, backends.NewRPCBackend(conn))
  if err != nil {
    log.Fatalf("Failed to instantiate a trigger contract: %v", err)
  }
  // END IF YOU HAVE ALREADY DEPLOYED IT

  owner, err := trigger.GetOwner(nil)
  if err != nil {
    log.Fatalf("Failed to retrieve token name: %v", err)
  }
  fmt.Printf("owner address: 0x%x\n", owner)
}