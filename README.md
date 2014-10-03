epm-go
======

Ethereum Package Manager, written in `go`

Allows one to specify a suite of contracts to be deployed and setup from a simple `.package-definition` file.

Interface with blockchain is either in-process through `eth-go-mods/ethtest` or else RPC.

Formatting
----------
Ethereum input data and storage deals strictly in 32-byte segments or words, most conveniently represented as 64 hex characters. When representing data, strings are right padded while ints/hex are left padded.

EPM accepts integers, strings, and explicitly hexidecimal strings (ie. "0x45"). If your string is strictly hex characters but missing a "0x", it will be treated as a normal string. Addresses should be prefixed with "0x" whenever possible.

Values stored as EPM variables will be immediately converted to the proper hex representation. That is, if you store "dog", you will find it later as "0x646f67000000..."

Testing
-------
`go test` can be used to test the parser, or when run in `tests` to test the commands. To test a deployment suite, write a txt file consisting of query params (addr, storage) and the expected result. A fourth parameter can be included for storing the result as a variable for later queries
