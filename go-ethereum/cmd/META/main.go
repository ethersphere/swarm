package main

import (
	"github.com/ethereum/go-ethereum/crypto"
	"crypto/ecdsa"
	"fmt"
	"runtime"
	"os"
	
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/internal/debug"
	"github.com/ethereum/go-ethereum/node"
	//"github.com/ethereum/go-ethereum/ethclient"
	
	"github.com/ethereum/go-ethereum/META"
	METAapi "github.com/ethereum/go-ethereum/META/api"
	
	
	"gopkg.in/urfave/cli.v1"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"

)

const clientIdentifier = "METAd"

var (
	gitCommit string // Git SHA1 commit hash of the release (set via linker flags)
	app       = utils.NewApp(gitCommit, "META Wire daemon")
)

var (
	METAAccountFlag = cli.StringFlag{
		Name:  "metaaccount",
		Usage: "META account key file",
	}
	METAPortFlag = cli.StringFlag{
		Name:  "metaport",
		Usage: "META local http api port",
	}
	METANetworkIdFlag = cli.IntFlag{
		Name:  "metanetworkid",
		Usage: "Network identifier (integer, default 1666=meta testnet)",
		Value: network.NetworkId,
	}
	METAConfigPathFlag = cli.StringFlag{
		Name:  "metaconfig",
		Usage: "META config file path (datadir/META)",
	}
	EthAPI = cli.StringFlag{
		Name:  "ethapi",
		Usage: "URL of the Ethereum API provider",
		Value: node.DefaultIPCEndpoint("geth"),
	}
)

func init() {
	// Override flag defaults so bzzd can run alongside geth.
	utils.ListenPortFlag.Value = 31666
	utils.IPCPathFlag.Value = utils.DirectoryString{Value: "META.ipc"}
	utils.IPCApiFlag.Value = "admin, META, debug"

	// Set up the cli app.
	app.Commands = nil
	app.Action = METAd
	app.Flags = []cli.Flag{
		utils.IdentityFlag, // custom node name
		utils.DataDirFlag,
		//utils.BootnodesFlag,
		utils.KeyStoreDirFlag,
		utils.ListenPortFlag,
		//utils.NoDiscoverFlag,
		//utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.MaxPeersFlag,
		utils.NATFlag,
		utils.IPCDisabledFlag,
		utils.IPCApiFlag,
		utils.IPCPathFlag,
		// META-specific flags
		EthAPI,
		METAConfigPathFlag,
		METAPortFlag,
		METAAccountFlag,
		METANetworkIdFlag,
	}
	app.Flags = append(app.Flags, debug.Flags...)
	app.Before = func(ctx *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return debug.Setup(ctx)
	}
	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		return nil
	}
}


func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}


func METAd(ctx *cli.Context) error {
	stack := utils.MakeNode(ctx, clientIdentifier, gitCommit)
	registerMETAService(ctx, stack)
	utils.StartNode(stack)
	/*
	// Add bootnodes as initial peers.
	if ctx.GlobalIsSet(utils.BootnodesFlag.Name) {
		bootnodes := strings.Split(ctx.GlobalString(utils.BootnodesFlag.Name), ",")
		injectBootnodes(stack.Server(), bootnodes)
	} else {
		injectBootnodes(stack.Server(), defaultBootnodes)
	}
	*/
	stack.Wait()
	return nil
}

func registerMETAService(ctx *cli.Context, stack *node.Node) {
	prvkey := getAccount(ctx, stack)

	//chbookaddr := common.HexToAddress(ctx.GlobalString(ChequebookAddrFlag.Name))
	metadir := ctx.GlobalString(METAConfigPathFlag.Name)
	if metadir == "" {
		metadir = stack.InstanceDir()
	}
	metaconfig, err := METAapi.NewConfig(metadir, prvkey, ctx.GlobalUint64(METANetworkIdFlag.Name))
	if err != nil {
		utils.Fatalf("unable to configure META: %v", err)
	}
	metaport := ctx.GlobalString(METAPortFlag.Name)
	if len(metaport) > 0 {
		metaconfig.Port = metaport
	}

	//ethapi := ctx.GlobalString(EthAPI.Name)

	boot := func(ctx *node.ServiceContext) (node.Service, error) {
		// we will probably want this
		/*
		var client *ethclient.Client
		if ethapi == "" {
			err = fmt.Errorf("use ethapi flag to connect to a an eth client and talk to the blockchain")
		} else {
			client, err = ethclient.Dial(ethapi)
		}
		if err != nil {
			utils.Fatalf("Can't connect: %v", err)
		}
		return META.NewMETA(ctx, client, metaconfig)*/
		return META.NewMETA(ctx, metaconfig)
		
	}
	if err := stack.Register(boot); err != nil {
		utils.Fatalf("Failed to register the META service: %v", err)
	}
	
	glog.V(logger.Info).Infof("Boot %v", boot)
	return
}

func getAccount(ctx *cli.Context, stack *node.Node) *ecdsa.PrivateKey {
	keyid := ctx.GlobalString(METAAccountFlag.Name)
	if keyid == "" {
		utils.Fatalf("Option %q is required", METAAccountFlag.Name)
	}
	// Try to load the arg as a hex key file.
	if key, err := crypto.LoadECDSA(keyid); err == nil {
		glog.V(logger.Info).Infof("swarm account key loaded: %#x", crypto.PubkeyToAddress(key.PublicKey))
		return key
	}
	// Otherwise try getting it from the keystore.
	return decryptStoreAccount(stack.AccountManager(), keyid)
	
	return nil
}

/***
 * \todo implement decrypt, now only makes new key
 */
func decryptStoreAccount(accman *accounts.Manager, account string) *ecdsa.PrivateKey {
	pk, _ := crypto.GenerateKey()
	return pk
}
