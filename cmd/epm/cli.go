package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	color "github.com/daviddengcn/go-colortext"
	"github.com/eris-ltd/epm-go/chains"
	"github.com/eris-ltd/epm-go/epm"
	"github.com/eris-ltd/epm-go/utils"
	"github.com/eris-ltd/thelonious/monk"
	"github.com/eris-ltd/thelonious/monkdoug"
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
	_, h, _ := chains.GetHead()
	fmt.Println("Available refs:")
	for rk, rv := range r {
		if strings.Contains(rv, h) {
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
	typ, id, err := chains.GetHead()
	if err == nil {
		fmt.Println("Current head:", path.Join(typ, id))
	}
	exit(err)
}

// create ~/.decerver tree and drop default monk/gen configs in there
func cliInit(c *cli.Context) {
	exit(utils.InitDecerverDir())
}

// install a dapp
// TODO hmph
func cliFetch(c *cli.Context) {
	exit(monk.FetchInstallChain(c.Args().First()))
}

// deploy the genblock into a local .decerver-local
// and install into the global tree
// possibly checkout the newly deployed
// chain agnostic!
func cliDeploy(c *cli.Context) {
	chainType, err := chains.ResolveChainType(c.String("type"))
	ifExit(err)
	name := c.String("name")
	rpc := c.Bool("rpc")

	// if genesis or config are not specified
	// use defaults set by `epm -init`
	// and copy into working dir
	deployConf := c.String("config")
	tempConf := ".config.json"

	if deployConf == "" {
		deployConf = path.Join(utils.Blockchains, chainType, "config.json")
	}

	// if config doesnt exist, lay it
	if _, err := os.Stat(deployConf); err != nil {
		utils.InitDataDir(path.Join(utils.Blockchains, chainType))
		m := newChain(chainType, rpc)
		ifExit(m.WriteConfig(deployConf))
	}

	ifExit(utils.Copy(deployConf, tempConf))
	vi(tempConf)

	var chainId string
	if chainType == "thelonious" {
		deployGen := c.String("genesis")
		tempGen := path.Join(ROOT, "genesis.json")
		utils.InitDataDir(ROOT)

		if deployGen == "" {
			deployGen = path.Join(utils.Blockchains, "thelonious", "genesis.json")
		}
		if _, err := os.Stat(deployGen); err != nil {
			err := utils.WriteJson(monkdoug.DefaultGenesis, deployGen)
			ifExit(err)
		}
		ifExit(utils.Copy(deployGen, tempGen))
		vi(tempGen)
		chainId, err = monk.DeployChain(ROOT, tempGen, tempConf)
		ifExit(err)
		err = monk.InstallChain(ROOT, name, tempGen, tempConf, chainId)
		ifExit(err)
	} else {
		chain := newChain(chainType, rpc)
		chainId, err = DeployChain(chain, ROOT, tempConf)
		ifExit(err)
		if chainId == "" {
			exit(fmt.Errorf("ChainId must not be empty. How else would we ever find you?!"))
		}
		fmt.Println(ROOT, name, chainType, tempConf, chainId)
		err = InstallChain(chain, ROOT, name, chainType, tempConf, chainId)
		ifExit(err)
	}

	logger.Warnf("Deployed and installed chain: %s/%s", chainType, chainId)
	if c.Bool("checkout") {
		ifExit(chains.ChangeHead(chainType, chainId))
		logger.Warnf("Checked out chain: %s/%s", chainType, chainId)
	}
	exit(nil)
}

// change the currently active chain
func cliCheckout(c *cli.Context) {
	args := c.Args()
	if len(args) == 0 {
		exit(fmt.Errorf("Please specify the chain to checkout"))
	}
	head := args[0]

	typ, id, err := chains.ResolveChain(head)
	ifExit(err)

	if err := chains.ChangeHead(typ, id); err != nil {
		exit(err)
	}
	logger.Infoln("Checked out new head: ", path.Join(typ, id))
	exit(nil)
}

// add a new reference to a chainId
func cliAddRef(c *cli.Context) {
	chain := c.Args().Get(0)
	ref := c.Args().Get(1)
	if ref == "" {
		log.Fatal(`add-ref requires a name to be specified as well, \n
                        eg. "add-ref 14c32 mychain"`)
	}

	typ, id, err := chains.SplitRef(chain)
	if err != nil {
		exit(fmt.Errorf(`Error: specify the type in the first 
                argument as '<type>/<chainId>'`))
	}
	exit(chains.AddRef(typ, id, ref))
}

// run a node on a chain
func cliRun(c *cli.Context) {
	run := c.Args().First()
	chainType, chainId, err := chains.ResolveChain(run)
	ifExit(err)
	logger.Infof("Running chain %s/%s\n", chainType, chainId)
	chain := loadChain(c, chainType, path.Join(utils.Blockchains, chainType, chainId))
	chain.WaitForShutdown()
}

// TODO: multi types
func cliRunDapp(c *cli.Context) {
	dapp := c.Args().First()
	chainType := "thelonious"
	chainId, err := chains.ChainIdFromDapp(dapp)
	ifExit(err)
	logger.Infoln("Running chain ", chainId)
	chain := loadChain(c, chainType, path.Join(utils.Blockchains, chainType, chainId))
	chain.WaitForShutdown()
}

// edit a config value
func cliConfig(c *cli.Context) {
	var (
		root      string
		chainType string
		chainId   string
		err       error
	)
	rpc := c.Bool("rpc")
	if c.IsSet("type") {
		chainType = c.String("type")
		root = path.Join(utils.Blockchains, chainType)
	} else {
		chain := c.String("chain")
		chainType, chainId, err = chains.ResolveChain(chain)
		ifExit(err)
		root = path.Join(utils.Blockchains, chainType, chainId)
	}

	m := newChain(chainType, rpc)
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

// remove a chain
func cliRemove(c *cli.Context) {
	root := resolveRoot(c)

	if confirm("This will permanently delete the directory: " + root) {
		// remove the directory
		os.RemoveAll(root)
		// remove from head (if current head)
		_, h, _ := chains.GetHead()
		if strings.Contains(root, h) {
			chains.NullHead()
		}
		// remove refs
		refs, err := chains.GetRefs()
		ifExit(err)
		for k, v := range refs {
			if strings.Contains(root, v) {
				os.Remove(path.Join(utils.Blockchains, "refs", k))
			}
		}
	}
}

// run a single epm on-chain command (endow, deploy)
func cliCommand(c *cli.Context) {
	chainType, _, _ := chains.ResolveChain(c.String("ref"))
	root := resolveRoot(c)

	chain := loadChain(c, chainType, root)

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

// deploy a pdx file on a chain
func cliDeployPdx(c *cli.Context) {
	if len(c.Args()) > 0 {
		logger.Errorln("Did not understand command.")
		logger.Errorln("Run `epm -help` to see the list of commands")
		exit(nil)
	}

	contractPath := c.String("c")
	dontClear := c.Bool("dont-clear")
	interactive := c.Bool("i")
	diffStorage := c.Bool("diff")
	packagePath := c.String("p")

	chainRoot, err := resolveRoot(c)
	ifExit(err)
	chainType, _, _ := chains.ResolveChain(c.String("chain"))
	// hierarchy : name > chainId > db > config > HEAD > default

	// Startup the chain
	var chain epm.Blockchain
	chain = loadChain(c, chainType, chainRoot)

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
