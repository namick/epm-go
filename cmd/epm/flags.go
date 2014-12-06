package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func specifiedFlags() map[string]bool {
	// compute a map of the flags that have been set
	// for those that are set, we will overwrite default/config-file
	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})
	return setFlags
}

func setLogLevel(flags map[string]bool, config *int, loglevel int) {
	if flags["log"] {
		*config = loglevel
	}
}

func setKeysFile(flags map[string]bool, config *string, keyfile string) {
	var err error
	if keyfile != defaultKeys {
		*config, err = filepath.Abs(keyfile)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}

func setGenesisPath(flags map[string]bool, config *string, genfile string) {
	var err error
	if *config != defaultGenesis && genfile != "" {
		*config, err = filepath.Abs(genfile)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}

func setContractPath(flags map[string]bool, config *string, cpath string) {
	var err error
	if cpath != defaultContractPath {
		*config, err = filepath.Abs(cpath)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}

func setDb(flags map[string]bool, config *string, dbpath string) {
	var err error
	if flags["db"] {
		*config, err = filepath.Abs(dbpath)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}

func setDifficulty(flags map[string]bool, config *int, d int) {
	*config = defaultDifficulty
	if flags["dif"] {
		*config = d
	}
}
