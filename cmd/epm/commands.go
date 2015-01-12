package main

import (
	"github.com/codegangsta/cli"
)

var (
	cleanCmd = cli.Command{
		Name:   "clean",
		Usage:  "clean epm related directories",
		Action: cliCleanPullUpdate,
	}

	plopCmd = cli.Command{
		Name:   "plop",
		Usage:  "epm plop <config | genesis>",
		Action: cliPlop,
	}

	refsCmd = cli.Command{
		Name:   "refs",
		Usage:  "display the chain references",
		Action: cliRefs,
	}

	headCmd = cli.Command{
		Name:   "head",
		Usage:  "display the current working chain",
		Action: cliHead,
	}

	initCmd = cli.Command{
		Name:   "init",
		Usage:  "initialize the epm tree in ~/.decerver",
		Action: cliInit,
	}

	fetchCmd = cli.Command{
		Name:   "fetch",
		Usage:  "asssemble a chain from dapp info",
		Action: cliFetch,
	}

	newCmd = cli.Command{
		Name:   "new",
		Usage:  "create a new chain and install into the decerver tree",
		Action: cliNew,
		Flags: []cli.Flag{
			newCheckoutFlag,
			newConfigFlag,
			newGenesisFlag,
			nameFlag,
			typeFlag,
		},
	}

	checkoutCmd = cli.Command{
		Name:   "checkout",
		Usage:  "change the current working chain",
		Action: cliCheckout,
	}

	addRefCmd = cli.Command{
		Name:   "add-ref",
		Usage:  "add a new reference to a chain id",
		Action: cliAddRef,
	}

	runCmd = cli.Command{
		Name:   "run",
		Usage:  "run a chain by reference or id",
		Action: cliRun,
		Flags: []cli.Flag{
			mineFlag,
			chainFlag,
		},
	}

	runDappCmd = cli.Command{
		Name:   "run-dapp",
		Usage:  "run a chain by dapp name",
		Action: cliRunDapp,
		Flags: []cli.Flag{
			mineFlag,
		},
	}

	configCmd = cli.Command{
		Name:   "config",
		Usage:  "epm config <config key 1>:<config value 1> <config key 2>:<config value 2> ...",
		Action: cliConfig,
		Flags: []cli.Flag{
			chainFlag,
			typeFlag,
		},
	}

	commandCmd = cli.Command{
		Name:   "cmd",
		Usage:  "epm cmd deploy contract.lll",
		Action: cliCommand,
		Flags: []cli.Flag{
			chainFlag,
			contractPathFlag,
		},
	}

	removeCmd = cli.Command{
		Name:   "rm",
		Usage:  "remove a chain from the global directory",
		Action: cliRemove,
	}

	deployCmd = cli.Command{
		Name:   "deploy",
		Usage:  "deploy a .pdx file onto a chain",
		Action: cliDeploy,
		Flags: []cli.Flag{
			chainFlag,
			diffFlag,
			dontClearFlag,
			contractPathFlag,
		},
	}

	consoleCmd = cli.Command{
		Name:   "console",
		Usage:  "run epm in interactive mode",
		Action: cliConsole,
		Flags: []cli.Flag{
			chainFlag,
			diffFlag,
			dontClearFlag,
			contractPathFlag,
		},
	}
)