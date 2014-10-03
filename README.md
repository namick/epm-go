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


Deploy and Test
---------------
```
func main(){
    // Startup the EthChain (ala `eth-go-mods/ethtest/ethereum.go`)
    // or use one you've already got
    eth := NewEthNode()
    // Create ChainInterface instance (hides details of rpc vs. in-process from epm)
    // Here we use in-process (D)
    ethD := epm.NewEthD(eth)
    // setup EPM object with ChainInterface
    e := epm.NewEPM(ethD)
    // epm parse the package definition file
    err := e.Parse(path.Join(epm.TestPath, "test_parse.epm"))
    if err != nil{
        fmt.Println(err)
        os.Exit(0)
    }
    // epm execute jobs
    e.ExecuteJobs()
    // wait for a block to be mined
    ch := make(chan ethreact.Event, 1)
    eth.Ethereum.Reactor().Subscribe("newBlock", ch)
    _ =<- ch
    // test the results against the check file
    e.Test(path.Join(epm.TestPath, "test_parse.epm-check"))
}
```
