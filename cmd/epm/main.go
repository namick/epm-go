package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/eris-ltd/epm-go/utils"
	"github.com/eris-ltd/thelonious/monklog"
	"os"
	"path"
	"runtime"
)

var (
	GoPath = os.Getenv("GOPATH")

	logger *monklog.Logger = monklog.NewLogger("EPM-CLI")

	// location for blockchain database before install
	ROOT = ".decerver-local"

	// epm extensions
	PkgExt  = "pdx"
	TestExt = "pdt"

	defaultContractPath = "." //path.Join(utils.ErisLtd, "eris-std-lib")
	defaultPackagePath  = "."
	defaultGenesis      = ""
	defaultKeys         = ""
	defaultDatabase     = ".chain"
	defaultLogLevel     = 2
	defaultDiffStorage  = false
)

func main() {
	logger.Infoln("DEFAULT CONTRACT PATH: ", defaultContractPath)

	app := cli.NewApp()
	app.Name = "epm"
	app.Usage = ""
	app.Version = "0.1.0"
	app.Author = "Ethan Buchman"
	app.Email = "ethan@erisindustries.com"
	//	app.EnableBashCompletion = true

	app.Before = before
	app.Flags = []cli.Flag{
		// which chain
		chainFlag,

		// log
		logLevelFlag,

		// rpc
		rpcFlag,
		rpcHostFlag,
		rpcPortFlag,
		rpcLocalFlag,
	}

	app.Commands = []cli.Command{
		cleanCmd,
		plopCmd,
		refsCmd,
		headCmd,
		initCmd,
		fetchCmd,
		newCmd,
		checkoutCmd,
		addRefCmd,
		runCmd,
		runDappCmd,
		configCmd,
		commandCmd,
		removeCmd,
		deployCmd,
		consoleCmd,
	}

	// clean, update, or install
	// exit
	/*
		if *clean || *pull || *update {
			cleanPullUpdate()
			exit(nil)
		}
	*/
	run(app)

	monklog.Flush()
}

func before(c *cli.Context) error {
	utils.InitLogging(path.Join(utils.Logs, "epm"), "", c.Int("log"), "")
	if _, err := os.Stat(utils.Decerver); err != nil {
		exit(fmt.Errorf("Could not find decerver tree. Did you run `epm init`?"))
	}
	return nil
}

// so we can catch panics
func run(app *cli.App) {
	defer func() {
		if r := recover(); r != nil {
			trace := make([]byte, 2048)
			count := runtime.Stack(trace, true)
			fmt.Printf("Panic: ", r)
			fmt.Printf("Stack of %d bytes: %s", count, trace)
		}
	}()

	app.Run(os.Args)
}
