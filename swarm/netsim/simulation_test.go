// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package netsim

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	colorable "github.com/mattn/go-colorable"
)

var (
	loglevel = flag.Int("loglevel", 2, "verbosity of logs")
)

func init() {
	flag.Parse()
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

// noopService is the service that does not do anything
// but implemnts node.Service interface.
type noopService struct{}

func newNoopService() node.Service {
	return &noopService{}
}

func (t *noopService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

func (t *noopService) APIs() []rpc.API {
	return []rpc.API{}
}

func (t *noopService) Start(server *p2p.Server) error {
	return nil
}

func TestSimulationWithHTTPServer(t *testing.T) {
	log.Debug("Init simulation")
	sim := NewSimulation(
		map[string]ServiceFunc{
			"noop": func(_ *adapters.ServiceContext, b *sync.Map) (node.Service, func(), error) {
				return newNoopService(), nil, nil
			},
		},
		&SimulationOptions{
			WithHTTP: true,
		},
	)
	defer sim.Close()
	log.Debug("Done.")

	_, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	log.Debug("Starting sim round and let it time out...")
	//first test that running without sending to the channel will actually
	//block the simulation, so let it time out
	result := sim.Run(ctx, func(ctx context.Context, sim *Simulation) error {
		log.Debug("Just start the sim without any action and wait for the timeout")
		//ensure with a Sleep that simulation doesn't terminate before the timeout
		time.Sleep(2 * time.Second)
		return nil
	})

	if result.Error != nil {
		if result.Error.Error() == "context deadline exceeded" {
			log.Debug("Expected timeout error received")
		} else {
			t.Fatal(result.Error)
		}
	}

	//now run it again and send the expected signal on the waiting channel,
	//then close the simulation
	log.Debug("Starting sim round and wait for frontend signal...")
	//this time the timeout should be long enough so that it doesn't kick in too early
	ctx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	go sendRunSignal(t)
	result = sim.Run(ctx, func(ctx context.Context, sim *Simulation) error {
		log.Debug("This run waits for the run signal from `frontend`...")
		//ensure with a Sleep that simulation doesn't terminate before the signal is received
		time.Sleep(2 * time.Second)
		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	log.Debug("Test terminated successfully")
}

func sendRunSignal(t *testing.T) {
	//We need to first wait for the sim HTTP server to start running...
	time.Sleep(2 * time.Second)
	//then we can send the signal

	log.Debug("Sending run signal to simulation: POST /runsim...")
	resp, err := http.Post(fmt.Sprintf("http://localhost:%s/runsim", DefaultHTTPSimPort), "application/json", nil)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	log.Debug("Signal sent")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("err %s", resp.Status)
	}
}

func (t *noopService) Stop() error {
	return nil
}
