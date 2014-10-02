package main

import (
    "fmt"
    "os"
    "path"
    "github.com/project-douglas/lllc-server"
    "github.com/project-douglas/epm-go"
    "github.com/eris-ltd/eth-go-mods/ethtest"
    "github.com/eris-ltd/eth-go-mods/ethchain"
    "github.com/eris-ltd/eth-go-mods/ethreact"
)

var GoPath = os.Getenv("GOPATH")

func NewEthNode() *ethtest.EthChain{
    lllcserver.PathToLLL = path.Join("/Users/BatBuddha/cpp-ethereum/build/lllc/lllc") 
    eth := ethtest.NewEth(nil)
    ethchain.GENDOUG = nil
    ethchain.GenesisConfig = "genesis.json"
    eth.Config.RootDir = ".ethchain"
    eth.Config.LogLevel = 5
    eth.Config.DougDifficulty = 14
    eth.Init() 
    eth.Config.Mining = true
    eth.Start()
    return eth
}


func main(){
    // Startup the EthChain
    eth := NewEthNode()
    // Create ChainInterface instance
    ethD := epm.NewEthD(eth)
    // setup EPM object with ChainInterface
    e := epm.NewEPM(ethD)
    // epm parse the package definition file
    err := e.Parse(path.Join(epm.TestPath, "test_parse.epm"))
    if err != nil{
        fmt.Println(err)
        os.Exit(0)
    }
    // epm execute jobs
    e.ExecuteJobs()
    ch := make(chan ethreact.Event, 1)
    eth.Ethereum.Reactor().Subscribe("newBlock", ch)
    _ =<- ch
    e.Test(path.Join(epm.TestPath, "test_parse.epm-check"))
}

