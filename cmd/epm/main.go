package main

import (
	"flag"
	"fmt"
	color "github.com/daviddengcn/go-colortext"
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

	logger *monklog.Logger = monklog.NewLogger("EPM-CLI")

	ROOT = ".temp"

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

	// config setup
	config    = flag.String("config", "monk-config.json", "pick config file")
	genesis   = flag.String("genesis", "genesis.json", "Set a genesis.json or genesis.pdx file")
	name      = flag.String("name", "", "Set the chain by name")
	chainId   = flag.String("id", "", "Set the chain by id")
	chainType = flag.String("type", "thel", "Set the chain type (thelonious, genesis, bitcoin, ethereum)")

	// epm commands
	clean    = flag.Bool("clean", false, "Clear out epm related dirs")
	pull     = flag.Bool("pull", false, "Pull and install the latest epm")
	update   = flag.Bool("update", false, "Re-install epm")
	install  = flag.Bool("install", false, "Install a chain into the decerver")
	ini      = flag.Bool("init", false, "Initialize a monkchain config")
	deploy   = flag.Bool("deploy", false, "Deploy a monkchain")
	checkout = flag.Bool("checkout", false, "Checkout a chain (ie. change HEAD)")
	addRef   = flag.String("add-ref", "", "Add a new reference to a chainId")
	refs     = flag.Bool("refs", false, "List the available references")
	head     = flag.Bool("head", false, "Print the currently active chain")

	// chain options
	difficulty = flag.Int("dif", 14, "Set the mining difficulty")
	mining     = flag.Bool("mine", false, "To mine or not to mine, that is the question")
	noGenDoug  = flag.Bool("no-gendoug", false, "Turn off gendoug mechanics")

	// epm options
	interactive = flag.Bool("i", false, "Run epm in interactive mode")
	diffStorage = flag.Bool("diff", false, "Show a diff of all contract storage")
	dontClear   = flag.Bool("dont-clear", false, "Stop epm from clearing the epm cache on startup")

	// paths
	contractPath = flag.String("c", defaultContractPath, "Set the contract root path")
	packagePath  = flag.String("p", ".", "Set a .package-definition file")
	keys         = flag.String("k", "", "Set a keys file")
	database     = flag.String("db", ".chain", "Set the location of the root directory")
	logLevel     = flag.Int("log", 5, "Set the eth log level")

	// remote
	rpc     = flag.Bool("rpc", false, "Fire commands over rpc")
	rpcHost = flag.String("host", "localhost", "Set the rpc host")
	rpcPort = flag.String("port", "40404", "Set the rpc port")
)

func main() {
	flag.Parse()

	utils.InitLogging(path.Join(utils.Logs, "epm"), "", *logLevel, "")

	// clean, update, or install
	// exit
	if *clean || *pull || *update {
		cleanPullUpdate()
		os.Exit(0)
	}

	if *refs {
		r, err := utils.GetRefs()
		h, _ := utils.GetHead()
		fmt.Println("Available refs:")
		for rk, rv := range r {
			if rv == h || rk == h {
				color.ChangeColor(color.Yellow, true, color.None, false)
				fmt.Printf("%s \t : \t %s\n", rk, rv)
				color.ResetColor()
			} else {
				fmt.Printf("%s \t : \t %s\n", rk, rv)
			}
		}
		exit(err)
	}

	if *head {
		chainHead, err := utils.GetHead()
		if err == nil {
			fmt.Println("Current head:", chainHead)
		}
		exit(err)
	}

	var err error

	// create ~/.decerver tree and drop monk/gen configs
	// exit
	if *ini {
		exit(inity())
	}

	// deploy the genblock, install into ~/.decerver
	// possibly checkout the newly deployed
	// exit
	if *deploy {
		chainId, err := monk.DeployChain(ROOT, *genesis, *config)
		ifExit(err)
		if *install {
			err := monk.InstallChain(ROOT, *name, *genesis, *config, chainId)
			ifExit(err)
		}
		if *checkout {
			exit(utils.ChangeHead(chainId))
		}
		exit(nil)
	}

	if *install {
		chainId, err := monk.ChainIdFromDb(ROOT)
		ifExit(err)
		exit(monk.InstallChain(ROOT, *name, *genesis, *config, chainId))
	}

	// change the currently active chain
	// exit
	if *checkout {
		args := flag.Args()
		if len(args) == 0 {
			exit(fmt.Errorf("Please specify the chain to checkout"))
		}
		if err := utils.ChangeHead(args[0]); err != nil {
			exit(err)
		}
		logger.Infoln("Checked out new head: ", args[0])
		exit(nil)
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
	// should be: name flag > chainId flag > db flag > config file > HEAD > old default
	var chainRoot string
	if *name != "" || *chainId != "" {
		// these will check the name and id
		switch *chainType {
		case "thel", "thelonious", "monk":
			*chainType = "thelonious"
		case "btc", "bitcoin":
			*chainType = "bitcoin"
		case "eth", "ethereum":
			*chainType = "ethereum"
		case "gen", "genesis":
			*chainType = "thelonious"
		}

		chainRoot = utils.ResolveChain(*chainType, *name, *chainId)

		if chainRoot == "" {
			// chainRoot, err = utils.GetHead()
			// if err != nil{
			//     exit(fmt.Errorf("Error reading head file!"))
			// }
			// if chainRoot == ""{
			exit(fmt.Errorf("Could not locate chain by name %s or by id %s", *name, *chainId))
			// }
		}
	}

	// Startup the chain
	var chain epm.Blockchain
	logger.Debugln("Loading chain ", *chainType)
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
	ifExit(err)

	logger.Debugln("Contract root:", epm.ContractPath)

	// clear the cache
	if !*dontClear {
		err := os.RemoveAll(utils.Epm)
		if err != nil {
			logger.Errorln("Error clearing cache: ", err)
		}
		utils.InitDataDir(utils.Epm)
	}

	// setup EPM object with ChainInterface
	e, err := epm.NewEPM(chain, epm.LogFile)
	ifExit(err)

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
	ifExit(err)

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
