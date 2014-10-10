package main

import (
    "fmt"
    "os"
    "path"
    "path/filepath"
    "flag"
    "io/ioutil"
    "strings"
    "github.com/eris-ltd/epm-go"
    "github.com/eris-ltd/eth-go-mods/ethtest"
    "github.com/eris-ltd/eth-go-mods/ethchain"
    "github.com/eris-ltd/eth-go-mods/ethreact"
)

var GoPath = os.Getenv("GOPATH")

// adjust these to suit all your deformed nefarious extension name desires. Muahahaha
var PkgExt = "pkg-def"
var TestExt = "pkg-def-test"

/*
    epm-go cli:
        General:
            - `epm-go` will look for a .package-definition file in the current directory, and expect all contracts to have paths beginning in the current dir
        Paths:
            - `-c` allows one to set the contract root (ie. the pkg-defn file has contract paths starting from here)
            - `-p` allows one to specify the location of a .pkg-defn file. The corresponding test file is expected to be in the same location
        Eth:
            - by default, a fresh eth-instance will be started with no genesis doug. To specify:
                - `-g` allows one to set a genesis.json configuration file
                - `-k` allows one to set a keys.txt file (with one hex-encoded private key per line)
                - `-db` allows one to set the location of an eth levelDB database to use
            - the `-rpc`, `-rpcHost` and `-rpcPort` flags specify to use rpc and the params. `-rpc` alone will use the defaults, while using one of host/port will choose the default for the other
            - the `d`, `-host` and `-port` commands allow one to specify to pass commands through a deCerver, and to set the host/port 
*/

var (
    contractPath = flag.String("c", ".", "Set the contract root path")
    packagePath = flag.String("p", ".", "Set a .package-definition file")
    genesis = flag.String("g", "", "Set a genesis.json file")
    keys = flag.String("k", "", "Set a keys file")
    database = flag.String("db", ".ethchain", "Set the location of an eth-go root directory")
    logLevel = flag.Int("log", 0, "Set the eth log level")
    difficulty = flag.Int("dif", 14, "Set the mining difficulty")
    mining = flag.Bool("mine", true, "To mine or not to mine, that is the question")
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

    // make ~/.epm-go and ~/.epm-go/.tmp for modified contract files
    epm.CheckMakeTmp()

    // comb directory for package-definition file
    // exits on error
    dir, pkg, test_ := getPkgDefFile(*packagePath)        

    // Startup the EthChain
    // uses flag variables (global) for config
    eth := NewEthNode()
    // Create ChainInterface instance
    ethD := epm.NewEthD(eth)
    // setup EPM object with ChainInterface
    e := epm.NewEPM(ethD)

    // epm parse the package definition file
    err = e.Parse(path.Join(dir, pkg+"."+PkgExt))
    if err != nil{
        fmt.Println(err)
        os.Exit(0)
    }

    // epm execute jobs
    e.ExecuteJobs()
    // wait for a block
    ch := make(chan ethreact.Event, 1)
    eth.Ethereum.Reactor().Subscribe("newBlock", ch)
    _ =<- ch
    if test_{
        e.Test(path.Join(dir, pkg+"."+TestExt))
    }
    //eth.GetStorage()
}

// configure and start an in-process eth node
// all paths should be made absolute
func NewEthNode() *ethtest.EthChain{
    // empty ethchain object
    eth := ethtest.NewEth(nil)
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
    // initialize and start
    eth.Init() 
    eth.Start()
    return eth
}

// looks for pkg-def file
// exits if error (none or more than 1)
// returns dir of pkg, name of pkg (no extension) and whether or not there's a test file
func getPkgDefFile(pkgPath string) (string, string, bool) {
    fmt.Println("pkg path:", pkgPath)
    var name string
    var test_ bool

    // if its not a directory, look for a corresponding test file
    f, err := os.Stat(pkgPath)
    if err != nil{
        fmt.Println(err)
        os.Exit(0)
    }
    if !f.IsDir(){
        dir, fil := path.Split(pkgPath)
        spl := strings.Split(fil, ".")
        pkg := spl[0]
        ext := spl[1] 
        if ext != PkgExt{
            fmt.Printf("Did not understand extension. Got %s, expected %s\n", ext, PkgExt)
            os.Exit(0)
        }
        
        _, err := os.Stat(path.Join(dir, pkg, TestExt))
        if err != nil{
            fmt.Printf("There was no test found for package-definition %s. Deploying without test ...\n", name)
            test_ = false
        } else {
            test_ = true
        }
        return dir, name, test_
    }


    // read dir for files
    files, err := ioutil.ReadDir(pkgPath)
    if err != nil{
        fmt.Println("Could not read directory:", err)
        os.Exit(0)
    }
    // find all package-defintion and package-definition-test files
    candidates := make(map[string]int)
    candidates_test := make(map[string]int)
    for _, f := range files{
        name := f.Name()
        spl := strings.Split(name, ".")
        if len(spl) < 2{
            continue
        }
        name = spl[0]
        ext := spl[1]
        if ext == PkgExt{
            candidates[name] = 1
        } else if ext == TestExt {
            candidates_test[name] = 1
        }
    }
    // exit if too many or no options
    if len(candidates) > 1{
        fmt.Println("More than one package-definition file available. Please select with the '-p' flag")
        os.Exit(0)
    } else if len(candidates) == 0{
        fmt.Println("No package-definition files found for extensions", PkgExt, TestExt)
        os.Exit(0)
    }
    // this should run once (there's only one candidate)
    for k, _ := range candidates{
        name = k
        if candidates_test[name] == 1{
            test_ = true
        } else{
            fmt.Printf("There was no test found for package-definition %s. Deploying without test ...\n", name)
            test_ = false
        }
    }
    return pkgPath, name, test_
}