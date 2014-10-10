package epm

import (
    "fmt"
)


func (e *EPM) PrevState() map[string]map[string]string{
    return e.state
}

func (e *EPM) CurrentState() map[string]map[string]string{
    return e.eth.State()
}

func StorageDiff(pre, post map[string]map[string]string) map[string]map[string]string{
    diff := make(map[string]map[string]string)
    // for each account in pre, compare all elements. remove accounts from post
    for addr, acct := range pre{
        diff[addr] = make(map[string]string)
        acct2 := post[addr]
        // for each storage in the pre acct, check for diff in 2. delete from 2
        for k, v := range acct{
            v2, ok := acct2[k]
            if !ok{
                diff[addr][k] = "0x"
            } else if v2 != v{
                diff[addr][k] = v2
                delete(acct2, k) // delete this entry from post account
            } else if v2 == v{
                delete(acct2, k)
            }
        }
        // what's left are new entries
        for k, v := range acct2{
            diff[addr][k] = v
        }
        delete(post, addr) // delete this address from post accounts
   }
   // whats left are new accts
   for addr, acct := range post{
       diff[addr] = make(map[string]string)
       for k, v := range acct{
            diff[addr][k] = v
       }

   }
   return diff
}

func PrettyPrintAcctDiff(dif map[string]map[string]string) string{
    result := ""
    for addr, acct := range dif{
        if len(acct) == 0{
            continue
        }
        result += addr + ":\n"
        for store, val := range acct{
            result += "\t"+store+": "+val+"\n"
        }
    }
    return result
}

func PrintDiff(pre, post map[string]map[string]string) {
    fmt.Println("pre")
    fmt.Println(PrettyPrintAcctDiff(pre))
    fmt.Println("\n\n")
    fmt.Println("post")
    fmt.Println(PrettyPrintAcctDiff(post))
    fmt.Println("\n\n")
    fmt.Println("diff")
    diff := StorageDiff(pre, post)
    fmt.Println(PrettyPrintAcctDiff(diff))
}


/*
    Implementations of ChainInterface
        - ethD: in process access to an eth-go-mods EthChain
        - ethRPC: json-rpc calls to an eth-node
*/

// interface to ethereum obj for interacting with chain
// could be rpc, eth node, etc.
type ChainInterface interface{
    Get(query string, args []string) (string, error)
    Push(kind string, args []string) (string, error)
    State() map[string]map[string]string // the full chain state ... this won't scale
}


// for an eth daemon
// implements ChainInterface
// uses an simplified EthChain interface for in-process txs and lookups
type EthD struct{
    chain EthChain
}

type EthChain interface{
    Tx(string, string)
    Msg(string, []string)
    DeployContract(string, string) string
    GetStorageAt(string, string) string
    GetState() map[string]map[string]string
    GenDoug() string
}


func NewEthD(ethChain EthChain) ChainInterface{
    return &EthD{chain:ethChain}
}

func (e *EthD) Get(query string, args []string) (string, error){
    var ret string
    switch(query){
        case "get":
            addr := args[0]
            storage := args[1]
            ret = e.chain.GetStorageAt(addr, storage)
        case "gendoug":
           ret = e.chain.GenDoug()
    }
    return ret, nil
}

func (e *EthD) Push(kind string, args []string) (string, error){
    switch(kind){
        case "create":
            bytecode := args[0]
            hexAddr := e.chain.DeployContract(bytecode, "bytes")
            return hexAddr, nil
        case "tx":
           addr := args[0]
           data := args[1:] 
           e.chain.Msg(addr, data)
           return "", nil
        case "endow":
           addr := args[0]
           amt := args[1] 
           e.chain.Tx(addr, amt)
           return "", nil
    }
    return "", nil
}

func (e *EthD) State() map[string]map[string]string {
   return e.chain.GetState()
}

// TODO: RPC .
