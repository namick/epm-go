package main

import (
	"github.com/eris-ltd/epm-go"
	"path/filepath"

	// modules
	"github.com/eris-ltd/decerver-interfaces/glue/eth"
	"github.com/eris-ltd/decerver-interfaces/glue/genblock"
	"github.com/eris-ltd/decerver-interfaces/glue/monkrpc"
	"github.com/eris-ltd/thelonious/monk"
)

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
	// then try to read local config file to overwrite defaults
	// (if it doesnt exist, it will be saved)
	m.ReadConfig(*config)

	// then apply cli flags
	setFlags := specifiedFlags()
	setDb(setFlags, &(m.Config.RootDir), *database)
	setLogLevel(setFlags, &(m.Config.LogLevel), *logLevel)
	setKeysFile(setFlags, &(m.Config.KeyFile), *keys)
	setGenesisPath(setFlags, &(m.Config.GenesisConfig), *genesis)
	setContractPath(setFlags, &(m.Config.ContractPath), *contractPath)

	if chainRoot != "" {
		m.Config.RootDir = chainRoot
	}

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
