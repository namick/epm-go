package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	color "github.com/daviddengcn/go-colortext"
	"github.com/eris-ltd/decerver/interfaces/dapps"
	"github.com/eris-ltd/epm-go/chains"
	"github.com/eris-ltd/epm-go/epm"
	"github.com/eris-ltd/epm-go/utils"
	mutils "github.com/eris-ltd/modules/monkutils" // for fetch
	"github.com/eris-ltd/thelonious"               // for fetch
	"github.com/eris-ltd/thelonious/monkcrypto"    // keygen
	"github.com/eris-ltd/thelonious/monkutil"      // for fetch
	"io/ioutil"
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
	root, _, chainId, err := resolveRootFlag(c)
	ifExit(err)
	switch c.Args().First() {
	case "genesis":
		ifExit(utils.Copy(path.Join(utils.Blockchains, "thelonious", "genesis.json"), "genesis.json"))
	case "config":
		ifExit(utils.Copy(path.Join(utils.Blockchains, "thelonious", "config.json"), "config.json"))
	case "chainid":
		fmt.Println(chainId)
	case "vars":
		b, err := ioutil.ReadFile(path.Join(root, EPMVars))
		ifExit(err)
		fmt.Println(string(b))
	default:
		logger.Errorln("Plop options: config, genesis, chainid, vars")
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
		fmt.Println(path.Join(typ, id))
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

// fetch a genesis block and state from a peer server
func cliFetch(c *cli.Context) {
	if len(c.Args()) == 0 {
		ifExit(fmt.Errorf("Must specify a peerserver address"))
	}

	chainType := "thelonious"
	peerserver := c.Args()[0]
	peerserver = "http://" + peerserver

	chainId, err := thelonious.GetChainId(peerserver)
	ifExit(err)

	rootDir := chains.ComposeRoot(chainType, monkutil.Bytes2Hex(chainId))

	monkutil.Config = &monkutil.ConfigManager{ExecPath: rootDir, Debug: true, Paranoia: true}
	utils.InitLogging(rootDir, "", 5, "")
	db := mutils.NewDatabase("database", false)
	monkutil.Config.Db = db

	genesisBlock, err := thelonious.GetGenesisBlock(peerserver)
	ifExit(err)

	db.Put([]byte("GenesisBlock"), genesisBlock.RlpEncode())
	db.Put([]byte("ChainID"), chainId)

	hash := genesisBlock.GetRoot()
	hashB, ok := hash.([]byte)
	if !ok {
		ifExit(fmt.Errorf("State root is not []byte:", hash))
	}
	logger.Warnf("Fetching state %x\n", hashB)
	err = thelonious.GetGenesisState(peerserver, monkutil.Bytes2Hex(hashB), db)
	ifExit(err)
	db.Close()

	// get genesis.json
	g, err := thelonious.GetGenesisJson(peerserver)
	ifExit(err)
	err = ioutil.WriteFile(path.Join(rootDir, "genesis.json"), g, 0600)
	ifExit(err)

	// drop config
	chain := newChain(chainType, false)
	chain.SetProperty("RootDir", rootDir)
	ifExit(chain.WriteConfig(path.Join(rootDir, "config.json")))

	logger.Warnf("Fetched genesis block for chain %x", chainId)
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
	deployConf := c.String("config")
	deployGen := c.String("genesis")
	tempConf := ".config.json"
	editCfg := c.Bool("edit-config")
	// if we provide genesis, dont open editor for genesis
	noEditor := c.IsSet("genesis")

	chainId := deployInstallChain(tmpRoot, deployConf, deployGen, tempConf, chainType, rpc, editCfg, noEditor)

	if c.Bool("checkout") {
		ifExit(chains.ChangeHead(chainType, chainId))
		logger.Warnf("Checked out chain: %s/%s", chainType, chainId)
	}

	// update refs
	updateRefs(chainType, chainId, forceName, name)
}

func updateRefs(chainType, chainId, forceName, name string) {
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
}

func deployInstallChain(tmpRoot, deployConf, deployGen, tempConf, chainType string, rpc, editCfg, noEditor bool) string {
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
	if editCfg {
		ifExit(editor(tempConf))
	}

	// deploy and install chain
	chainId, err := DeployChain(chain, tmpRoot, tempConf, deployGen, noEditor)
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
	ifExit(chain.Shutdown())
	return chainId
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
		ifExit(editor(configPath))
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
		// we only remove refs if its not a multi
		if !c.IsSet("multi") {
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
		// if there are no chains left, wipe the dir
		dir := path.Dir(root)
		fs, _ := ioutil.ReadDir(dir)
		if len(fs) == 0 {
			if confirm("Remove the directory " + dir + "?") {
				os.RemoveAll(dir)
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

func cliKeygen(c *cli.Context) {
	name := ""
	if len(c.Args()) > 0 {
		name = c.Args()[0]
	}

	key := monkcrypto.GenerateNewKeyPair()
	prv := key.PrivateKey
	addr := key.Address()
	a := hex.EncodeToString(addr)
	if name != "" {
		name += "-"
	}
	name += a
	prvHex := hex.EncodeToString(prv)

	err := ioutil.WriteFile(path.Join(utils.Keys, name), []byte(prvHex), 0600)
	ifExit(err)
	fmt.Println(name)
}

func cliInstall(c *cli.Context) {
	if len(c.Args()) == 0 {
		ifExit(fmt.Errorf("Please provide a path to the dapp to install"))
	}
	dappPath := c.Args()[0]
	dappName := path.Base(dappPath)
	if len(c.Args()) > 1 {
		dappName = c.Args()[1]
	}
	pdxPath := path.Join(dappPath, "contracts")

	r := make([]byte, 8)
	rand.Read(r)
	tmpRoot := path.Join(utils.Scratch, "epm", hex.EncodeToString(r))

	chainType := "thelonious"

	forceName := c.String("force-name")
	name := c.String("name")
	deployConf := c.String("config")
	deployGen := c.String("genesis")
	tempConf := ".config.json"
	editCfg := c.Bool("edit-config")
	diffStorage := c.Bool("diff")
	rpc := c.Bool("rpc")
	// if we provide genesis, dont open editor for genesis
	noEditor := c.IsSet("genesis")

	// install chain
	chainId := deployInstallChain(tmpRoot, deployConf, deployGen, tempConf, chainType, rpc, editCfg, noEditor)

	ifExit(chains.ChangeHead(chainType, chainId))
	logger.Warnf("Checked out chain: %s/%s", chainType, chainId)

	updateRefs(chainType, chainId, forceName, name)

	// deploy pdx
	contractPath := c.String("c")

	chainRoot := chains.ComposeRootMulti("thelonious", chainId, "0")

	// Startup the chain
	logger.Warnf("Starting up chain:", chainRoot)
	var chain epm.Blockchain
	chain = loadChain(c, "thelonious", chainRoot)

	if !c.IsSet("c") {
		// contractPath = defaultContractPath
		contractPath = pdxPath
	}
	var err error
	epm.ContractPath, err = filepath.Abs(contractPath)
	ifExit(err)

	logger.Debugln("Contract root:", epm.ContractPath)

	// clear cache
	err = os.RemoveAll(utils.Epm)
	if err != nil {
		logger.Errorln("Error clearing cache: ", err)
	}
	utils.InitDataDir(utils.Epm)

	// setup EPM object with ChainInterface
	e, err := epm.NewEPM(chain, epm.LogFile)
	ifExit(err)
	e.ReadVars(path.Join(chainRoot, EPMVars))

	// comb directory for package-definition file
	// exits on error
	dir, pkg, test_ := getPkgDefFile(pdxPath)

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

	var rootContract string
	b, err := ioutil.ReadFile(path.Join(chainRoot, EPMVars))
	ifExit(err)
	sp := strings.Split(string(b), "\n")
	for _, s := range sp {
		sp := strings.Split(s, ":")
		name := sp[0]
		val := sp[1]
		if name == "ROOT" {
			rootContract = val
		}
	}
	// TODO: fetch root contract from vars...

	// install dapp into decerver tree
	p := path.Join(utils.Dapps, dappName)
	ifExit(utils.Copy(dappPath, p))

	// update package.json with chainid and root contract
	p = path.Join(p, "package.json")
	b, err = ioutil.ReadFile(p)
	ifExit(err)
	var pkgFile dapps.PackageFile
	var monkData dapps.MonkData
	err = json.Unmarshal(b, &pkgFile)
	ifExit(err)
	deps := pkgFile.ModuleDependencies
	for i, d := range deps {
		if d.Name == "monk" {
			data := d.Data // json.RawMessage
			err := json.Unmarshal(*data, &monkData)
			ifExit(err)
			monkData.ChainId = "0x" + chainId
			monkData.RootContract = rootContract
			b, err := json.Marshal(monkData)
			ifExit(err)
			raw := json.RawMessage(b)
			deps[i].Data = &raw
			break
		}
	}
	pkgFile.ModuleDependencies = deps
	b, err = json.MarshalIndent(pkgFile, "", "\t")
	ifExit(err)
	err = ioutil.WriteFile(p, b, 0600)
	ifExit(err)
}
