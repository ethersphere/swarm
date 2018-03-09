package swarmdblog_test

import (
	"github.com/ethereum/go-ethereum/swarmdb/log"
	"testing"
)

func TestLogDebug(t *testing.T) {
	l := swarmdblog.NewLogger()
	err := l.Debug("Debug string")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogTrace(t *testing.T) {
	l := swarmdblog.NewLogger()
	err := l.Trace("Trace string")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogCloud(t *testing.T) {
	l := swarmdblog.NewLogger()
	err := l.Cloud("Cloud string")
	if err != nil {
		t.Fatal(err)
	}

}

func TestLogMining(t *testing.T) {
	l := swarmdblog.NewLogger()
	err := l.Mining("Mining string")
	if err != nil {
		t.Fatal(err)
	}

}

func TestLogNetstats(t *testing.T) {
	l := swarmdblog.NewLogger()
	err := l.Netstats("Netstats string")
	if err != nil {
		t.Fatal(err)
	}

}
func TestLogTCP(t *testing.T) {
	l := swarmdblog.NewLogger()
	err := l.TCP("TCP string")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogHTTP(t *testing.T) {
	l := swarmdblog.NewLogger()
	err := l.HTTP("HTTP string")
	if err != nil {
		t.Fatal(err)
	}
}



