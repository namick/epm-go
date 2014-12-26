package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/eris-ltd/epm-go/epm"
	"os"
	"path/filepath"
)

func setLogLevel(c *cli.Context, m epm.Blockchain) {
	logLevel := c.Int("log")
	if c.IsSet("log") {
		m.SetProperty("LogLevel", logLevel)
	}
}

func setKeysFile(c *cli.Context, m epm.Blockchain) {
	keys := c.String("keys")
	if c.IsSet("k") {
		//if keyfile != defaultKeys {
		keysAbs, err := filepath.Abs(keys)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		m.SetProperty("KeyFile", keysAbs)
	}
}

func setGenesisPath(c *cli.Context, m epm.Blockchain) {
	genesis := c.String("genesis")
	if c.IsSet("genesis") {
		//if *config != defaultGenesis && genfile != "" {
		genAbs, err := filepath.Abs(genesis)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		m.SetProperty("GenesisPath", genAbs)
	}
}

func setContractPath(c *cli.Context, m epm.Blockchain) {
	contractPath := c.String("c")
	if c.IsSet("c") {
		//if cpath != defaultContractPath {
		cPathAbs, err := filepath.Abs(contractPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		m.SetProperty("ContractPath", cPathAbs)
	}
}

func setMining(c *cli.Context, m epm.Blockchain) {
	tomine := c.Bool("mine")
	if c.IsSet("mine") {
		m.SetProperty("Mining", tomine)
	}
}

func setDb(c *cli.Context, config *string, dbpath string) {
	var err error
	if c.IsSet("db") {
		*config, err = filepath.Abs(dbpath)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}

var (
	nameFlag = cli.StringFlag{
		Name:   "name",
		Value:  "",
		Usage:  "specify a ref name",
		EnvVar: "",
	}

	idFlag = cli.StringFlag{
		Name:   "id",
		Value:  "",
		Usage:  "set the chain by id",
		EnvVar: "",
	}

	typeFlag = cli.StringFlag{
		Name:   "type",
		Value:  "thelonious",
		Usage:  "set the chain type (thelonious, genesis, bitcoin, ethereum)",
		EnvVar: "",
	}

	interactiveFlag = cli.BoolFlag{
		Name:   "i",
		Usage:  "Run epm in interactive mode",
		EnvVar: "",
	}

	diffFlag = cli.BoolFlag{
		Name:   "diff",
		Usage:  "Show a diff of all contract storage",
		EnvVar: "",
	}

	dontClearFlag = cli.BoolFlag{
		Name:   "dont-clear",
		Usage:  "Stop epm from clearing the epm cache on startup",
		EnvVar: "",
	}

	contractPathFlag = cli.StringFlag{
		Name:  "c",
		Value: defaultContractPath,
		Usage: "set the contract path",
	}

	pdxPathFlag = cli.StringFlag{
		Name:  "p",
		Value: ".",
		Usage: "deploy a .pdx file",
	}

	logLevelFlag = cli.IntFlag{
		Name:   "log",
		Value:  2,
		Usage:  "set the log level",
		EnvVar: "EPM_LOG",
	}

	mineFlag = cli.BoolFlag{
		Name:  "mine, commit",
		Usage: "commit blocks",
	}

	rpcFlag = cli.BoolFlag{
		Name:   "rpc",
		Usage:  "run commands over rpc",
		EnvVar: "",
	}

	rpcHostFlag = cli.StringFlag{
		Name:  "host",
		Value: ".",
		Usage: "set the rpc host",
	}

	rpcPortFlag = cli.IntFlag{
		Name:  "port",
		Value: 5,
		Usage: "set the rpc port",
	}

	deployInstallFlag = cli.BoolFlag{
		Name:  "install, i",
		Usage: "install the chain following deploy",
	}
	deployCheckoutFlag = cli.BoolFlag{
		Name:  "checkout, o",
		Usage: "checkout the chain into head",
	}
	deployConfigFlag = cli.StringFlag{
		Name:  "config, c",
		Usage: "specify config file",
	}
	deployGenesisFlag = cli.StringFlag{
		Name:  "genesis, g",
		Usage: "specify genesis file",
	}

	installCheckoutFlag = cli.BoolFlag{
		Name:  "checkout, o, c",
		Usage: "checkout the chain into head",
	}

	globalFlag = cli.BoolFlag{
		Name:  "global",
		Usage: "edit the global default config",
	}
)
