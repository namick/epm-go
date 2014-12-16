package main

import (
	"github.com/eris-ltd/epm-go"
	"log"
	"os"
	"path"
	"path/filepath"

	// modules
	"github.com/eris-ltd/decerver-interfaces/glue/eth"
	"github.com/eris-ltd/decerver-interfaces/glue/genblock"
	"github.com/eris-ltd/decerver-interfaces/glue/monkrpc"
	"github.com/eris-ltd/decerver-interfaces/glue/utils"
	"github.com/eris-ltd/thelonious/monk"
)

func loadChain(chainRoot string) epm.Blockchain {
	logger.Debugln("Loading chain ", *chainType)
	switch *chainType {
	case "thel", "thelonious", "monk":
		if *rpc {
			return NewMonkRpcModule(chainRoot)
		} else {
			return NewMonkModule(chainRoot)
		}
	case "btc", "bitcoin":
		if *rpc {
			log.Fatal("Bitcoin rpc not implemented yet")
		} else {
			log.Fatal("Bitcoin not implemented yet")
		}
	case "eth", "ethereum":
		if *rpc {
			log.Fatal("Eth rpc not implemented yet")
		} else {
			return NewEthModule(chainRoot)
		}
	case "gen", "genesis":
		return NewGenModule(chainRoot)
	}
	return nil
}

// TODO: if we are passed a chainRoot but also db is set
//   we should copy from the chainroot to db
// For now, if a chainroot is provided, we use that dir directly

// configure and start an in-process thelonious  node
// all paths should be made absolute
func NewMonkModule(chainRoot string) epm.Blockchain {
	// empty ethchain object with default config
	m := monk.NewMonk(nil)

	// we need to overwrite the default monk config with our defaults
	m.Config.RootDir, _ = filepath.Abs(defaultDatabase)
	m.Config.LogLevel = defaultLogLevel

	// if the HEAD is set, it overrides the default
	if c, err := utils.GetHead(); err == nil && c != "" {
		m.Config.RootDir, _ = utils.ResolveChain("thelonious", c, c)
		//path.Join(utils.Blockchains, "thelonious", c)
	}

	// if the chainRoot is set, it overwrites the head
	if chainRoot != "" {
		m.Config.RootDir = chainRoot
	}

	// if there's a config file in the root dir, use that
	// else fall back on default or flag
	c := path.Join(m.Config.RootDir, "config.json")
	if _, err := os.Stat(c); err == nil {
		m.ReadConfig(c)
	} else {
		m.ReadConfig(*config)
	}

	// then apply cli flags
	setFlags := specifiedFlags()
	setDb(setFlags, &(m.Config.RootDir), *database)
	setLogLevel(setFlags, &(m.Config.LogLevel), *logLevel)
	setKeysFile(setFlags, &(m.Config.KeyFile), *keys)
	setGenesisPath(setFlags, &(m.Config.GenesisConfig), *genesis)
	setContractPath(setFlags, &(m.Config.ContractPath), *contractPath)

	logger.Infoln("Root directory: ", m.Config.RootDir)
	// load and set GenesisConfig object
	setGenesis(setFlags, m)

	// set LLL path
	epm.LLLURL = m.Config.LLLPath

	// initialize and start
	m.Init()
	m.Start()
	return m
}

// Deploy genesis blocks using EPM
func NewGenModule(chainRoot string) epm.Blockchain {
	// empty ethchaIn object
	// note this will load `eth-config.json` into Config if it exists
	m := genblock.NewGenBlockModule(nil)

	// we need to overwrite the default monk config with our defaults
	m.Config.RootDir, _ = filepath.Abs(defaultDatabase)
	m.Config.LogLevel = defaultLogLevel

	// if the HEAD is set, it overrides the default
	if c, err := utils.GetHead(); err != nil && c != "" {
		m.Config.RootDir, _ = utils.ResolveChain("thelonious", c, "")
	}

	// then try to read local config file to overwrite defaults
	// (if it doesnt exist, it will be saved)
	m.ReadConfig(*config)

	// then apply cli flags
	setFlags := specifiedFlags()
	setDb(setFlags, &(m.Config.RootDir), *database)
	setLogLevel(setFlags, &(m.Config.LogLevel), *logLevel)
	setKeysFile(setFlags, &(m.Config.KeyFile), *keys)
	setContractPath(setFlags, &(m.Config.ContractPath), *contractPath)

	if chainRoot != "" {
		m.Config.RootDir = chainRoot
	}

	// set LLL path
	epm.LLLURL = m.Config.LLLPath

	// initialize and start
	m.Init()
	m.Start()
	return m
}

// Rpc module for talking to running thelonious node supporting rpc server
func NewMonkRpcModule(chainRoot string) epm.Blockchain {
	// empty ethchain object
	// note this will load `eth-config.json` into Config if it exists
	m := monkrpc.NewMonkRpcModule()

	// we need to overwrite the default monk config with our defaults
	m.Config.RootDir, _ = filepath.Abs(defaultDatabase)
	m.Config.LogLevel = defaultLogLevel

	// if the HEAD is set, it overrides the default
	if c, err := utils.GetHead(); err != nil && c != "" {
		m.Config.RootDir, _ = utils.ResolveChain("thelonious", c, "")
	}

	// then try to read local config file to overwrite defaults
	// (if it doesnt exist, it will be saved)
	m.ReadConfig(*config)

	// then apply cli flags
	setFlags := specifiedFlags()
	setDb(setFlags, &(m.Config.RootDir), *database)
	setLogLevel(setFlags, &(m.Config.LogLevel), *logLevel)
	setKeysFile(setFlags, &(m.Config.KeyFile), *keys)
	setContractPath(setFlags, &(m.Config.ContractPath), *contractPath)

	if chainRoot != "" {
		m.Config.RootDir = chainRoot
	}

	// set LLL path
	epm.LLLURL = m.Config.LLLPath

	// initialize and start
	m.Init()
	m.Start()
	return m
}

// configure and start an in-process eth node
func NewEthModule(chainRoot string) epm.Blockchain {
	// empty ethchain object
	m := eth.NewEth(nil)

	// we need to overwrite the default monk config with our defaults
	m.Config.RootDir, _ = filepath.Abs(defaultDatabase)
	m.Config.LogLevel = defaultLogLevel

	// if the HEAD is set, it overrides the default
	if c, err := utils.GetHead(); err != nil && c != "" {
		m.Config.RootDir, _ = utils.ResolveChain("ethereum", c, "")
	}

	// then try to read local config file to overwrite defaults
	// (if it doesnt exist, it will be saved)
	m.ReadConfig(*config)

	// then apply cli flags
	setFlags := specifiedFlags()
	setDb(setFlags, &(m.Config.RootDir), *database)
	setLogLevel(setFlags, &(m.Config.LogLevel), *logLevel)
	setKeysFile(setFlags, &(m.Config.KeyFile), *keys)
	setContractPath(setFlags, &(m.Config.ContractPath), *contractPath)

	if chainRoot != "" {
		m.Config.RootDir = chainRoot
	}

	// set LLL path
	epm.LLLURL = m.Config.LLLPath

	// initialize and start
	m.Init()
	m.Start()
	return m
}
