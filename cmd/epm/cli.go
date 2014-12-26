package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	color "github.com/daviddengcn/go-colortext"
	"github.com/eris-ltd/epm-go/chains"
	"github.com/eris-ltd/epm-go/epm"
	"github.com/eris-ltd/epm-go/utils"
	"github.com/eris-ltd/thelonious/monk"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var EPMVars = "epm.vars"

// TODO !
func cliCleanPullUpdate(c *cli.Context) {

}

// plop the config or genesis defaults into current dir
func cliPlop(c *cli.Context) {
	switch c.Args().First() {
	case "genesis":
		ifExit(utils.Copy(path.Join(utils.Blockchains, "genesis.json"), "genesis.json"))
	case "config":
		ifExit(utils.Copy(path.Join(utils.Blockchains, "config.json"), "config.json"))
	default:
		logger.Errorln("Unknown plop option. Should be 'config' or 'genesis'")
	}
	exit(nil)

}

// list the refs
func cliRefs(c *cli.Context) {
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
func cliHead(c *cli.Context) {
	chainHead, err := chains.GetHead()
	if err == nil {
		fmt.Println("Current head:", chainHead)
	}
	exit(err)
}

// create ~/.decerver tree and drop default monk/gen configs in there
func cliInit(c *cli.Context) {
	exit(monk.InitChain())
}

// install a dapp
func cliFetch(c *cli.Context) {
	exit(monk.FetchInstallChain(c.Args().First()))
}

// deploy the genblock into a local .decerver-local
// possibly install into ~/.decerver
//   (will move local dir and local configs)
// possibly checkout the newly deployed
func cliDeploy(c *cli.Context) {
	// if genesis or config are not specified
	// use defaults set by `epm -init`
	// and copy into working dir
	deployGen := c.String("genesis")
	deployConf := c.String("config")
	tempGen := ".genesis.json"
	tempConf := ".config.json"

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

	err = monk.InstallChain(ROOT, c.String("name"), tempGen, tempConf, chainId)
	ifExit(err)

	if c.Bool("checkout") {
		ifExit(chains.ChangeHead(chainId))
	}
	logger.Warnf("Deployed and installed chain: %s", chainId)
	exit(nil)
}

// install a local chain into the decerver tree
func cliInstall(c *cli.Context) {
	if _, err := os.Stat(ROOT); err != nil {
		exit(fmt.Errorf("No %s directory found. There must be a %s present in order to install", ROOT, ROOT))
	}

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
func cliCheckout(c *cli.Context) {
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
func cliAddRef(c *cli.Context) {
	ref := c.Args().Get(0)
	name := c.Args().Get(1)
	if name == "" {
		log.Fatal(`add-ref requires a name to be specified as well, \n
                        eg. "add-ref 14c32 mychain"`)
	}
	exit(chains.AddRef(ref, name))
}

// TODO: multi types
func cliRun(c *cli.Context) {
	run := c.Args().First()
	chainType := c.String("type")
	fmt.Println("type: ", chainType)
	if run == "" {
		chainHead, err := chains.GetHead()
		ifExit(err)
		run = chainHead
	}
	chainId, err := chains.ResolveChainId(chainType, run, run)
	if err != nil {

	}
	ifExit(err)
	logger.Infoln("Running chain ", chainId)
	chain := loadChain(c, path.Join(utils.Blockchains, "thelonious", chainId))
	chain.WaitForShutdown()
}

func cliRunDapp(c *cli.Context) {
	dapp := c.Args().First()
	chainId, err := chains.ChainIdFromDapp(dapp)
	ifExit(err)
	logger.Infoln("Running chain ", chainId)
	chain := loadChain(c, path.Join(utils.Blockchains, "thelonious", chainId))
	chain.WaitForShutdown()
}

func cliConfig(c *cli.Context) {
	global := c.Bool("global")
	var root string
	if global {
		root = utils.Blockchains
	} else {
		root = resolveRoot(c)
	}

	m := newChain(c)
	m.ReadConfig(path.Join(root, "config.json"))

	args := c.Args()
	for _, a := range args {
		sp := strings.Split(a, ":")
		key := sp[0]
		value := sp[1]
		if err := m.SetProperty(key, value); err != nil {
			logger.Errorln(err)
		}
	}
	m.WriteConfig(path.Join(root, "config.json"))
}

func cliRemove(c *cli.Context) {
	root := resolveRoot(c)
	if confirm("This will permanently delete the directory: " + root) {
		os.RemoveAll(root)
	}
}

func cliCommand(c *cli.Context) {
	root := resolveRoot(c)
	chain := loadChain(c, root)

	args := c.Args()
	if len(args) < 3 {
		exit(fmt.Errorf("You must specify a command and at least 2 arguments"))
	}
	cmd := args[0]
	args = args[1:]

	job := epm.NewJob(cmd, args)

	contractPath := c.String("c")
	if !c.IsSet("c") {
		contractPath = defaultContractPath
	}

	var err error
	epm.ContractPath, err = filepath.Abs(contractPath)
	ifExit(err)

	e, err := epm.NewEPM(chain, epm.LogFile)
	ifExit(err)
	e.ReadVars(path.Join(root, EPMVars))

	e.AddJob(job)
	e.ExecuteJobs()
	e.WriteVars(path.Join(root, EPMVars))
	e.Commit()
}

func cliDeployPdx(c *cli.Context) {
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

	if !c.IsSet("c") {
		contractPath = defaultContractPath
	}
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
	e.ReadVars(path.Join(chainRoot, EPMVars))

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
	// write epm variables to file
	e.WriteVars(path.Join(chainRoot, EPMVars))
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
