package main

import (
    "github.com/codegangsta/cli"
	"fmt"
	color "github.com/daviddengcn/go-colortext"
	"github.com/eris-ltd/epm-go/utils"
	"github.com/eris-ltd/epm-go/chains"
	"github.com/eris-ltd/epm-go/epm"
	"github.com/eris-ltd/thelonious/monk"
	"log"
	"os"
	"path"
	"path/filepath"

)


func cliCleanPullUpdate(c *cli.Context){

}

// plop the config or genesis defaults into current dir
func cliPlop(c *cli.Context){
    switch c.Args().First(){
    case "genesis":
        ifExit(utils.Copy(path.Join(utils.Blockchains, "genesis.json"), "genesis.json"))
    case "config":
        ifExit(utils.Copy(path.Join(utils.Blockchains, "config.json"), "config.json"))
    default:
        logger.Errorln("Unknown plop option. Should be 'config' or 'genesis'")
    }
    exit(nil)

}

// list the refs (git branch)
func cliRefs(c *cli.Context){
    r, err := chains.GetRefs()
    h, _ := chains.GetHead()
    fmt.Println("Available refs:")
    for rk, rv := range r {
        if rv == h || rk == h {
            color.ChangeColor(color.Green, true, color.None, false)
            fmt.Printf("%s \t : \t %s\n", rk, rv)
            color.ResetColor()
        } else {
            fmt.Printf("%s \t : \t %s\n", rk, rv)
        }
    }
    exit(err)
}

// print current head
func cliHead(c *cli.Context){
    chainHead, err := chains.GetHead()
    if err == nil {
        fmt.Println("Current head:", chainHead)
    }
    exit(err)
}

// create ~/.decerver tree and drop default monk/gen configs in there
func cliInit(c *cli.Context){
    exit(monk.InitChain())
}

// install a dapp
func cliFetch(c *cli.Context){
    exit(monk.FetchInstallChain(c.Args().First()))
}

// deploy the genblock into a local .decerver-local
// possibly install into ~/.decerver
//   (will move local dir and local configs)
// possibly checkout the newly deployed
func cliDeploy(c *cli.Context){
    // if genesis or config are not specified
    // use defaults set by `epm -init`
    // and copy into working dir
    deployGen := c.String("genesis")
    deployConf := c.String("config")
    tempGen := "genesis.json"
    tempConf := "config.json"

    if deployGen == "" {
        deployGen = path.Join(utils.Blockchains, "genesis.json")
    }
    if deployConf == "" {
        deployConf = path.Join(utils.Blockchains, "config.json")
    }

    ifExit(utils.Copy(deployGen, tempGen))
    vi(tempGen)

    ifExit(utils.Copy(deployConf, tempConf))
    vi(tempConf)

    chainId, err := monk.DeployChain(ROOT, tempGen, tempConf)
    ifExit(err)
    if c.Bool("install") {
        err := monk.InstallChain(ROOT, c.String("name"), tempGen, tempConf, chainId)
        ifExit(err)
    }
    if c.Bool("checkout") {
        ifExit(chains.ChangeHead(chainId))
    }
    exit(nil)
}


// install a local chain into the decerver tree
func cliInstall(c *cli.Context){
    var config string
    var genesis string
    name := c.String("name")
    // if config/genesis present locally, set them
    if _, err := os.Stat("config.json"); err == nil {
        config = "config.json"
    }
    if _, err := os.Stat("genesis.json"); err == nil {
        genesis = "genesis.json"
    }

    // if not found locally or specified, fail
    if config == "" {
        exit(fmt.Errorf("No config.json found. There must be a config.json in the present directory to install the chain"))
    }
    if genesis == "" {
        exit(fmt.Errorf("No genesis.json found. There must be a genesis.json in the present directory to install the chain"))
    }
    chainId, err := monk.ChainIdFromDb(ROOT)
    ifExit(err)
    logger.Infoln("Installing chain ", chainId)
    ifExit(monk.InstallChain(ROOT, name, genesis, config, chainId))
    if c.Bool("checkout") {
        ifExit(chains.ChangeHead(chainId))
    }
}

// change the currently active chain
func cliCheckout(c *cli.Context){
    args := c.Args()
    if len(args) == 0 {
        exit(fmt.Errorf("Please specify the chain to checkout"))
    }
    if err := chains.ChangeHead(args[0]); err != nil {
        exit(err)
    }
    logger.Infoln("Checked out new head: ", args[0])
    exit(nil)
}

// add a new reference to a chainId
func cliAddRef(c *cli.Context){
    ref := c.Args().Get(0)
    name := c.Args().Get(1)
    if name == "" {
        log.Fatal(`add-ref requires a name to be specified as well, \n
                        eg. "add-ref 14c32 mychain"`)
    }
    exit(chains.AddRef(ref, name))
}

// TODO: multi types
func cliRun(c *cli.Context){
    run := c.Args().First()
    chainType := c.String("type")
    fmt.Println("type: ", chainType)
    chainId, err := chains.ResolveChainId(chainType, run, run)
    ifExit(err)
    logger.Infoln("Running chain ", chainId)
    chain := loadChain(c, path.Join(utils.Blockchains, "thelonious", chainId))
    chain.WaitForShutdown()
}

func cliRunDapp(c *cli.Context){
    dapp := c.Args().First()
    chainId, err := chains.ChainIdFromDapp(dapp)
    ifExit(err)
    logger.Infoln("Running chain ", chainId)
    chain := loadChain(c, path.Join(utils.Blockchains, "thelonious", chainId))
    chain.WaitForShutdown()
}

func cliDeployPdx(c *cli.Context){
    var err error
	if len(c.Args()) > 0 {
		logger.Errorln("Did not understand command. Did you forget a - ?")
		logger.Errorln("Run `epm -help` to see the list of commands")
		exit(nil)
	}

    name := c.String("name")
    chainId := c.String("id")
    chainType := c.String("type")
    contractPath := c.String("c")
    dontClear := c.Bool("dont-clear")
    interactive := c.Bool("i")
    diffStorage := c.Bool("diff")
    packagePath := c.String("p")

	// Find the chain's db
	// hierarchy : name > chainId > db > config > HEAD > default
	var chainRoot string
	if name != "" || chainId != "" {
		chainRoot, err = chains.ResolveChain(chainType, name, chainId)
		ifExit(err)
	}

	// Startup the chain
	var chain epm.Blockchain
	chain = loadChain(c, chainRoot)

	epm.ContractPath, err = filepath.Abs(contractPath)
	ifExit(err)

	logger.Debugln("Contract root:", epm.ContractPath)

	// clear the cache
	if !dontClear {
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
	if interactive {
		e.Diff = true
		e.Repl()
		os.Exit(0)
	}

	// comb directory for package-definition file
	// exits on error
	dir, pkg, test_ := getPkgDefFile(packagePath)

	// epm parse the package definition file
	err = e.Parse(path.Join(dir, pkg+"."+PkgExt))
	ifExit(err)

	if diffStorage {
		e.Diff = true
	}

	// epm execute jobs
	e.ExecuteJobs()
	// wait for a block
	e.Commit()
	// run tests
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


