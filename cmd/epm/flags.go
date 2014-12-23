package main

import (
	"fmt"
	"github.com/eris-ltd/epm-go/epm"
	"os"
	"path/filepath"
    "github.com/codegangsta/cli"
)

func setLogLevel(c *cli.Context, m epm.Blockchain) {
    logLevel := c.Int("log")
	if c.IsSet("log") {
		m.SetProperty("LogLevel", logLevel)
	}
}

func setKeysFile(c *cli.Context, m epm.Blockchain) {
    keys :=  c.String("keys")
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

func setGenesisPath(c *cli.Context, m epm.Blockchain){
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

func setContractPath(c *cli.Context, m epm.Blockchain){
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
