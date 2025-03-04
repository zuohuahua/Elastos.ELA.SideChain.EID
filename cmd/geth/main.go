// Copyright 2014 The Elastos.ELA.SideChain.EID Authors
// This file is part of Elastos.ELA.SideChain.EID.
//
// Elastos.ELA.SideChain.EID is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Elastos.ELA.SideChain.EID is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Elastos.ELA.SideChain.EID. If not, see <http://www.gnu.org/licenses/>.

// geth is the official command-line client for Ethereum.
package main

import (
	"errors"
	"fmt"
	"github.com/elastos/Elastos.ELA.SideChain.EID/core/vm"
	"github.com/elastos/Elastos.ELA.SideChain.EID/core/vm/did"
	"math"
	"os"
	"runtime"
	godebug "runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elastos/Elastos.ELA.SideChain.EID/accounts"
	"github.com/elastos/Elastos.ELA.SideChain.EID/accounts/keystore"
	chainbridge_core "github.com/elastos/Elastos.ELA.SideChain.EID/chainbridge-core"
	"github.com/elastos/Elastos.ELA.SideChain.EID/cmd/utils"
	"github.com/elastos/Elastos.ELA.SideChain.EID/common"
	"github.com/elastos/Elastos.ELA.SideChain.EID/consensus/pbft"
	"github.com/elastos/Elastos.ELA.SideChain.EID/console"
	"github.com/elastos/Elastos.ELA.SideChain.EID/core"
	"github.com/elastos/Elastos.ELA.SideChain.EID/core/events"
	"github.com/elastos/Elastos.ELA.SideChain.EID/eth"
	"github.com/elastos/Elastos.ELA.SideChain.EID/eth/downloader"
	"github.com/elastos/Elastos.ELA.SideChain.EID/ethclient"
	"github.com/elastos/Elastos.ELA.SideChain.EID/internal/debug"
	"github.com/elastos/Elastos.ELA.SideChain.EID/les"
	"github.com/elastos/Elastos.ELA.SideChain.EID/log"
	"github.com/elastos/Elastos.ELA.SideChain.EID/metrics"
	"github.com/elastos/Elastos.ELA.SideChain.EID/node"
	"github.com/elastos/Elastos.ELA.SideChain.EID/smallcrosstx"
	"github.com/elastos/Elastos.ELA.SideChain.EID/spv"
	"github.com/elastos/Elastos.ELA.SideChain.EID/withdrawfailedtx"

	elacom "github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/core/contract"

	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"github.com/elastic/gosigar"
	"golang.org/x/crypto/ripemd160"
	"gopkg.in/urfave/cli.v1"
	"path/filepath"
)

