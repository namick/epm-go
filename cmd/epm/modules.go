package main

import (
	"github.com/eris-ltd/epm-go/chains"
	"github.com/eris-ltd/epm-go/epm"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/codegangsta/cli"

	// modules
	"github.com/eris-ltd/decerver-interfaces/glue/eth"
	"github.com/eris-ltd/decerver-interfaces/glue/genblock"
	"github.com/eris-ltd/decerver-interfaces/glue/monkrpc"
	"github.com/eris-ltd/thelonious/monk"
)

func newChain(c *cli.Context) epm.Blockchain {
	rpc := c.Bool("rpc")
	switch c.String("type") {
	case "thel", "thelonious", "monk":
		if rpc {
			return monkrpc.NewMonkRpcModule()
		} else {
			return monk.NewMonk(nil)
		}
	case "btc", "bitcoin":
		if rpc {
			log.Fatal("Bitcoin rpc not implemented yet")
		} else {
			log.Fatal("Bitcoin not implemented yet")
		}
	case "eth", "ethereum":
		if rpc {
			log.Fatal("Eth rpc not implemented yet")
		} else {
			//	return NewEthModule(c, chainRoot)
		}
	case "gen", "genesis":
		//return NewGenModule(c, chainRoot)
	}
	return nil

}

// chainroot is a full path to the dir
func loadChain(c *cli.Context, chainRoot string) epm.Blockchain {
	rpc := c.Bool("rpc")
	logger.Debugln("Loading chain ", c.String("type"))
	switch c.String("type") {
	case "thel", "thelonious", "monk":
		if rpc {
			return NewMonkRpcModule(c, chainRoot)
		} else {
			return NewMonkModule(c, chainRoot)
		}
	case "btc", "bitcoin":
		if rpc {
			log.Fatal("Bitcoin rpc not implemented yet")
		} else {
			log.Fatal("Bitcoin not implemented yet")
		}
	case "eth", "ethereum":
		if rpc {
			log.Fatal("Eth rpc not implemented yet")
		} else {
			//	return NewEthModule(c, chainRoot)
		}
	case "gen", "genesis":
		//return NewGenModule(c, chainRoot)
	}
	return nil
}

// TODO: if we are passed a chainRoot but also db is set
//   we should copy from the chainroot to db
// For now, if a chainroot is provided, we use that dir directly

func configureRootDir(c *cli.Context, m epm.Blockchain, chainRoot string) {
	// we need to overwrite the default monk config with our defaults
	root, _ := filepath.Abs(defaultDatabase)
	m.SetProperty("RootDir", root)

	// if the HEAD is set, it overrides the default
	if c, err := chains.GetHead(); err == nil && c != "" {
		root, _ = chains.ResolveChain("thelonious", c, c)
		m.SetProperty("RootDir", root)
		//path.Join(utils.Blockchains, "thelonious", c)
	}

	// if the chainRoot is set, it overwrites the head
	if chainRoot != "" {
		m.SetProperty("RootDir", chainRoot)
	}

	if c.Bool("rpc") {
		m.SetProperty("RootDir", path.Join(m.Property("RootDir").(string), "rpc"))
	}
}

func readConfigFile(c *cli.Context, m epm.Blockchain) {
	// if there's a config file in the root dir, use that
	// else fall back on default or flag
	configFlag := c.String("config")
	s := path.Join(m.Property("RootDir").(string), "config.json")
	if _, err := os.Stat(s); err == nil {
		m.ReadConfig(s)
	} else {
		m.ReadConfig(configFlag)
	}
}

func applyFlags(c *cli.Context, m epm.Blockchain) {
	// then apply cli flags
	setLogLevel(c, m)
	setKeysFile(c, m)
	setGenesisPath(c, m)
	setContractPath(c, m)
	setMining(c, m)
}

func setupModule(c *cli.Context, m epm.Blockchain, chainRoot string) {
	// TODO: kinda bullshit and useless since we set log level at epm
	// m.Config.LogLevel = defaultLogLevel

	configureRootDir(c, m, chainRoot)
	readConfigFile(c, m)
	applyFlags(c, m)

	logger.Infoln("Root directory: ", m.Property("RootDir").(string))

	// initialize and start
	m.Init()
	m.Start()
}

// configure and start an in-process thelonious  node
// all paths should be made absolute
func NewMonkModule(c *cli.Context, chainRoot string) epm.Blockchain {
	m := monk.NewMonk(nil)
	setupModule(c, m, chainRoot)
	return m
}

func NewGenModule(c *cli.Context, chainRoot string) epm.Blockchain {
	m := genblock.NewGenBlockModule(nil)
	setupModule(c, m, chainRoot)
	return m
}

// Rpc module for talking to running thelonious node supporting rpc server
func NewMonkRpcModule(c *cli.Context, chainRoot string) epm.Blockchain {
	m := monkrpc.NewMonkRpcModule()
	setupModule(c, m, chainRoot)
	return m
}

// configure and start an in-process eth node
func NewEthModule(c *cli.Context, chainRoot string) epm.Blockchain {
	m := eth.NewEth(nil)
	setupModule(c, m, chainRoot)
	return m
}
