package main

import (
    "fmt"
    "os"
    "path"
    "github.com/eris-ltd/epm-go"
    "github.com/eris-ltd/thelonious/monk"
    "github.com/eris-ltd/thelonious/monkchain"
)

var GoPath = os.Getenv("GOPATH")

func NewMonkModule() *monk.MonkModule{
    //lllcserver.PathToLLL = path.Join("/Users/BatBuddha/cpp-ethereum/build/lllc/lllc") 
    m := monk.NewMonk(nil)
    monkchain.GENDOUG = nil
    monkchain.GenesisConfig = "genesis.json"
    m.Config.RootDir = ".ethchain"
    m.Config.LogLevel = 0
    m.Config.DougDifficulty = 14
    m.Init() 
    m.Config.Mining = false
    m.Start()
    return m
}

// test the epm test file mechanism
func main(){
    // Startup the EthChain
    m := NewMonkModule()
    // Create ChainInterface instance
    //ethD := epm.NewEthD(eth)
    // setup EPM object with ChainInterface
    e := epm.NewEPM(m, ".epm-log-deploy-test")
    // subscribe to new blocks..
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

