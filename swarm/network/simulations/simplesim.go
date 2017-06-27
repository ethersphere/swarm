package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm"
	"github.com/ethereum/go-ethereum/swarm/api"
)

var (
	verbose     = flag.Bool("v", false, "be verbose")
	veryVerbose = flag.Bool("vv", false, "be very verbose")
)

type Demo struct {
	Services adapters.Services
	Net      *simulations.Network
}

func main() {
	flag.Parse()

	if *veryVerbose {
		log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
	} else if *verbose {
		log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
	}

	if err := run(); err != nil {
		fmt.Println("err", err)
		log.Crit("error running demo", "err", err)
	}
}

func run() error {

	demoLog := log.New()
	demoLog.SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stdout, log.TerminalFormat(false))))

	demoLog.Info("starting the SIM...")

	demo, err := newDemo()
	if err != nil {
		return err
	}
	if err := start(demo.Net); err != nil {
		return err
	}

	srv := simulations.NewServer(&simulations.ServerConfig{
		NewAdapter: func() adapters.NodeAdapter {
			return adapters.NewSimAdapter(demo.Services)
		},

		ExternalNetworks: map[string]*simulations.Network{
			"demo": demo.Net,
		},
	})
	mux := http.NewServeMux()
	mux.Handle("/networks/", srv)
	mux.Handle("/networks", srv)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	httpSrv := http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", port),
		Handler: mux,
	}

	demoLog.Info(fmt.Sprintf("starting demo server on %s...", httpSrv.Addr))
	go httpSrv.ListenAndServe()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigc)
	<-sigc
	demoLog.Info("Shutting down...")
	return httpSrv.Close()
}

func newDemo() (*Demo, error) {
	demo := &Demo{
	//dataDir: dataDir,
	}
	demo.Services = adapters.Services{
		"swarmpss": newSwarmPssService,
	}
	adapter := adapters.NewSimAdapter(demo.Services)
	demo.Net = simulations.NewNetwork(adapter, &simulations.NetworkConfig{
		ID: "demo",
	})
	return demo, nil
}

func start(net *simulations.Network) error {

	// process shapshot
	jsonsnapshot, err := ioutil.ReadFile("snapshot10nodes.json")
	if err != nil {
		return err
	}
	snapshot := &simulations.Snapshot{}
	err = json.Unmarshal(jsonsnapshot, snapshot)
	if err != nil {
		return err
	}
	for _, node := range snapshot.Nodes {
		node.Config.Services = []string{"swarmpss"}
	}
	err = net.Load(snapshot)
	return err
}

func newSwarmPssService(ctx *adapters.ServiceContext) (node.Service, error) {
	dir := filepath.Join("", ctx.Config.ID.String())
	config, err := api.NewConfig(dir, common.Address{}, ctx.Config.PrivateKey, 5239)
	if err != nil {
		return nil, err
	}
	//basically it is needed to run the demo in Docker, but config.ListenAddr is only on master of go-ethereum,
	//not on network-testing-framework which is the current go-ethereum branch we are working with.
	//todo : change this when working against the go-ethereum master branch
	//	config.ListenAddr = "0.0.0.0"

	//config.EnsRoot = d.ensAddr

	swapEnabled := false
	syncEnabled := false
	pssEnabled := true

	allowedOrigins := "*"
	//we set that the Port to "0" to avoid the pss node listen to port 8500 (by default) to avoid collision with the "real" swarm node.
	config.Port = "0"

	return swarm.NewSwarm(ctx.NodeContext, nil, config, swapEnabled, syncEnabled, allowedOrigins, pssEnabled, false)
}
