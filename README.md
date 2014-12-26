[![Stories in Ready](https://badge.waffle.io/eris-ltd/deCerver.png?label=ready&title=Ready)](https://waffle.io/eris-ltd/deCerver)[![GoDoc](https://godoc.org/github.com/epm-go?status.png)](https://godoc.org/github.com/eris-ltd/epm-go)

Eris Package Manager: The Smart Contract Package Manager
======

Eris Package Manager, written in `go`. EPM makes it easy to spin up blockchains and to deploy suites of contracts or transactions on them.

At its core is a domain specifc language for specifying contract suite deploy sequences, and a git-like interface for managing multiple blockchains.

For tutorials, see the [website](https://epm.io)

epm-go uses the same spec as the [ruby original](https://github.com/project-douglas/epm), but this is subject to change in an upcoming version. If you have input, please make an issue.

Blockchains
-----------

EPM aims to be chain agnostic, by using `module` wrappers satisfying a [blockchain interface](https://github.com/eris-ltd/decerver-interfaces/blob/master/modules/modules.go#L49), built for compatibility with the eris `decerver` ecosystem. 
While theoretically any chain can be supported (provided it satisfies the interface), there is currently support for 

- `thelonious` (in-process and rpc), 
- `ethereum` (in-process), 
- `genesisblock` (for deployments of `thelonious` genesis blocks), 
- `bitcoin` (through `blockchain.info` api wrapper).

We will continue to add support and functionality as time admits.
If you would like epm to be able to work with your blockchain or software, submit a pull-request to `eris-ltd/decerver-interfaces` 
with a wrapper for your chain in `decerver-interfaces/glue` that satisfies the `Blockchain` interface, 
as defined in `epm-go/epm/epm.go`. See the other wrappers in `decerver-interfaces/glue` for examples and inspiration.

Install
--------

1. [Install go](https://golang.org/doc/install)
2. `go get github.com/eris-ltd/epm-go`
3. `cd $GOPATH/src/github.com/eris-ltd/epm-go/cmd/epm`
4. `go get -d .`
5. `go install`

Formatting
----------
Ethereum input data and storage deals strictly in 32-byte segments or words, most conveniently represented as 64 hex characters. 
When representing data, strings are right padded while ints/hex are left padded. 

*IMPORTANT*: contracts deployed with `epm` by default use our version of the LLL compiler, 
which in addition to adding some opcodes, changes strings to also be left padded. 
I repeat, both strings and ints/hex are left padded by default. The reason for this was it simplified the `eris-std-lib`.
Stay tuned for improvement :)

EPM accepts integers, strings, and explicitly hexidecimal strings (ie. "0x45"). 
If your string is strictly hex characters but missing a "0x", it will still be treated as a normal string, so add the "0x".
Addresses should be prefixed with "0x" whenever possible. Integers in base-10 will be handled, hopefully ok.

Values stored as EPM variables will be immediately converted to the proper hex representation. 
That is, if you store "dog", you will find it later as `0x0000000000000000000000000000000000000000000000000000646f67`.

Testing
-------
`go test` can be used to test the parser, or when run in `cmd/tests/` to test the commands. 
To test a deployment suite, write a `.pdt` file with the same name as the `.pdx`, where each line consists of query params (address, storage) and the expected result. 
A fourth parameter can be included for storing the result as a variable for later queries. 
You can test this by running `go run main.go` in `cmd/tests/`. 
See [here](`https://github.com/eris-ltd/eris-std-lib/blob/master/DTT/tests/c3d.pdt`) for examples.

Directory
--------
As part of the larger suite of Eris libraries centered on the `deCerver`, epm works out of the core directory in `~/.decerver/blockchains`. 
A `HEAD` file tracks the currently active chain and a `refs` directory allows chains to be named. 
Otherwise, chains are specified by their `chainId` (signed hash of the genesis block).

Command Line Interface
----------------------
To install the command line tool, cd into `epm-go/cmd/epm/` and hit `go install`. 
You can get the dependencies with `go get -d`.
Assuming your `go bin` is on your path, the cli is accessible as `epm`. 
Simply running that will look for a `.pdx` file in your current directory, deploy it, and run the tests if there are any (in a `.pdt`) file.

Commands:
- `epm init`
    - Initialize the decerver directory tree and default configuration files
- `epm deploy`
    - Deploy a genesis block from a genesis.json file. 
    - The block is saved in a hidden temp folder in the working directory. 
    - It can be installed into the `~/.decerver` with `epm install`, or all at once with `epm deploy -install`. 
    - Specify a particular `config.json` and `genesis.json` with the `-config` and `-genesis` flags, respectively.
    - Otherwise, it will use the default `config.json` and `genesis.json`
    - Deploy will automatically open vim for you to edit config files as you deem fit, before saving them to the working directory
- `epm install`
    - Install a chain into the decerver tree.
    - Specify a particular `config.json` and `genesis.json` with the `-config` and `-genesis` flags, respectively.
    - Otherwise, it will use the `config.json` and `genesis.json` present in the working directory (from calling `deploy`)
    - Trying to install before deploy will fail.
- `epm checkout <chainId/name>`
    - Checkout a chain, making it the current working chain. 
    - It will be written to the top of the `~/.decerver/blockchains/HEAD` file. 
    - Use `epm deploy -checkout` to checkout a chain immediately after deployment, but before installing.
    - Or run `epm deploy -install -checkout` to kill all birds with one stone.
- `epm fetch <dappname>`
    - Deploy and install a chain from a package.json and genesis.json in a dapp repository
    - Easiest way to sync with a dapp specific chain using just the package.json and genesis.json
- `epm run <dappname>`
    - Run a chain by dapp name. Expects the chain to have been installed (probably by `epm -fetch`)
- `epm plop <genesis | config>`
    - Plop the default genesis.json or config.json (respectively) into the current working directory
- `epm [clean | pull | update]`
    - Clean epm related dirs, pull updates to the source code, and re-install the software, respectively.
- `epm add-ref <chainId> <name>`
    - Create a new named reference to a chainId
    - Note you can avoid this by using, for example, `epm deploy -install -name <name>`, to name the chain during installation
- `epm [refs | head]`
    - Display the available references, or the current head, respectively.

Flags
- `-no-gendoug` can be added to force simplified protocols without a genesis doug
- `-config` to specify a chain config file
- `-genesis` to specify a genesis.json config file
- `-name` to attempt to select a chain by name
- `-chainId` to attempt to select a chain by Id
- `-type` to specify the protocol type (eg `bitcoin`, `eth`, `thel`, `genblock`, etc.)
- `-i` to boot into interactive epm mode
- `-diff` to display storage diffs, as specified by wrapping the commands to be diffed in `!{<diffname>` and `!}<diffname>`
- `-dont-clear` to stop epm from clearing its smart contract cache every time (primarily used to handle modify deploys)
- `-c` to specify the contract root path
- `-p` to specify the `.pdx` file
- `-k` to specify a `.keys` file
- `-db` to set the root directory
- `-log` to set the log level
- `-rpc` to talk to chains using rpc
- `-host` and `-port` to set the host and port for rpc communication

To enable tab completion on the cli, add the following to your `~/.bashrc`: `PROG=epm source $GOPATH/src/github.com/codegangsta/cli/autocomplete/bash_autocomplete`

