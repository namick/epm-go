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
	app.Action = cliDeployPdx
	app.EnableBashCompletion = true

	// TODO: global flags only work on global command!
	app.Flags = []cli.Flag{

		// which chain
		nameFlag,
		idFlag,
		typeFlag,

		// epm options
		interactiveFlag,
		diffFlag,
		dontClearFlag,
		contractPathFlag,
		pdxPathFlag,

		// log
		logLevelFlag,

		// rpc
		rpcFlag,
		rpcHostFlag,
		rpcPortFlag,
	}

	app.Commands = []cli.Command{
		cleanCmd,
		plopCmd,
		refsCmd,
		headCmd,
		initCmd,
		fetchCmd,
		deployCmd,
		installCmd,
		checkoutCmd,
		addRefCmd,
		runCmd,
		runDappCmd,
		configCmd,
		commandCmd,
	}

	utils.InitLogging(path.Join(utils.Logs, "epm"), "", 2, "")

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

	// fail if `epm -init` has not been run
	// TODO: put this everywhere it needs to be...
	//ifExit(checkInit())
}

func run(app *cli.App) {
	defer func() {
		if r := recover(); r != nil {
			trace := make([]byte, 1048)
			count := runtime.Stack(trace, true)
			fmt.Printf("Stack of %d bytes: %s", count, trace)
		}
	}()

	app.Run(os.Args)
}
