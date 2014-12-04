package main

import (
	"flag"
	"fmt"
	"github.com/eris-ltd/epm-go"
	"os"
	"path"
	"path/filepath"
)

var GoPath = os.Getenv("GOPATH")

// adjust these to suit all your deformed nefarious extension name desires. Muahahaha
// but actually don't because you might break something ;)
var PkgExt = "pdx"
var TestExt = "pdt"

var (
	defaultContractPath = "."
	defaultPackagePath  = "."
	defaultGenesis      = ""
	defaultKeys         = ""
	defaultDatabase     = ".ethchain"
	defaultLogLevel     = 0
	defaultDifficulty   = 14
	defaultMining       = false
	defaultDiffStorage  = false

	contractPath = flag.String("c", defaultContractPath, "Set the contract root path")
	packagePath  = flag.String("p", ".", "Set a .package-definition file")
	genesis      = flag.String("g", "", "Set a genesis.json file")
	keys         = flag.String("k", "", "Set a keys file")
	database     = flag.String("db", ".ethchain", "Set the location of an eth-go root directory")
	logLevel     = flag.Int("log", 0, "Set the eth log level")
	difficulty   = flag.Int("dif", 14, "Set the mining difficulty")
	mining       = flag.Bool("mine", false, "To mine or not to mine, that is the question")
	diffStorage  = flag.Bool("diff", false, "Show a diff of all contract storage")
	clean        = flag.Bool("clean", false, "Clear out epm related dirs")
	update       = flag.Bool("update", false, "Pull and install the latest epm")
	install      = flag.Bool("install", false, "Re-install epm")
	interactive  = flag.Bool("i", false, "Run epm in interactive mode")
	noGenDoug    = flag.Bool("no-gendoug", false, "Turn off gendoug mechanics")

//    rpc = flag.Bool("rpc", false, "Fire commands over rpc")
//    rpcHost = flag.String("rpcHost", "localhost", "Set the rpc host")
//    rpcPort = flag.String("rpcPort", "", "Set the rpc port")
//    host = flag.String("host", "localhost", "Set the decerver host")
//    port = flag.String("port", "", "Set the decerver port")
)

func main() {
	flag.Parse()

	if *clean || *update || *install {
		cleanUpdateInstall()
		os.Exit(0)
	}

	var err error
	epm.ContractPath, err = filepath.Abs(*contractPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// make ~/.epm-go and ~/.epm-go/.tmp for modified contract files
	epm.CheckMakeTmp()

	// Startup the chain
	chain := NewMonkModule()

	// Create ChainInterface instance
	//ethD := epm.NewEthD(eth)
	// setup EPM object with ChainInterface
	e := epm.NewEPM(chain, ".epm-log")
	// subscribe to new blocks..
	//e.Ch = Subscribe(eth, "newBlock")

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
		fmt.Println(err)
		os.Exit(0)
	}

	if *diffStorage {
		e.Diff = true
	}

	// epm execute jobs
	e.ExecuteJobs()
	// wait for a block
	e.WaitForBlock()
	if test_ {
		results, err := e.Test(path.Join(dir, pkg+"."+TestExt))
		if err != nil {
			fmt.Println(err)
			if results != nil {
				fmt.Println("failed tests:", results.FailedTests)
			}
		}
	}
	//eth.GetStorage()
}
