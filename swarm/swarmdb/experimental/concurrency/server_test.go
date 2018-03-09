package server_test

import (
	//	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/swarmdb"
	"github.com/ethereum/go-ethereum/swarmdb/server"
	"io/ioutil"
	"os"
	"testing"
)

// func NewTCPIPServer(swarmdb SwarmDB, l net.Listener) *TCPIPServer
func testTCPIPServer(t *testing.T, f func(*server.TCPIPServer)) {
	datadir, err := ioutil.TempDir("", "tcptest")
	if err != nil {
		t.Fatalf("unable to create temp dir: %v", err)
	}
	os.RemoveAll(datadir)
	defer os.RemoveAll(datadir)
	swarmdb := swarmdb.NewSwarmDB()
	svr := server.NewTCPIPServer(swarmdb, nil)
	if err != nil {
		fmt.Println("hashdb open error")
	}
	f(svr)
}
