package main

import (
    "fmt"
    "os"
    "path"
    "github.com/eris-ltd/epm-go"
    "github.com/eris-ltd/thelonious/monk"
    "github.com/eris-ltd/thelonious/ethchain"
    "github.com/eris-ltd/thelonious/ethreact"
)

var GoPath = os.Getenv("GOPATH")

func Subscribe(eth *monk.EthChain, event string) chan ethreact.Event{
    ch := make(chan ethreact.Event, 1) 
    eth.Ethereum.Reactor().Subscribe(event, ch)
    return ch
}

func NewEthNode() *monk.EthChain{
    //lllcserver.PathToLLL = path.Join("/Users/BatBuddha/cpp-ethereum/build/lllc/lllc") 
    eth := monk.NewEth(nil)
    ethchain.GENDOUG = nil
    ethchain.GenesisConfig = "genesis.json"
    eth.Config.RootDir = ".ethchain"
    eth.Config.LogLevel = 0
    eth.Config.DougDifficulty = 14
    eth.Init() 
    eth.Config.Mining = false
    eth.Start()
    return eth
}

// test the epm test file mechanism
func main(){
    // Startup the EthChain
    eth := NewEthNode()
    // Create ChainInterface instance
    ethD := epm.NewEthD(eth)
    // setup EPM object with ChainInterface
    e := epm.NewEPM(ethD, ".epm-log-deploy-test")
    // subscribe to new blocks..
    e.Ch = Subscribe(eth, "newBlock")
    // epm parse the package definition file
    err := e.Parse(path.Join(epm.TestPath, "test_parse.epm"))
    if err != nil{
        fmt.Println(err)
        os.Exit(0)
    }
    // epm execute jobs
    e.ExecuteJobs()
    e.WaitForBlock()
    e.Test(path.Join(epm.TestPath, "test_parse.epm-check"))

    //epm.PrintDiff(e.PrevState(), e.CurrentState())
    //eth.GetStorage()
}

