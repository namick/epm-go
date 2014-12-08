[![Stories in Ready](https://badge.waffle.io/eris-ltd/deCerver.png?label=ready&title=Ready)](https://waffle.io/eris-ltd/deCerver)

epm-go
======
Eris Package Manager, written in `go`.

Allows one to specify a suite of contracts to be deployed and setup from a simple `.pdx` (package definition)package file. Also provides `git-like` interface for working with chains.

Interfaces with blockchains are through `decerver-interfaces` blockchain modules. There is currently support for `thelonious` (in-process and rpc), `ethereum` (in-process), and `genesisblock` (for deployments of `thelonious` genesis blocks), and basic support for `bitcoin` (through `blockchain.info`).

Formatting
----------
Ethereum input data and storage deals strictly in 32-byte segments or words, most conveniently represented as 64 hex characters. When representing data, strings are right padded while ints/hex are left padded.

EPM accepts integers, strings, and explicitly hexidecimal strings (ie. "0x45"). If your string is strictly hex characters but missing a "0x", it will be treated as a normal string. Addresses should be prefixed with "0x" whenever possible. Integers in base-10 will be handled, hopefully ok.

Values stored as EPM variables will be immediately converted to the proper hex representation. That is, if you store "dog", you will find it later as "0x646f67000000..."

Testing
-------
`go test` can be used to test the parser, or when run in `cmd/tests/` to test the commands. To test a deployment suite, write a txt file consisting of query params (addr, storage) and the expected result. A fourth parameter can be included for storing the result as a variable for later queries. You can test this by running `go run main.go` in `cmd/tests/`. See the test files in `cmd/tests/definitions/` for examples.

Directory
--------
As part of the larger suite of Eris libraries centered on the `deCerver`, epm works out of the core directory in `~/.decerver/blockchains`. A `HEAD` file tracks the currently active chain and a `refs` directory allows chains to be named. Otherwise, chains are specified by their `chainId` (signed hash of the genesis block).

Command Line Interface
----------------------
To install the command line tool, cd into `epm-go/cmd/epm/` and hit `go install`. Assuming your `go bin` is on your path, the cli is accessible as `epm`. Simply running that will look for a `.pdx` file in your current directory, deploy it, and run the tests if there are any (in a `.pdt`) file.

Commands:
- `epm -init <path>`
    - Initialize the decerver directory tree 
    - Deposits default config files for a thelonious blockchain and a genesis deployment in <path>.
- `epm -deploy`
    - Deploy a genesis block from a genesis.json file. 
    - The block is saved in a hidden temp folder in the working directory. 
    - It can be installed into the `~/.decerver` with `epm -install`, or all at once with `epm -deploy -install`. 
    - Specify a particular `config.json` and `genesis.json` with the `-config` and `-genesis` flags, respectively.
- `epm -checkout <chainId/name>`
    - Checkout a chain, making it the current working chain. 
    - It will be written to the top of the `~/.decerver/blockchains/HEAD` file. 
    - Use `epm -deploy -checkout` to checkout a chain immediately after deployment, but before installing.
    - Or run `epm -deploy -install -checkout` to kill all birds with one stone.
- `epm -[clean | pull | update]`
    - Clean epm related dirs, pull updates to the source code, and re-install the software, respectively.
- `epm -add-ref <chainId> -name <name>`
    - Create a new named reference to a chainId
- `epm -[refs | head]`
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
- `-c` to specify the contract root path
- `-p` to specify the `.pdx` file
- `-k` to specify a `.keys` file
- `-db` to set the root directory
- `-log` to set the log level
- `-rpc` to talk to chains using rpc
- `-host` and `-port` to set the host and port for rpc communication


