package main

import (
	"github.com/codegangsta/cli"
)

var (
	cleanCmd = cli.Command{
		Name:      "clean",
		ShortName: "",
		Usage:     "clean epm related directories",
		Action:    cliCleanPullUpdate,
	}

	plopCmd = cli.Command{
		Name:      "plop",
		ShortName: "plop",
		Usage:     "epm plop <config | genesis>",
		Action:    cliPlop,
	}

	refsCmd = cli.Command{
		Name:      "refs",
		ShortName: "refs",
		Usage:     "display the chain references",
		Action:    cliRefs,
	}

	headCmd = cli.Command{
		Name:      "head",
		ShortName: "head",
		Usage:     "display the current working chain",
		Action:    cliHead,
	}

	initCmd = cli.Command{
		Name:      "init",
		ShortName: "init",
		Usage:     "initialize the epm tree in ~/.decerver",
		Action:    cliInit,
	}

	fetchCmd = cli.Command{
		Name:      "fetch",
		ShortName: "fetch",
		Usage:     "asssemble a chain from dapp info",
		Action:    cliFetch,
	}

	deployCmd = cli.Command{
		Name:      "deploy",
		ShortName: "deploy",
		Usage:     "deploy a chain",
		Action:    cliDeploy,
		Flags: []cli.Flag{
			deployInstallFlag,
			deployCheckoutFlag,
			deployConfigFlag,
			deployGenesisFlag,
			nameFlag,
			typeFlag,
		},
	}

	installCmd = cli.Command{
		Name:      "install",
		ShortName: "install",
		Usage:     "install a chain into the decerver tree",
		Action:    cliInstall,
		Flags: []cli.Flag{
			installCheckoutFlag,
			nameFlag,
		},
	}

	checkoutCmd = cli.Command{
		Name:      "checkout",
		ShortName: "checkout",
		Usage:     "change the current working chain",
		Action:    cliCheckout,
	}

	addRefCmd = cli.Command{
		Name:      "add-ref",
		ShortName: "add-ref",
		Usage:     "add a new reference to a chain id",
		Action:    cliAddRef,
	}

	runCmd = cli.Command{
		Name:      "run",
		ShortName: "run",
		Usage:     "run a chain by reference or id",
		Action:    cliRun,
		Flags: []cli.Flag{
			mineFlag,
			typeFlag,
		},
	}

	runDappCmd = cli.Command{
		Name:      "run-dapp",
		ShortName: "run-dapp",
		Usage:     "run a chain by dapp name",
		Action:    cliRunDapp,
		Flags: []cli.Flag{
			mineFlag,
		},
	}
)
