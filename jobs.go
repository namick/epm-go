package epm

import (
    "os"
    "io/ioutil"
    "fmt"
    "strings"
    "path"
    "github.com/project-douglas/lllc-server"
    "github.com/eris-ltd/eth-go-mods/ethutil"
    "github.com/eris-ltd/eth-go-mods/ethcrypto"
)

var GOPATH = os.Getenv("GOPATH")
var ContractPath = path.Join(GOPATH, "src", "github.com", "eris-ltd", "eris")

func (e *EPM) ExecuteJobs(){
    for _, j := range e.jobs{
        e.ExecuteJob(j)
    }
}

func (e *EPM) Deploy(args []string){
    contract := args[0]
    key := args[1]
    
    // compile contract
    b, err := lllcserver.CompileLLLWrapper(path.Join(ContractPath, contract))
    if err != nil{
    }

    addr, _ := e.eth.Push("create", []string{ethutil.Bytes2Hex(b)})

    //TODO: make sure addr is hex

    // assign contract addr to key
    e.vars[key] = addr
}

func (e *EPM) ModifyDeploy(args []string){
    contract := args[0]
    key := args[1]
    args = args[2:]
    fmt.Println("modifyinG")
    newName := Modify(path.Join(ContractPath, contract), args) 

    e.Deploy([]string{newName, key})
}

func (e *EPM) Transact(args []string){
    e.eth.Push("tx", args)
}

func (e *EPM) Query(args []string){

}

func (e *EPM) Log(args []string){

}

func (e *EPM) Set(args []string){

}

func (e *EPM) Endow(args []string){

}

// apply substitution/replace pairs from args to contract
// save in temp file
func Modify(contract string, args []string) string{
    fmt.Println("contract:", []byte(contract))
    b, err := ioutil.ReadFile(contract)
    if err != nil{
        fmt.Println("could not open file", contract)
        fmt.Println(err)
        os.Exit(0)
    }

    lll := string(b)

    for len(args) > 0 {
        sub := args[0]
        rep := args[1]
        lll = strings.Replace(lll, sub, rep, -1)
        args = args[2:]
    }
    
    newPath := path.Join(".tmp", ethutil.Bytes2Hex(ethcrypto.Sha3Bin([]byte(contract))))
    err = ioutil.WriteFile(newPath, []byte(lll), 0644)
    if err != nil{
        fmt.Println("could not write file", newPath, err)
        os.Exit(0)
    }
    return newPath
}

