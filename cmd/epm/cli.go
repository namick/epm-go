package main

import (
	"crypto/rand"
	"encoding/hex"
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

// TODO: pull, update

func cliClean(c *cli.Context) {
	toclean := c.Args().First()
	if toclean == "" {
		exit(fmt.Errorf("You must enter a directory or file to wipe"))
	}
	dir := path.Join(utils.Decerver, toclean)
	exit(utils.ClearDir(dir))
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

// duplicate a chain
func cliCp(c *cli.Context) {
	args := c.Args()
	var (
		root  string
		typ   string
		id    string
		err   error
		multi string
	)
	if len(args) == 0 {
		log.Fatal(`To copy a chain, specify a chain and a new name, \n eg. "cp thel/14c32 chaincopy"`)

	} else if len(args) == 1 {
		multi = args.Get(0)
		// copy the checked out chain
		typ, id, err = chains.GetHead()
		ifExit(err)
		if id == "" {
			log.Fatal(`No chain is checked out. To copy a chain, specify a chainId and an new name, \n eg. "cp thel/14c32 chaincopy"`)
		}
		root = chains.ComposeRoot(typ, id)
	} else {
		ref := args.Get(0)
		multi = args.Get(1)
		root, typ, id, err = resolveRoot(ref, false, "")
		ifExit(err)
	}
	newRoot := chains.ComposeRootMulti(typ, id, multi)
	if c.Bool("bare") {
		err = utils.InitDataDir(newRoot)
		ifExit(err)
		err = utils.Copy(path.Join(root, "config.json"), path.Join(newRoot, "config.json"))
		ifExit(err)
	} else {
		err = utils.Copy(root, newRoot)
		ifExit(err)
	}
	chain := newChain(typ, false)
	configureRootDir(c, chain, newRoot)
	chain.WriteConfig(path.Join(newRoot, "config.json"))
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

// deploy the genblock into a random folder in scratch
// and install into the global tree (must compute chainId before we know where to put it)
// possibly checkout the newly deployed
// chain agnostic!
func cliNew(c *cli.Context) {
	chainType, err := chains.ResolveChainType(c.String("type"))
	ifExit(err)
	name := c.String("name")
	forceName := c.String("force-name")
	rpc := c.GlobalBool("rpc")

	r := make([]byte, 8)
	rand.Read(r)
	tmpRoot := path.Join(utils.Scratch, "epm", hex.EncodeToString(r))

	// if genesis or config are not specified
	// use defaults set by `epm init`
	// and copy into working dir
	deployConf := c.String("config")
	deployGen := c.String("genesis")
	tempConf := ".config.json"

	if deployConf == "" {
		if rpc {
			deployConf = path.Join(utils.Blockchains, chainType, "rpc", "config.json")
		} else {
			deployConf = path.Join(utils.Blockchains, chainType, "config.json")
		}
	}

	chain := newChain(chainType, rpc)

	// if config doesnt exist, lay it
	if _, err := os.Stat(deployConf); err != nil {
		utils.InitDataDir(path.Join(utils.Blockchains, chainType))
		if rpc {
			utils.InitDataDir(path.Join(utils.Blockchains, chainType, "rpc"))
		}
		ifExit(chain.WriteConfig(deployConf))
	}
	// copy and edit temp
	ifExit(utils.Copy(deployConf, tempConf))
	vi(tempConf)

	// deploy and install chain
	chainId, err := DeployChain(chain, tmpRoot, tempConf, deployGen)
	ifExit(err)
	if chainId == "" {
		exit(fmt.Errorf("ChainId must not be empty. How else would we ever find you?!"))
	}
	err = InstallChain(chain, tmpRoot, chainType, tempConf, chainId, rpc)
	ifExit(err)

	s := fmt.Sprintf("Deployed and installed chain: %s/%s", chainType, chainId)
	if rpc {
		s += " with rpc"
	}
	logger.Warnln(s)

	if c.Bool("checkout") {
		ifExit(chains.ChangeHead(chainType, chainId))
		logger.Warnf("Checked out chain: %s/%s", chainType, chainId)
	}

	// update refs
	if forceName != "" {
		err := chains.AddRefForce(chainType, chainId, forceName)
		if err != nil {
			ifExit(err)
		}
		logger.Warnf("Created ref %s to point to chain %s\n", forceName, chainId)
	} else if name != "" {
		err := chains.AddRef(chainType, chainId, name)
		if err != nil {
			ifExit(err)
		}
		logger.Warnf("Created ref %s to point to chain %s\n", name, chainId)
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

// remove a reference from a chainId
func cliRmRef(c *cli.Context) {
	args := c.Args()
	if len(args) == 0 {
		exit(fmt.Errorf("Please specify the ref to remove"))
	}
	ref := args[0]

	_, _, err := chains.ResolveChain(ref)
	ifExit(err)
	err = os.Remove(path.Join(utils.Refs, ref))
	ifExit(err)
}

// add a new reference to a chainId
func cliAddRef(c *cli.Context) {
	args := c.Args()
	var typ string
	var id string
	var err error
	var ref string
	if len(args) == 1 {
		ref = args.Get(0)
		typ, id, err = chains.GetHead()
		ifExit(err)
		if id == "" {
			log.Fatal(`No chain is checked out. To add a ref, specify both a chainId and a name, \n eg. "epm add thel/14c32 mychain"`)
		}
	} else {
		chain := args.Get(0)
		ref = args.Get(1)
		typ, id, err = chains.SplitRef(chain)

		if err != nil {
			exit(fmt.Errorf(`Error: specify the type in the first 
                argument as '<type>/<chainId>'`))
		}
	}
	exit(chains.AddRef(typ, id, ref))
}

// run a node on a chain
func cliRun(c *cli.Context) {
	root, chainType, chainId, err := resolveRootFlag(c)
	ifExit(err)
	logger.Infof("Running chain %s/%s\n", chainType, chainId)
	chain := loadChain(c, chainType, root)
	chain.WaitForShutdown()
}

// TODO: multi types
// TODO: deprecate in exchange for -dapp flag on run
func cliRunDapp(c *cli.Context) {
	dapp := c.Args().First()
	chainType := "thelonious"
	chainId, err := chains.ChainIdFromDapp(dapp)
	ifExit(err)
	logger.Infoln("Running chain ", chainId)
	chain := loadChain(c, chainType, chains.ComposeRoot(chainType, chainId))
	chain.WaitForShutdown()
}

// edit a config value
func cliConfig(c *cli.Context) {
	var (
		root      string
		chainType string
		err       error
	)
	rpc := c.GlobalBool("rpc")
	if c.IsSet("type") {
		chainType = c.String("type")
		root = path.Join(utils.Blockchains, chainType)
	} else {
		root, chainType, _, err = resolveRootFlag(c)
		ifExit(err)
	}

	configPath := path.Join(root, "config.json")
	if c.Bool("vi") {
		vi(configPath)
	} else {
		m := newChain(chainType, rpc)
		err = m.ReadConfig(configPath)
		ifExit(err)

		args := c.Args()
		for _, a := range args {
			sp := strings.Split(a, ":")
			if len(sp) != 2 {
				logger.Errorln("Invalid arg")
				continue
			}
			key := sp[0]
			value := sp[1]
			if err := m.SetProperty(key, value); err != nil {
				logger.Errorln(err)
			}
		}
		m.WriteConfig(path.Join(root, "config.json"))
	}
}

// remove a chain
func cliRemove(c *cli.Context) {
	if len(c.Args()) < 1 {
		exit(fmt.Errorf("Error: specify the chain ref as an argument"))
	}
	root, _, _, err := resolveRootArg(c)
	ifExit(err)

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
	root, chainType, _, err := resolveRootFlag(c)
	ifExit(err)

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
func cliDeploy(c *cli.Context) {
	packagePath := "."
	if len(c.Args()) > 0 {
		packagePath = c.Args()[0]
	}

	contractPath := c.String("c")
	dontClear := c.Bool("dont-clear")
	diffStorage := c.Bool("diff")

	chainRoot, chainType, _, err := resolveRootFlag(c)
	ifExit(err)
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

func cliConsole(c *cli.Context) {

	contractPath := c.String("c")
	dontClear := c.Bool("dont-clear")
	diffStorage := c.Bool("diff")

	chainRoot, chainType, _, err := resolveRootFlag(c)
	ifExit(err)
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

	if diffStorage {
		e.Diff = true
	}
	e.Repl()
}
