package main

import (
	"fmt"
	"github.com/eris-ltd/thelonious/monk"
	"github.com/eris-ltd/thelonious/monkdoug"
	"os"
	"path/filepath"
    "github.com/codegangsta/cli"
)

func setLogLevel(c *cli.Context, config *int, loglevel int) {
	if c.IsSet("log") {
		*config = loglevel
	}
}

func setKeysFile(c *cli.Context, config *string, keyfile string) {
	var err error
	if c.IsSet("k") {
		//if keyfile != defaultKeys {
		*config, err = filepath.Abs(keyfile)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}

func setGenesisPath(c *cli.Context, config *string, genfile string) {
	var err error
	if c.IsSet("genesis") {
		//if *config != defaultGenesis && genfile != "" {
		*config, err = filepath.Abs(genfile)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}

func setContractPath(c *cli.Context, config *string, cpath string) {
	var err error
	if c.IsSet("c") {
		//if cpath != defaultContractPath {
		*config, err = filepath.Abs(cpath)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
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

func setDifficulty(c *cli.Context, config *int, d int) {
	*config = defaultDifficulty
	if c.IsSet("difficulty") {
		*config = d
	}
}

// TODO: handle properly (deployed already vs not...)
func setGenesis(c *cli.Context, noGenDoug bool, difficulty int, m *monk.MonkModule) {
	// Handle genesis config
	g := monkdoug.LoadGenesis(m.Config.GenesisConfig)
	if noGenDoug {
		g.NoGenDoug = true
		logger.Infoln("No gendoug")
	}
	setDifficulty(c, &(g.Difficulty), difficulty)
	g.Consensus = "constant"

	m.SetGenesis(g)
}
