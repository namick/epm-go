package main

import (
    "fmt"
    "os"
    "path/filepath"
    "flag"
    "github.com/eris-ltd/epm-go"
    "github.com/eris-ltd/thelonious/monk"
    "github.com/eris-ltd/thelonious/ethchain"
    "github.com/eris-ltd/thelonious/ethreact"
)

var GoPath = os.Getenv("GOPATH")

// adjust these to suit all your deformed nefarious extension name desires. Muahahaha
var PkgExt = "pdx"
var TestExt = "pdt"

var (
    contractPath = flag.String("c", ".", "Set the contract root path")
    packagePath = flag.String("p", ".", "Set a .package-definition file")
    genesis = flag.String("g", "", "Set a genesis.json file")
    keys = flag.String("k", "", "Set a keys file")
    database = flag.String("db", ".ethchain", "Set the location of an eth-go root directory")
    logLevel = flag.Int("log", 5, "Set the eth log level")
    difficulty = flag.Int("dif", 14, "Set the mining difficulty")
    mining = flag.Bool("mine", false, "To mine or not to mine, that is the question")
    clean = flag.Bool("clean", false, "Clear out epm related dirs")
    update = flag.Bool("update", false, "Pull and install the latest epm")
    install = flag.Bool("install", false, "Re-install epm")
//    rpc = flag.Bool("rpc", false, "Fire commands over rpc")
//    rpcHost = flag.String("rpcHost", "localhost", "Set the rpc host")
//    rpcPort = flag.String("rpcPort", "", "Set the rpc port")
//    host = flag.String("host", "localhost", "Set the deCerver host")
//    port = flag.String("port", "", "Set the deCerver port")
)

func main(){
    flag.Parse()

    var err error
    epm.ContractPath, err = filepath.Abs(*contractPath)
    if err != nil{
        fmt.Println(err)
        os.Exit(0)
    }

    epm.CheckMakeTmp()

    // Startup the EthChain
    // uses flag variables (global) for config
    eth := NewEthNode()
    // Create ChainInterface instance
    ethD := epm.NewEthD(eth)
    // setup EPM object with ChainInterface
    e := epm.NewEPM(ethD, ".epm-log")
    // subscribe to new blocks..
    e.Ch = Subscribe(eth, "newBlock")

    e.Diff = true

    e.Repl()
}


// subscribe on the channel
func Subscribe(eth *monk.EthChain, event string) chan ethreact.Event{
    ch := make(chan ethreact.Event, 1) 
    eth.Ethereum.Reactor().Subscribe(event, ch)
    return ch
}

// configure and start an in-process eth node
// all paths should be made absolute
func NewEthNode() *monk.EthChain{
    // empty ethchain object
    eth := monk.NewEth(nil)
    // configure, configure, configure
    ethchain.GENDOUG = nil
    var err error
    if *keys != ""{
        eth.Config.KeyFile, err = filepath.Abs(*keys)
        if err != nil{
            fmt.Println(err)
            os.Exit(0)
        }
    }
    if *genesis != ""{
        eth.Config.GenesisConfig, err = filepath.Abs(*genesis)
        if err != nil{
            fmt.Println(err)
            os.Exit(0)
        }
        eth.Config.ContractPath, err = filepath.Abs(*contractPath)
        if err != nil{
            fmt.Println(err)
            os.Exit(0)
        }
    }
    eth.Config.RootDir, err = filepath.Abs(*database)
    if err != nil{
        fmt.Println(err)
        os.Exit(0)
    }
    eth.Config.LogLevel = *logLevel
    eth.Config.DougDifficulty = *difficulty
    eth.Config.Mining = *mining

    // set LLL path
    epm.LLLURL = eth.Config.LLLPath

    // initialize and start
    eth.Init() 
    eth.Start()
    return eth
}