const (
	clientIdentifier = "geth" // Client identifier to advertise over the network
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, gitDate, "the Elastos.ELA.SideChain.EID command line interface")
	// flags that configure the node
	nodeFlags = []cli.Flag{
		utils.IdentityFlag,
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.BootnodesFlag,
		utils.BootnodesV4Flag,
		utils.BootnodesV5Flag,
		utils.DataDirFlag,
		utils.AncientFlag,
		utils.KeyStoreDirFlag,
		utils.ExternalSignerFlag,
		utils.NoUSBFlag,
		utils.SmartCardDaemonPathFlag,
		utils.OverrideIstanbulFlag,
		utils.DashboardEnabledFlag,
		utils.DashboardAddrFlag,
		utils.DashboardPortFlag,
		utils.DashboardRefreshFlag,
		utils.EthashCacheDirFlag,
		utils.EthashCachesInMemoryFlag,
		utils.EthashCachesOnDiskFlag,
		utils.EthashDatasetDirFlag,
		utils.EthashDatasetsInMemoryFlag,
		utils.EthashDatasetsOnDiskFlag,
		utils.TxPoolLocalsFlag,
		utils.TxPoolNoLocalsFlag,
		utils.TxPoolJournalFlag,
		utils.TxPoolRejournalFlag,
		utils.TxPoolPriceLimitFlag,
		utils.TxPoolPriceBumpFlag,
		utils.TxPoolAccountSlotsFlag,
		utils.TxPoolGlobalSlotsFlag,
		utils.TxPoolAccountQueueFlag,
		utils.TxPoolGlobalQueueFlag,
		utils.TxPoolLifetimeFlag,
		utils.SyncModeFlag,
		utils.ExitWhenSyncedFlag,
		utils.GCModeFlag,
		utils.LightServeFlag,
		utils.LightLegacyServFlag,
		utils.LightIngressFlag,
		utils.LightEgressFlag,
		utils.LightMaxPeersFlag,
		utils.LightLegacyPeersFlag,
		utils.LightKDFFlag,
		utils.UltraLightServersFlag,
		utils.UltraLightFractionFlag,
		utils.UltraLightOnlyAnnounceFlag,
		utils.WhitelistFlag,
		utils.CacheFlag,
		utils.CacheDatabaseFlag,
		utils.CacheTrieFlag,
		utils.CacheGCFlag,
		utils.CacheNoPrefetchFlag,
		utils.ListenPortFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.MiningEnabledFlag,
		utils.MinerThreadsFlag,
		utils.MinerLegacyThreadsFlag,
		utils.MinerNotifyFlag,
		utils.MinerGasTargetFlag,
		utils.MinerLegacyGasTargetFlag,
		utils.MinerGasLimitFlag,
		utils.MinerGasPriceFlag,
		utils.MinerLegacyGasPriceFlag,
		utils.MinerEtherbaseFlag,
		utils.MinerLegacyEtherbaseFlag,
		utils.MinerExtraDataFlag,
		utils.MinerLegacyExtraDataFlag,
		utils.MinerRecommitIntervalFlag,
		utils.MinerNoVerfiyFlag,
		utils.NATFlag,
		utils.NoDiscoverFlag,
		utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.DeveloperFlag,
		utils.DeveloperPeriodFlag,
		utils.TestnetFlag,
		utils.RinkebyFlag,
		utils.GoerliFlag,
		utils.VMEnableDebugFlag,
		utils.NetworkIdFlag,
		utils.EthStatsURLFlag,
		utils.FakePoWFlag,
		utils.NoCompactionFlag,
		utils.GpoBlocksFlag,
		utils.GpoPercentileFlag,
		utils.EWASMInterpreterFlag,
		utils.EVMInterpreterFlag,
		configFileFlag,
		utils.SpvMonitoringAddrFlag,
		utils.PassBalance,
		utils.BlackContractAddr,
		utils.PreConnectOffset,
		utils.PbftKeyStore,
		utils.PbftKeystorePassWord,
		utils.PbftIPAddress,
		utils.PbftDposPort,
		utils.OldDIDMigrateAddrFlag,
		utils.OldDIDMigrateHeightFlag,
		utils.DocArraySortHeightFlag,
		utils.CheckCustomizeDIDBeginHeightFlag,
		utils.PbftMinerAddress,
		utils.DynamicArbiter,
		utils.FrozenAccount,
		utils.UpdateArbiterListToLayer1Flag,
	}

	rpcFlags = []cli.Flag{
		utils.RPCEnabledFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		utils.RPCCORSDomainFlag,
		utils.RPCVirtualHostsFlag,
		utils.GraphQLEnabledFlag,
		utils.GraphQLListenAddrFlag,
		utils.GraphQLPortFlag,
		utils.GraphQLCORSDomainFlag,
		utils.GraphQLVirtualHostsFlag,
		utils.RPCApiFlag,
		utils.WSEnabledFlag,
		utils.WSListenAddrFlag,
		utils.WSPortFlag,
		utils.WSApiFlag,
		utils.WSAllowedOriginsFlag,
		utils.IPCDisabledFlag,
		utils.IPCPathFlag,
		utils.InsecureUnlockAllowedFlag,
		utils.RPCGlobalGasCap,
	}

	whisperFlags = []cli.Flag{
		utils.WhisperEnabledFlag,
		utils.WhisperMaxMessageSizeFlag,
		utils.WhisperMinPOWFlag,
		utils.WhisperRestrictConnectionBetweenLightClientsFlag,
	}

	metricsFlags = []cli.Flag{
		utils.MetricsEnabledFlag,
		utils.MetricsEnabledExpensiveFlag,
		utils.MetricsEnableInfluxDBFlag,
		utils.MetricsInfluxDBEndpointFlag,
		utils.MetricsInfluxDBDatabaseFlag,
		utils.MetricsInfluxDBUsernameFlag,
		utils.MetricsInfluxDBPasswordFlag,
		utils.MetricsInfluxDBTagsFlag,
	}
)

