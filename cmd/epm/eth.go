package main

import (
	"flag"
	"fmt"
	"github.com/eris-ltd/epm-go"
	"github.com/eris-ltd/thelonious/monk"
	"github.com/eris-ltd/thelonious/monkdoug"
	"os"
	"path/filepath"
)

// configure and start an in-process eth node
// all paths should be made absolute
func NewEthModule() epm.Blockchain {
	// empty ethchain object
	// note this will load `eth-config.json` into Config if it exists
	m := monk.NewMonk(nil)

	// we need to overwrite the default monk config with our defaults
	m.Config.RootDir, _ = filepath.Abs(defaultDatabase)
	m.Config.LogLevel = defaultLogLevel
	// then try to read local config file to overwrite defaults
	// (if it doesnt exist, it will be saved)
	m.ReadConfig("eth-config.json")
	m.Config.Mining = defaultMining

	// then apply cli flags

	// compute a map of the flags that have been set
	// if set, overwrite default/config-file
	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})
	var err error
	if setFlags["db"] {
		m.Config.RootDir, err = filepath.Abs(*database)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
	if setFlags["log"] {
		m.Config.LogLevel = *logLevel
	}
	if setFlags["mine"] {
		m.Config.Mining = *mining
	}

	if *keys != defaultKeys {
		m.Config.KeyFile, err = filepath.Abs(*keys)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
	if *genesis != defaultGenesis {
		m.Config.GenesisConfig, err = filepath.Abs(*genesis)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		m.Config.ContractPath, err = filepath.Abs(*contractPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}

	// Handle genesis config
	g := monkdoug.LoadGenesis(m.Config.GenesisConfig)
	if *noGenDoug {
		g.NoGenDoug = true
		fmt.Println("no gendoug!")
	}
	g.Difficulty = defaultDifficulty
	if setFlags["dif"] {
		g.Difficulty = *difficulty
	}

	g.Consensus = "constant"

	m.SetGenesis(g)

	// set LLL path
	epm.LLLURL = m.Config.LLLPath

	// initialize and start
	m.Init()
	m.Start()
	return m
}
