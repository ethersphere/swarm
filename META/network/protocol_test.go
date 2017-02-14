package network

import (
	//"fmt"
	"testing"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

func init() {
	
}

func TestNewMETAAnnounce(t *testing.T) {
	a := NewMETAAnnounce()
	
	a.SetUuid(0x12345678)
	a.SetCommand(META_ANNOUNCE_IPO)
	a.GetUuid()
	a.GetCommand()
}

func TestNewMETATmpName(t *testing.T) {
	a := NewMETATmpName()
	
	a.SetUuid(0x12345678)
	a.SetCommand(META_CUSTOM)
	a.Swarm = storage.ZeroKey
	a.Name = "schmardian"
	t.Logf("%v\n", a)
}