func init() {
	// Initialize the CLI app and start Geth
	app.Action = geth
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2013-2019 The Elastos.ELA.SideChain.EID Authors"
	app.Commands = []cli.Command{
		// See chaincmd.go:
		initCommand,
		importCommand,
		exportCommand,
		importPreimagesCommand,
		exportPreimagesCommand,
		copydbCommand,
		removedbCommand,
		dumpCommand,
		inspectCommand,
		// See accountcmd.go:
		accountCommand,
		walletCommand,
		// See consolecmd.go:
		consoleCommand,
		attachCommand,
		javascriptCommand,
		// See misccmd.go:
		makecacheCommand,
		makedagCommand,
		versionCommand,
		licenseCommand,
		// See config.go
		dumpConfigCommand,
		// See retesteth.go
		retestethCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, rpcFlags...)
	app.Flags = append(app.Flags, consoleFlags...)
	app.Flags = append(app.Flags, debug.Flags...)
	app.Flags = append(app.Flags, whisperFlags...)
	app.Flags = append(app.Flags, metricsFlags...)

	app.Before = func(ctx *cli.Context) error {
		logdir := ""
		if ctx.GlobalBool(utils.DashboardEnabledFlag.Name) {
			logdir = (&node.Config{DataDir: utils.MakeDataDir(ctx)}).ResolvePath("logs")
		}
		if err := debug.Setup(ctx, logdir); err != nil {
			return err
		}
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		console.Stdin.Close() // Resets terminal mode.
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// prepare manipulates memory cache allowance and setups metric system.
// This function should be called before launching devp2p stack.
func prepare(ctx *cli.Context) {
	// If we're a full node on mainnet without --cache specified, bump default cache allowance
	if ctx.GlobalString(utils.SyncModeFlag.Name) != "light" && !ctx.GlobalIsSet(utils.CacheFlag.Name) && !ctx.GlobalIsSet(utils.NetworkIdFlag.Name) {
		// Make sure we're not on any supported preconfigured testnet either
		if !ctx.GlobalIsSet(utils.TestnetFlag.Name) && !ctx.GlobalIsSet(utils.RinkebyFlag.Name) && !ctx.GlobalIsSet(utils.GoerliFlag.Name) && !ctx.GlobalIsSet(utils.DeveloperFlag.Name) {
			// Nope, we're really on mainnet. Bump that cache up!
			log.Info("Bumping default cache on mainnet", "provided", ctx.GlobalInt(utils.CacheFlag.Name), "updated", 4096)
			ctx.GlobalSet(utils.CacheFlag.Name, strconv.Itoa(4096))
		}
	}
	// If we're running a light client on any network, drop the cache to some meaningfully low amount
	if ctx.GlobalString(utils.SyncModeFlag.Name) == "light" && !ctx.GlobalIsSet(utils.CacheFlag.Name) {
		log.Info("Dropping default light client cache", "provided", ctx.GlobalInt(utils.CacheFlag.Name), "updated", 128)
		ctx.GlobalSet(utils.CacheFlag.Name, strconv.Itoa(128))
	}
	// Cap the cache allowance and tune the garbage collector
	var mem gosigar.Mem
	// Workaround until OpenBSD support lands into gosigar
	// Check https://github.com/elastic/gosigar#supported-platforms
	if runtime.GOOS != "openbsd" {
		if err := mem.Get(); err == nil {
			allowance := int(mem.Total / 1024 / 1024 / 3)
			if cache := ctx.GlobalInt(utils.CacheFlag.Name); cache > allowance {
				log.Warn("Sanitizing cache to Go's GC limits", "provided", cache, "updated", allowance)
				ctx.GlobalSet(utils.CacheFlag.Name, strconv.Itoa(allowance))
			}
		}
	}
	// Ensure Go's GC ignores the database cache for trigger percentage
	cache := ctx.GlobalInt(utils.CacheFlag.Name)
	gogc := math.Max(20, math.Min(100, 100/(float64(cache)/1024)))

	log.Info("Sanitizing Go's GC trigger", "percent", int(gogc))
	godebug.SetGCPercent(int(gogc))

	// Start metrics export if enabled
	utils.SetupMetrics(ctx)

	// Start system runtime metrics collection
	go metrics.CollectProcessMetrics(3 * time.Second)
}

// geth is the main entry point into the system if no special subcommand is ran.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func geth(ctx *cli.Context) error {
	if args := ctx.Args(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}
	prepare(ctx)
	node := makeFullNode(ctx)
	defer node.Close()
	startNode(ctx, node)
	node.Wait()
	return nil
}

// calculate the ELA mainchain address from the sidechain (ie. this chain)
// genesis block hash for corresponding crosschain transactions
// refer to https://github.com/elastos/Elastos.ELA.Client/blob/dev/cli/wallet/wallet.go
// for the original ELA-CLI implementation
func calculateGenesisAddress(genesisBlockHash string) (string, error) {
	// unlike Ethereum, the ELA hash values do not contain 0x prefix
	if strings.HasPrefix(genesisBlockHash, "0x") {
		genesisBlockHash = genesisBlockHash[2:]
	}
	genesisBlockBytes, err := hex.DecodeString(genesisBlockHash)
	if err != nil {
		return "", errors.New("genesis block hash to bytes failed")
	}
	reversedGenesisBlockBytes := elacom.BytesReverse(genesisBlockBytes)
	reversedGenesisBlockStr := elacom.BytesToHexString(reversedGenesisBlockBytes)

	log.Info(fmt.Sprintf("genesis program hash: %v", reversedGenesisBlockStr))

	buf := new(bytes.Buffer)
	buf.WriteByte(byte(len(reversedGenesisBlockBytes)))
	buf.Write(reversedGenesisBlockBytes)
	buf.WriteByte(byte(elacom.CROSSCHAIN))

	sum168 := func(prefix byte, code []byte) []byte {
		hash := sha256.Sum256(code)
		md160 := ripemd160.New()
		md160.Write(hash[:])
		return md160.Sum([]byte{prefix})
	}
	genesisProgramHash, err := elacom.Uint168FromBytes(sum168(byte(contract.PrefixCrossChain), buf.Bytes()))
	if err != nil {
		return "", errors.New("genesis block bytes to program hash failed")
	}

	genesisAddress, err := genesisProgramHash.ToAddress()
	if err != nil {
		return "", errors.New("genesis block hash to genesis address failed")
	}
	log.Info(fmt.Sprintf("genesis address: %v ", genesisAddress))

	return genesisAddress, nil
}

func startSpv(ctx *cli.Context, stack *node.Node) {

	var SpvDataDir string
	switch {
	case ctx.GlobalIsSet(utils.DataDirFlag.Name):
		SpvDataDir = ctx.GlobalString(utils.DataDirFlag.Name)
	case ctx.GlobalBool(utils.DeveloperFlag.Name):
		SpvDataDir = "" // unless explicitly requested, use memory databases
	case ctx.GlobalBool(utils.TestnetFlag.Name):
		SpvDataDir = filepath.Join(node.DefaultDataDir(), "testnet")
	case ctx.GlobalBool(utils.RinkebyFlag.Name):
		SpvDataDir = filepath.Join(node.DefaultDataDir(), "rinkeby")
	case ctx.GlobalBool(utils.GoerliFlag.Name):
		SpvDataDir = filepath.Join(node.DefaultDataDir(), "goerli")
	default:
		SpvDataDir = node.DefaultDataDir()
	}

	var spvCfg = &spv.Config{
		DataDir: SpvDataDir,
	}
	// prepare the SPV service config parameters
	switch {
	case ctx.GlobalBool(utils.TestnetFlag.Name):
		spvCfg.ActiveNet = "t"

	case ctx.GlobalBool(utils.RinkebyFlag.Name):
		spvCfg.ActiveNet = "r"
	case ctx.GlobalBool(utils.GoerliFlag.Name):
		spvCfg.ActiveNet = "g"
	}

	// prepare to start the SPV module
	// if --spvmoniaddr commandline parameter is present, use the parameter value
	// as the ELA mainchain address for the SPV module to monitor on
	// if no --spvmoniaddr commandline parameter is provided, use the sidechain genesis block hash
	// to generate the corresponding ELA mainchain address for the SPV module to monitor on
	var dynamicArbiterHeight uint64
	if ctx.GlobalString(utils.SpvMonitoringAddrFlag.Name) != "" {
		// --spvmoniaddr parameter is provided, set the SPV monitor address accordingly
		log.Info("SPV Start Monitoring... ", "SpvMonitoringAddr", ctx.GlobalString(utils.SpvMonitoringAddrFlag.Name))
		spvCfg.GenesisAddress = ctx.GlobalString(utils.SpvMonitoringAddrFlag.Name)
	} else {
		// --spvmoniaddr parameter is not provided
		// get the Ethereum node service to get the genesis block hash
		var fullnode *eth.Ethereum
		var lightnode *les.LightEthereum
		var ghash common.Hash

		// light node and full node are different types of node services
		if ctx.GlobalString(utils.SyncModeFlag.Name) == "light" {
			if err := stack.Service(&lightnode); err != nil {
				utils.Fatalf("Blockchain not running: %v", err)
			}
			ghash = lightnode.BlockChain().Genesis().Hash()
			dynamicArbiterHeight = lightnode.BlockChain().Config().DynamicArbiterHeight
		} else {
			if err := stack.Service(&fullnode); err != nil {
				utils.Fatalf("Blockchain not running: %v", err)
			}
			ghash = fullnode.BlockChain().Genesis().Hash()
			dynamicArbiterHeight = fullnode.BlockChain().Config().DynamicArbiterHeight
		}

		// calculate ELA mainchain address from the genesis block hash and set the SPV monitor address accordingly
		log.Info(fmt.Sprintf("Genesis block hash: %v", ghash.String()))
		if gaddr, err := calculateGenesisAddress(ghash.String()); err != nil {
			utils.Fatalf("Cannot calculate: %v", err)
		} else {
			log.Info(fmt.Sprintf("SPV Start Monitoring... : %v", gaddr))
			spvCfg.GenesisAddress = gaddr
		}
	}

	spv.GetDefaultSingerAddr = func() common.Address {
		var addr common.Address
		if wallets := stack.AccountManager().Wallets(); len(wallets) > 0 {
			if accounts := wallets[0].Accounts(); len(accounts) > 0 {
				addr = accounts[0].Address
			}
		}

		return addr
	}

	client, err := stack.Attach()
	if err != nil {
		log.Error("Attach client: ", "err", err)
	}
	if spvService, err := spv.NewService(spvCfg, client, stack.EventMux(), dynamicArbiterHeight); err != nil {
		utils.Fatalf("SPV service init error: %v", err)
	} else {
		MinedBlockSub := stack.EventMux().Subscribe(events.MinedBlockEvent{})
		OnDutySub := stack.EventMux().Subscribe(events.OnDutyEvent{})
		smallCroTxSub := stack.EventMux().Subscribe(events.CmallCrossTx{})
		go spv.MinedBroadcastLoop(MinedBlockSub, OnDutySub, smallCroTxSub)
		spvService.Start()
		stack.EventMux().Post(events.InitCurrentProducers{})
		spv.InitNextTurnDposInfo()
	}
}

// startNode boots up the system node and all registered protocols, after which
// it unlocks any requested accounts, and starts the RPC/IPC interfaces and the
// miner.
func startNode(ctx *cli.Context, stack *node.Node) {
	debug.Memsize.Add("node", stack)

	// Start up the node itself
	utils.StartNode(stack)

	// Unlock any account specifically requested
	unlockAccounts(ctx, stack)
	// Start auxiliary services if enabled
	var ethereum *eth.Ethereum
	if ctx.GlobalString(utils.SyncModeFlag.Name) != "light" {
		if err := stack.Service(&ethereum); err != nil {
			utils.Fatalf("Ethereum service not running: %v", err)
		}
		initChainBridge(ctx, stack, ethereum.BlockChain())
	}
	//start the SPV service
	//log.Info(fmt.Sprintf("Starting SPV service with config: %+v \n", *spvCfg))
	startSpv(ctx, stack)
	SetDIDParams(ctx, stack)
	startSmallCrossTx(ctx, stack)
	// Register wallet event handlers to open and auto-derive wallets
	evts := make(chan accounts.WalletEvent, 16)
	stack.AccountManager().Subscribe(evts)

	// Create a client to interact with local geth node.
	rpcClient, err := stack.Attach()
	if err != nil {
		utils.Fatalf("Failed to attach to self: %v", err)
	}
	ethClient := ethclient.NewClient(rpcClient)

	// Set contract backend for ethereum service if local node
	// is serving LES requests.
	if ctx.GlobalInt(utils.LightLegacyServFlag.Name) > 0 || ctx.GlobalInt(utils.LightServeFlag.Name) > 0 {
		var ethService *eth.Ethereum
		if err := stack.Service(&ethService); err != nil {
			utils.Fatalf("Failed to retrieve ethereum service: %v", err)
		}
		ethService.SetContractBackend(ethClient)
	}
	// Set contract backend for les service if local node is
	// running as a light client.
	if ctx.GlobalString(utils.SyncModeFlag.Name) == "light" {
		var lesService *les.LightEthereum
		if err := stack.Service(&lesService); err != nil {
			utils.Fatalf("Failed to retrieve light ethereum service: %v", err)
		}
		lesService.SetContractBackend(ethClient)
	}

	go func() {
		// Open any wallets already attached
		for _, wallet := range stack.AccountManager().Wallets() {
			if err := wallet.Open(""); err != nil {
				log.Warn("Failed to open wallet", "url", wallet.URL(), "err", err)
			}
		}
		// Listen for wallet event till termination
		for event := range evts {
			switch event.Kind {
			case accounts.WalletArrived:
				if err := event.Wallet.Open(""); err != nil {
					log.Warn("New wallet appeared, failed to open", "url", event.Wallet.URL(), "err", err)
				}
			case accounts.WalletOpened:
				status, _ := event.Wallet.Status()
				log.Info("New wallet appeared", "url", event.Wallet.URL(), "status", status)

				var derivationPaths []accounts.DerivationPath
				if event.Wallet.URL().Scheme == "ledger" {
					derivationPaths = append(derivationPaths, accounts.LegacyLedgerBaseDerivationPath)
				}
				derivationPaths = append(derivationPaths, accounts.DefaultBaseDerivationPath)

				event.Wallet.SelfDerive(derivationPaths, ethClient)

			case accounts.WalletDropped:
				log.Info("Old wallet dropped", "url", event.Wallet.URL())
				event.Wallet.Close()
			}
		}
	}()

	// Spawn a standalone goroutine for status synchronization monitoring,
	// close the node when synchronization is complete if user required.
	if ctx.GlobalBool(utils.ExitWhenSyncedFlag.Name) {
		go func() {
			sub := stack.EventMux().Subscribe(downloader.DoneEvent{})
			defer sub.Unsubscribe()
			for {
				event := <-sub.Chan()
				if event == nil {
					continue
				}
				done, ok := event.Data.(downloader.DoneEvent)
				if !ok {
					continue
				}
				if timestamp := time.Unix(int64(done.Latest.Time), 0); time.Since(timestamp) < 10*time.Minute {
					log.Info("Synchronisation completed", "latestnum", done.Latest.Number, "latesthash", done.Latest.Hash(),
						"age", common.PrettyAge(timestamp))
					stack.Stop()
				}
			}
		}()
	}

	if ctx.GlobalBool(utils.MiningEnabledFlag.Name) || ctx.GlobalBool(utils.DeveloperFlag.Name) {
		// Mining only makes sense if a full Ethereum node is running
		if ctx.GlobalString(utils.SyncModeFlag.Name) == "light" {
			utils.Fatalf("Light clients do not support mining")
		}

		// Set the gas price to the limits from the CLI and start mining
		gasprice := utils.GlobalBig(ctx, utils.MinerLegacyGasPriceFlag.Name)
		if ctx.IsSet(utils.MinerGasPriceFlag.Name) {
			gasprice = utils.GlobalBig(ctx, utils.MinerGasPriceFlag.Name)
		}
		ethereum.TxPool().SetGasPrice(gasprice)

		threads := ctx.GlobalInt(utils.MinerLegacyThreadsFlag.Name)
		if ctx.GlobalIsSet(utils.MinerThreadsFlag.Name) {
			threads = ctx.GlobalInt(utils.MinerThreadsFlag.Name)
		}
		if ethereum.Engine() != ethereum.BlockChain().GetDposEngine() {
			if err := ethereum.StartMining(threads); err != nil {
				utils.Fatalf("Failed to start mining: %v", err)
			}
		}
		//go startLayer2(ethereum.BlockChain())
	}

	//xxl add update Arbiter List To Layer1 get param
	isUpdateAbiterToLayer1 := ctx.GlobalBool(utils.UpdateArbiterListToLayer1Flag.Name)
	log.Info("xxl isUpdateAbiterToLayer1 flag is ", "isUpdateAbiterToLayer1", isUpdateAbiterToLayer1)
	if isUpdateAbiterToLayer1 {
		log.Info("xxl StartUpdateNode ")
		chainbridge_core.StartUpdateNode()
	}
}

func initChainBridge(ctx *cli.Context, stack *node.Node, blockChain *core.BlockChain) {
	accPath := ""
	if wallets := stack.AccountManager().Wallets(); len(wallets) > 0 {
		accPath = wallets[0].URL().Path
	}
	if accPath == "" {
		log.Info("is common sync node")
	}
	password := ""
	passwords := utils.MakePasswordList(ctx)
	if len(passwords) > 0 {
		password = passwords[0]
	} else {
		log.Info("is common sync node, no password")
	}
	engine := blockChain.GetDposEngine().(*pbft.Pbft)
	chainbridge_core.Init(engine, stack, accPath, password)
}

func startLayer2(blockChain *core.BlockChain) {
	engine := blockChain.GetDposEngine().(*pbft.Pbft)
	if engine.GetProducer() != nil {
		if chainbridge_core.Start() {
			engine.AnnounceDAddr()
		}
	}
}

// unlockAccounts unlocks any account specifically requested.
func unlockAccounts(ctx *cli.Context, stack *node.Node) {
	var unlocks []string
	inputs := strings.Split(ctx.GlobalString(utils.UnlockedAccountFlag.Name), ",")
	for _, input := range inputs {
		if trimmed := strings.TrimSpace(input); trimmed != "" {
			unlocks = append(unlocks, trimmed)
		}
	}
	// Short circuit if there is no account to unlock.
	if len(unlocks) == 0 {
		return
	}
	// If insecure account unlocking is not allowed if node's APIs are exposed to external.
	// Print warning log to user and skip unlocking.
	if !stack.Config().InsecureUnlockAllowed && stack.Config().ExtRPCEnabled() {
		utils.Fatalf("Account unlock with HTTP access is forbidden!")
	}
	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	passwords := utils.MakePasswordList(ctx)
	for i, account := range unlocks {
		unlockAccount(ks, account, i, passwords)
	}
}

func SetDIDParams(ctx *cli.Context, stack *node.Node) {
	switch {
	case ctx.GlobalBool(utils.TestnetFlag.Name):
		vm.InitDIDParams(did.TestNetDIDParams)
	case ctx.GlobalBool(utils.RinkebyFlag.Name):
		vm.InitDIDParams(did.RegNetDIDParams)
	case ctx.GlobalBool(utils.GoerliFlag.Name):
		vm.InitDIDParams(did.RegNetDIDParams)
	default:
		vm.InitDIDParams(did.MainNetDIDParams)
	}
}

func startSmallCrossTx(ctx *cli.Context, stack *node.Node) {
	var datadir string
	switch {
	case ctx.GlobalIsSet(utils.DataDirFlag.Name):
		datadir = ctx.GlobalString(utils.DataDirFlag.Name)
	case ctx.GlobalBool(utils.DeveloperFlag.Name):
		datadir = "" // unless explicitly requested, use memory databases
	case ctx.GlobalBool(utils.TestnetFlag.Name):
		datadir = filepath.Join(node.DefaultDataDir(), "testnet")
	case ctx.GlobalBool(utils.RinkebyFlag.Name):
		datadir = filepath.Join(node.DefaultDataDir(), "rinkeby")
	case ctx.GlobalBool(utils.GoerliFlag.Name):
		datadir = filepath.Join(node.DefaultDataDir(), "goerli")
	default:
		datadir = node.DefaultDataDir()
	}
	smallcrosstx.SmallCrossTxInit(datadir, stack.EventMux())

	withdrawfailedtx.FailedWithrawInit(datadir, stack.EventMux())
}
