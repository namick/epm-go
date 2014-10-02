package main

import (
    "fmt"
    "os"
    "path"
    "github.com/project-douglas/lllc-server"
    "github.com/project-douglas/epm-go"
    "github.com/eris-ltd/eth-go-mods/ethtest"
)

func main(){
    lllcserver.PathToLLL = path.Join("/Users/BatBuddha/cpp-ethereum/build/lllc/lllc") 
    // Startup the EthChain
    eth := ethtest.NewEth(nil)
    eth.Init()
    eth.Start()
    // Create ChainInterface instance
    ethD := epm.NewEthD(eth)
    // setup EPM object with ChainInterface
    e := epm.NewEPM(ethD)
    // epm parse the package definition file
    err := e.Parse("hi.txt")
    if err != nil{
        fmt.Println(err)
        os.Exit(0)
    }
    // epm execute jobs
    e.ExecuteJobs()
}

