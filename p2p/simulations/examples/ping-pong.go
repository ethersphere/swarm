// Copyright 2017 The go-ethereum Authors
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

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

var adapterType = flag.String("adapter", "sim", `node adapter to use (one of "sim", "exec" or "docker")`)

// main() starts a simulation network which contains nodes running a simple
// ping-pong protocol
func main() {
	flag.Parse()

	// set the log level to Trace
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	// register a single ping-pong service
	services := map[string]adapters.ServiceFunc{
		"ping-pong": func(ctx *adapters.ServiceContext) (node.Service, error) {
			return newPingPongService(ctx.Config.ID), nil
		},
	}
	adapters.RegisterServices(services)

	// create the NodeAdapter
	var adapter adapters.NodeAdapter

	switch *adapterType {

	case "sim":
		log.Info("using sim adapter")
		adapter = adapters.NewSimAdapter(services)

	case "exec":
		tmpdir, err := ioutil.TempDir("", "p2p-example")
		if err != nil {
			log.Crit("error creating temp dir", "err", err)
		}
		defer os.RemoveAll(tmpdir)
		log.Info("using exec adapter", "tmpdir", tmpdir)
		adapter = adapters.NewExecAdapter(tmpdir)

	case "docker":
		log.Info("using docker adapter")
		var err error
		adapter, err = adapters.NewDockerAdapter()
		if err != nil {
			log.Crit("error creating docker adapter", "err", err)
		}

	default:
		log.Crit(fmt.Sprintf("unknown node adapter %q", *adapterType))
	}

	// start the HTTP API
	log.Info("starting simulation server on 0.0.0.0:8888...")
	network := simulations.NewNetwork(adapter, &simulations.NetworkConfig{
		DefaultService: "ping-pong",
	})
	if err := http.ListenAndServe(":8888", simulations.NewServer(network)); err != nil {
		log.Crit("error starting simulation server", "err", err)
	}
}

// pingPongService runs a ping-pong protocol between nodes where each node
// sends a ping to all its connected peers every 10s and receives a pong in
// return
type pingPongService struct {
	id       discover.NodeID
	log      log.Logger
	closer   io.Closer
	received int64
}

func newPingPongService(id discover.NodeID) *pingPongService {
	return &pingPongService{
		id:  id,
		log: log.New("node.id", id),
	}
}

func (p *pingPongService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{{
		Name:     "ping-pong",
		Version:  1,
		Length:   2,
		Run:      p.Run,
		NodeInfo: p.Info,
	}}
}

func (p *pingPongService) APIs() []rpc.API {
	return nil
}

func (p *pingPongService) Start(server *p2p.Server) error {
	p.closer = initTracer()

	p.log.Info("ping-pong service starting")
	return nil
}

func (p *pingPongService) Stop() error {
	defer p.closer.Close()

	p.log.Info("ping-pong service stopping")
	return nil
}

func (p *pingPongService) Info() interface{} {
	return struct {
		Received int64 `json:"received"`
	}{
		atomic.LoadInt64(&p.received),
	}
}

const (
	pingMsgCode = iota
	pongMsgCode
)

type PingMsgPayload struct {
	MarshalledContext []byte
	Payload           string
	Version           uint64
}

type PongMsgPayload struct {
	MarshalledContext []byte
	Payload           string
	Version           uint64
}

// Run implements the ping-pong protocol which sends ping messages to the peer
// at 10s intervals, and responds to pings with pong messages.
func (p *pingPongService) Run(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
	log := p.log.New("peer.id", peer.ID())

	errC := make(chan error)
	go func() {
		for range time.Tick(5 * time.Second) {

			tracer := opentracing.GlobalTracer()

			sp := tracer.StartSpan("ping-operation")

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			err := tracer.Inject(
				sp.Context(),
				opentracing.Binary,
				writer)
			if err != nil {
				panic(err)
			}

			writer.Flush()

			pmp := &PingMsgPayload{
				b.Bytes(),
				"PING",
				7,
			}

			log.Info("sending ping with ctx", "ctx", fmt.Sprintf("%x", b.Bytes()))
			if err := p2p.Send(rw, pingMsgCode, pmp); err != nil {
				errC <- err
				return
			}

			sp.Finish()
		}
	}()
	go func() {
		for {
			msg, err := rw.ReadMsg()
			if err != nil {
				errC <- err
				return
			}

			log.Info("received message", "msg.code", msg.Code)
			atomic.AddInt64(&p.received, 1)
			if msg.Code == pingMsgCode {
				log.Info("received message ping")

				pmp := &PingMsgPayload{}

				if err := msg.Decode(pmp); err != nil {
					errC <- err
				}

				log.Info("received ping with ctx", "ctx", fmt.Sprintf("%x", pmp.MarshalledContext))

				tracer := opentracing.GlobalTracer()

				ctx, err := tracer.Extract(
					opentracing.Binary,
					bytes.NewReader(pmp.MarshalledContext))
				if err != nil {
					panic(err)
				}

				sp := tracer.StartSpan(
					"pong-operation",
					opentracing.ChildOf(ctx))

				var b bytes.Buffer
				writer := bufio.NewWriter(&b)

				err = tracer.Inject(
					sp.Context(),
					opentracing.Binary,
					writer)
				if err != nil {
					panic(err)
				}

				writer.Flush()

				pop := &PongMsgPayload{
					b.Bytes(),
					"PONG",
					8,
				}

				log.Info("sending pong")
				go p2p.Send(rw, pongMsgCode, pop)
				sp.Finish()
			}

			if msg.Code == pongMsgCode {
				log.Info("received message pong")
			}
		}
	}()
	return <-errC
}

func initTracer() (closer io.Closer) {
	fmt.Println("==== init Tracer ====")
	// Sample configuration for testing. Use constant sampling to sample every trace
	// and enable LogSpan to log every span via configured Logger.
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  "0.0.0.0:6831",
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	//jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := cfg.InitGlobalTracer(
		"ping-pong",
		jaegercfg.Logger(jLogger),
		//jaegercfg.Metrics(jMetricsFactory),
		//jaegercfg.Observer(rpcmetrics.NewObserver(jMetricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	if err != nil {
		panic(fmt.Sprintf("Could not initialize jaeger tracer: %s", err.Error()))
		return nil
	}

	return closer
}
