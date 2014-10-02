epm-go
======

Ethereum Package Manager, written in `go`

Allows one to specify a suite of contracts to be deployed and setup from a simple `.package-definition` file.

Interface with blockchain is either in-process through `eth-go-mods/ethtest` or else RPC.

Testing
-------
`tests` contains contracts and definition files for testing integration with in-process ethereum.
