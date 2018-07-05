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
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/simulations"
)

func (s *Simulation) initHTTPServer(opts *SimulationOptions) {
	//assign default port if nothing provided
	if opts.HTTPSimPort == "" {
		opts.HTTPSimPort = DefaultHTTPSimPort
	}
	log.Info(fmt.Sprintf("Initializing simulation server on 0.0.0.0:%s...", opts.HTTPSimPort))
	//initialize the HTTP server
	s.handler = simulations.NewServer(s.Net)
	s.runC = make(chan struct{})
	//add swarm specific routes to the HTTP server
	s.addSimulationRoutes()
	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%s", opts.HTTPSimPort),
		Handler: s.handler,
	}
}

func (s *Simulation) startHTTPServer(ctx context.Context) error {
	//start the HTTP server
	if s.httpSrv != nil {
		go s.httpSrv.ListenAndServe()
		log.Info("Waiting for frontend to be ready...(send POST /runsim to HTTP server)")
		//wait for the frontend to connect
		select {
		case <-s.runC:
		case <-ctx.Done():
			return ctx.Err()
		}
		log.Info("Received signal from frontend - starting simulation run.")
	}
	return nil
}

//register additional HTTP routes
func (s *Simulation) addSimulationRoutes() {
	s.handler.POST("/runsim", s.RunSimulation)
}

// StartNetwork starts all nodes in the network
func (s *Simulation) RunSimulation(w http.ResponseWriter, req *http.Request) {
	log.Debug("RunSimulation endpoint running")
	s.runC <- struct{}{}
	w.WriteHeader(http.StatusOK)
}
