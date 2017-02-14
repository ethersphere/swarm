package main 

import (
	"runtime"
	
	p2psimulations "github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/META/network/simulations"
)

// var server
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	c, quitc := simulations.NewSessionController()
	
	p2psimulations.StartRestApiServer("8888", c)
	// wait until server shuts down
	<-quitc

}
