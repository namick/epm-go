package epm

import (
)

/*
    Implementations of ChainInterface
        - ethD: in process access to an eth-go-mods EthChain
        - ethRPC: json-rpc calls to an eth-node
*/

// interface to ethereum obj for interacting with chain
// could be rpc, eth node, etc.
type ChainInterface interface{
    Get(query string, args []string) (string, error)
    Push(kind string, args []string)   (string, error)
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
}


func NewEthD(ethChain EthChain) ChainInterface{
    return &EthD{chain:ethChain}
}

func (e *EthD) Get(query string, args []string) (string, error){
    addr := args[0]
    storage := args[1]
    ret :=  e.chain.GetStorageAt(addr, storage)
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

// TODO: RPC .
