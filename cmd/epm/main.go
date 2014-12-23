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

    app.Flags = []cli.Flag{
        
        // which chain
        cli.StringFlag{
            Name: "name",
            Value: "",
            Usage: "set the chain by name",
            EnvVar: "",
        },
        cli.StringFlag{
            Name: "id",
            Value: "",
            Usage: "set the chain by id",
            EnvVar: "",
        },
        cli.StringFlag{
            Name: "type",
            Value: "",
            Usage: "set the chain type (thelonious, genesis, bitcoin, ethereum)",
            EnvVar: "",
        },

        // epm options
        cli.BoolFlag{
            Name: "i",
            Usage: "Run epm in interactive mode",
            EnvVar: "",
        },
        cli.BoolFlag{
            Name: "diff",
            Usage: "Show a diff of all contract storage",
            EnvVar: "",
        },
        cli.BoolFlag{
            Name: "dont-clear",
            Usage: "Stop epm from clearing the epm cache on startup",
            EnvVar: "",
        },
        cli.StringFlag{
            Name: "c",
            Value: defaultContractPath,
            Usage: "set the contract path",
        },
        cli.StringFlag{
            Name: "p",
            Value: ".",
            Usage: "deploy a .pdx file",
        },

        // log
        cli.IntFlag{
            Name: "log",
            Value: 2,
            Usage: "set the log level",
            EnvVar: "EPM_LOG",
        },
        
        // rpc
        cli.BoolFlag{
            Name: "rpc",
            Usage: "run commands over rpc",
            EnvVar: "",
        },
        cli.StringFlag{
            Name: "host",
            Value: ".",
            Usage: "set the rpc host",
        },
        cli.IntFlag{
            Name: "port",
            Value: 5,
            Usage: "set the rpc port",
        },
    }

    app.Commands = []cli.Command{
        {
            Name:      "clean",
            ShortName: "",
            Usage:     "clean epm related directories",
            Action: cliCleanPullUpdate,
        }, 
        {
            Name:      "plop",
            ShortName: "plop",
            Usage:     "epm plop <config | genesis>",
            Action: cliPlop,
        }, 
        {
            Name:      "refs",
            ShortName: "refs",
            Usage:     "display the chain references",
            Action: cliRefs,
        }, 
        {
            Name:      "head",
            ShortName: "head",
            Usage:     "display the current working chain",
            Action: cliHead,
        }, 
        {
            Name:      "init",
            ShortName: "init",
            Usage:     "initialize the epm tree in ~/.decerver",
            Action: cliInit,
        }, 
        {
            Name:      "fetch",
            ShortName: "fetch",
            Usage:     "asssemble a chain from dapp info",
            Action: cliFetch,
        }, 
        {
            Name:      "deploy",
            ShortName: "deploy",
            Usage:     "deploy a chain",
            Action: cliDeploy,
            Flags:  []cli.Flag{
                cli.BoolFlag{
                    Name: "install, i",
                    Usage: "install the chain following deploy", 
                },
                cli.BoolFlag{
                    Name: "checkout, o",
                    Usage: "checkout the chain into head",
                },
                cli.StringFlag{
                    Name: "config, c",
                    Usage: "specify config file",
                },
                cli.StringFlag{
                    Name: "genesis, g",
                    Usage: "specify genesis file",
                },
            },
        }, 
        {
            Name:      "install",
            ShortName: "install",
            Usage:     "install a chain into the decerver tree",
            Action: cliInstall,
        }, 
        {
            Name:      "checkout",
            ShortName: "checkout",
            Usage:     "change the current working chain",
            Action: cliCheckout,
        }, 
        {
            Name:      "add-ref",
            ShortName: "add-ref",
            Usage:     "add a new reference to a chain id",
            Action: cliAddRef,
        }, 
        {
            Name:      "run",
            ShortName: "run",
            Usage:     "run a chain by reference or id",
            Action: cliRun,
        }, 
        {
            Name:      "run-dapp",
            ShortName: "run-dapp",
            Usage:     "run a chain by dapp name",
            Action: cliRunDapp,
        }, 
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
