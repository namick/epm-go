package main

import (
	"flag"
	"github.com/eris-ltd/decerver-interfaces/glue/utils"
	"github.com/eris-ltd/epm-go"
	"github.com/eris-ltd/thelonious/monk"
	"github.com/eris-ltd/thelonious/monklog"
	"log"
	"os"
	"path"
	"path/filepath"
)

// TODO: use a CLI library!

var (
	GoPath = os.Getenv("GOPATH")

	logger *monklog.Logger = monklog.NewLogger("EPM")

	// adjust these to suit all your deformed nefarious extension name desires. Muahahaha
	// but actually don't because you might break something ;)
	PkgExt  = "pdx"
	TestExt = "pdt"

	defaultContractPath = "."
	defaultPackagePath  = "."
	defaultGenesis      = ""
	defaultKeys         = ""
	defaultDatabase     = ".chain"
	defaultLogLevel     = 5
	defaultDifficulty   = 14
	defaultMining       = false
	defaultDiffStorage  = false

	ini       = flag.Bool("init", false, "initialize a monkchain config")
	deploy    = flag.Bool("genesis-deploy", false, "deploy a monkchain")
	config    = flag.String("config", "monk-config.json", "pick config file")
	genesis   = flag.String("genesis", "genesis.json", "Set a genesis.json or genesis.pdx file")
	name      = flag.String("name", "", "Set the chain by name")
	chainId   = flag.String("id", "", "Set the chain by id")
	chainType = flag.String("type", "thel", "Set the chain type (thelonious, genesis, bitcoin, ethereum)")

	checkout = flag.String("checkout", "", "Checkout a chain")
	addRef   = flag.String("add-ref", "", "Add a new reference to a chainId")

	difficulty = flag.Int("dif", 14, "Set the mining difficulty")
	mining     = flag.Bool("mine", false, "To mine or not to mine, that is the question")
	noGenDoug  = flag.Bool("no-gendoug", false, "Turn off gendoug mechanics")

	interactive = flag.Bool("i", false, "Run epm in interactive mode")
	diffStorage = flag.Bool("diff", false, "Show a diff of all contract storage")
	clean       = flag.Bool("clean", false, "Clear out epm related dirs")
	update      = flag.Bool("update", false, "Pull and install the latest epm")
	install     = flag.Bool("install", false, "Re-install epm")

	contractPath = flag.String("c", defaultContractPath, "Set the contract root path")
	packagePath  = flag.String("p", ".", "Set a .package-definition file")
	keys         = flag.String("k", "", "Set a keys file")
	database     = flag.String("db", ".chain", "Set the location of the root directory")
	logLevel     = flag.Int("log", 5, "Set the eth log level")

	rpc     = flag.Bool("rpc", false, "Fire commands over rpc")
	rpcHost = flag.String("host", "localhost", "Set the rpc host")
	rpcPort = flag.String("port", "40404", "Set the rpc port")
)

func main() {
	flag.Parse()

	utils.InitLogging(path.Join(utils.Logs, "epm"), "", *logLevel, "")

	// clean, update, or install
	// exit
	if *clean || *update || *install {
		cleanUpdateInstall()
		os.Exit(0)
	}

	var err error

	// create ~/.decerver tree and drop monk/gen configs
	// exit
	if *ini {
		exit(inity())
	}

	// deploy the genblock, install into ~/.decerver
	// exit
	if *deploy {
		exit(monk.DeploySequence(*name, *genesis, *config))
	}

	// change the currently active chain
	// exit
	if *checkout != "" {
		exit(utils.ChangeHead(*checkout))
	}

	// add a new reference to a chainId
	// exit
	if *addRef != "" {
		if *name == "" {
			log.Fatal(`add-ref requires a name to specified as well, \n
                            eg. "add-ref 14c32 -name shitchain"`)
		}
		exit(utils.AddRef(*addRef, *name))
	}

	/*
	   Now we're actually booting a blockchain
	   and launching a .pdx or going interactive
	*/

	// Find the chain's db
	// If we can't find it by name or chainId
	// we default to flag, config file, or old default
	var chainRoot string
	if *name != "" || *chainId != "" {
		switch *chainType {
		case "thel", "thelonious", "monk":
			chainRoot = utils.ResolveChain("thelonious", *name, *chainId)
		case "btc", "bitcoin":
			chainRoot = utils.ResolveChain("bitcoin", *name, *chainId)
		case "eth", "ethereum":
			chainRoot = utils.ResolveChain("ethereum", *name, *chainId)
		case "gen", "genesis":
			chainRoot = utils.ResolveChain("thelonious", *name, *chainId)
		}

		if chainRoot == "" {
			log.Fatal("Could not locate chain by name %s or by id %s", *name, *chainId)
		}
	}

	// Startup the chain
	var chain epm.Blockchain
	switch *chainType {
	case "thel", "thelonious", "monk":
		if *rpc {
			chain = NewMonkRpcModule(chainRoot)
		} else {
			chain = NewMonkModule(chainRoot)
		}
	case "btc", "bitcoin":
		if *rpc {
			log.Fatal("Bitcoin rpc not implemented yet")
		} else {
			log.Fatal("Bitcoin not implemented yet")
		}
	case "eth", "ethereum":
		if *rpc {
			log.Fatal("Eth rpc not implemented yet")
		} else {
			chain = NewEthModule(chainRoot)
		}
	case "gen", "genesis":
		chain = NewGenModule(chainRoot)
	}

	epm.ContractPath, err = filepath.Abs(*contractPath)
	if err != nil {
		logger.Errorln(err)
		os.Exit(0)
	}
	logger.Debugln("Contract root:", epm.ContractPath)

	// setup EPM object with ChainInterface
	e := epm.NewEPM(chain, epm.LogFile)

	// if interactive mode, enable diffs and run the repl
	if *interactive {
		e.Diff = true
		e.Repl()
		os.Exit(0)
	}

	// comb directory for package-definition file
	// exits on error
	dir, pkg, test_ := getPkgDefFile(*packagePath)

	// epm parse the package definition file
	err = e.Parse(path.Join(dir, pkg+"."+PkgExt))
	if err != nil {
		logger.Errorln(err)
		os.Exit(0)
	}

	if *diffStorage {
		e.Diff = true
	}

	// epm execute jobs
	e.ExecuteJobs()
	// wait for a block
	e.Commit()
	if test_ {
		results, err := e.Test(path.Join(dir, pkg+"."+TestExt))
		if err != nil {
			logger.Errorln(err)
			if results != nil {
				logger.Errorln("Failed tests:", results.FailedTests)
			}
		}
	}
}

func inity() error {
	args := flag.Args()
	var p string
	if len(args) == 0 {
		p = "."
	} else {
		p = args[0]
	}
	return monk.InitChain(p)
}
