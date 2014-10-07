[![Stories in Ready](https://badge.waffle.io/eris-ltd/deCerver.png?label=ready&title=Ready)](https://waffle.io/eris-ltd/deCerver)

epm-go
======

Ethereum Package Manager, written in `go`

Allows one to specify a suite of contracts to be deployed and setup from a simple `.package-definition` file.

Interface with blockchain is either in-process through `eth-go-mods/ethtest` or else RPC.

Formatting
----------
Ethereum input data and storage deals strictly in 32-byte segments or words, most conveniently represented as 64 hex characters. When representing data, strings are right padded while ints/hex are left padded.

EPM accepts integers, strings, and explicitly hexidecimal strings (ie. "0x45"). If your string is strictly hex characters but missing a "0x", it will be treated as a normal string. Addresses should be prefixed with "0x" whenever possible. Integers in base-10 will be handled, hopefully ok.

Values stored as EPM variables will be immediately converted to the proper hex representation. That is, if you store "dog", you will find it later as "0x646f67000000..."

Testing
-------
`go test` can be used to test the parser, or when run in `cmd/tests/` to test the commands. To test a deployment suite, write a txt file consisting of query params (addr, storage) and the expected result. A fourth parameter can be included for storing the result as a variable for later queries. You can test this by running `go run main.go` in `cmd/tests/`. See the test files in `cmd/tests/definitions/` for examples.

Command Line Interface
----------------------
To install the command line tool, cd into `epm-go/cmd/epm-go/` and hit `go install`. Assuming your go bin is on your path, the cli is accessible as `epm-go`. Simply running that will look for a `.package-definition` file in your current directory, deploy it, and run the tests if there are any (in a `.package-definition-test`) file.

Note by default, epm-go starts a new eth-node in the current directory under `.ethchain`. Delete this directory if you wish to clear state before the next deploy.

Configuration options are as follows:

    General:
        - `epm-go` will look for a .package-definition file in the current directory, and expect all contracts to have paths beginning in the current dir
    Paths:
        - `-c` to set the contract root (ie. the pkg-defn file has contract paths starting from here)
        - `-p` to specify a .pkg-defn file. The corresponding test file is expected to be in the same location directory
    Eth:
        - by default, a fresh eth-instance will be started with no genesis doug. To specify:
            - `-g` to set a genesis.json configuration file
            - `-k` to set a keys.txt file (with one hex-encoded private key per line. The corresponding addresses should appear in genesis.json)
            - `-db` to set the location of an eth levelDB database to use
        - the `-rpc`, `-rpcHost` and `-rpcPort` flags to use rpc. `-rpc` alone will use the defaults, while using one of host/port will choose the default for the other
        - the `-d`, `-host` and `-port` flags specify to pass commands through a deCerver, and to set the host/port 
