package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	// Generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	sim := backends.NewSimulatedBackend(core.GenesisAccount{Address: auth.From, Balance: big.NewInt(10000000000)})

	// Deploy a token contract on the simulated blockchain
	_, _, token, err := DeployMyToken(auth, sim, new(big.Int), "Simulated blockchain tokens", 0, "SBT")
	if err != nil {
		log.Fatalf("Failed to deploy new token contract: %v", err)
	}
	// Print the current (non existent) and pending name of the contract
	name, _ := token.Name(nil)
	fmt.Println("Pre-mining name:", name)

	name, _ = token.Name(&bind.CallOpts{Pending: true})
	fmt.Println("Pre-mining pending name:", name)

	// Commit all pending transactions in the simulator and print the names again
	sim.Commit()

	name, _ = token.Name(nil)
	fmt.Println("Post-mining name:", name)

	name, _ = token.Name(&bind.CallOpts{Pending: true})
	fmt.Println("Post-mining pending name:", name)
}