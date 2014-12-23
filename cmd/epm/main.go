package main

import (
    "path"
    "os"
    "github.com/codegangsta/cli"
	"github.com/eris-ltd/thelonious/monklog"
	"github.com/eris-ltd/epm-go/utils"
)

var (
	GoPath = os.Getenv("GOPATH")

	logger *monklog.Logger = monklog.NewLogger("EPM-CLI")

	// location for blockchain database before install
	ROOT = ".decerver-local"

	// epm extensions
	PkgExt  = "pdx"
	TestExt = "pdt"

	defaultContractPath = "."
	defaultPackagePath  = "."
	defaultGenesis      = ""
	defaultKeys         = ""
	defaultDatabase     = ".chain"
	defaultLogLevel     = 5
	defaultDifficulty   = 14
	defaultMining       = false
	defaultDiffStorage  = false
)

func main() {

    app := cli.NewApp()
    app.Name = "epm"
    app.Usage = ""
    app.Action = cliDeployPdx

    // TODO: global flags only work on global command!
    // TODO: replace these blocks with oneliners by defining them elsewhere
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
    }

	utils.InitLogging(path.Join(utils.Logs, "epm"), "", 5, "")

	// clean, update, or install
	// exit
/*
	if *clean || *pull || *update {
		cleanPullUpdate()
		exit(nil)
	}
*/
    app.Run(os.Args)

	// fail if `epm -init` has not been run
    // TODO: put this everywhere it needs to be...
	//ifExit(checkInit())
}
