package epm

import (
    "os"
    "io/ioutil"
    "fmt"
    "strings"
    "path"
    "bufio"
    "github.com/project-douglas/lllc-server"
    "github.com/eris-ltd/eth-go-mods/ethutil"
    "github.com/eris-ltd/eth-go-mods/ethcrypto"
)

var GOPATH = os.Getenv("GOPATH")
var ContractPath = path.Join(GOPATH, "src", "github.com", "project-douglas", "epm-go", "tests", "contracts")
var TestPath = path.Join(GOPATH, "src", "github.com", "project-douglas", "epm-go", "tests", "definitions")

func (e *EPM) ExecuteJobs(){
    for _, j := range e.jobs{
        fmt.Println("job!", j.cmd, j.args)
        e.ExecuteJob(j)
    }
}

func (e *EPM) Test(filename string){
    lines := []string{}
    f, _ := os.Open(filename)
    scanner := bufio.NewScanner(f)
    for scanner.Scan(){
        t := scanner.Text()
        lines = append(lines, t)
    }
    
    failed := 0
    for i, line := range lines{
        err := e.ExecuteTest(line, i)
        if err != nil{
            failed += 1
            fmt.Println(err)
        }
    }
    if failed == 0{
        fmt.Println("passed all tests")
    } else {

        fmt.Println("failed %d/%d", failed, len(lines))
    }
}

func (e *EPM) ExecuteTest(line string, i int) error{
    args := strings.Split(line, " ")
    args = e.VarSub(args)
    fmt.Println("test!", i)
    fmt.Println(args)
    if len(args) < 3 || len(args) > 4{
        return fmt.Errorf("invalid number of args for test on line %d", i)
    }
    // a test is 'addr storage expected'
    // if there's a fourth, its the variable name to store the result under
    addr := args[0]
    storage := args[1]
    expected := args[2]

    // retrieve the value
    val, _ := e.eth.Get("get", []string{addHex(addr), addHex(storage)})
    //val, _ := e.eth.Get("get", []string{addr, storage})

    if stripHex(expected) != stripHex(val){
        fmt.Println("bytes:", []byte(val), []byte(expected))
        return fmt.Errorf("Test %d failed. Got: %s, expected %s", i, val, expected)
    }

    // store the value
    if len(args) == 4{
        e.StoreVar(args[3], val)
    }
    fmt.Println(e.vars)
    return nil
}

func (e *EPM) StoreVar(key, val string){
    if key[:2] == "{{" && key[len(key)-2:] == "}}"{
        key = key[2:len(key)-2]
    }
    e.vars[key] = val

}

func addHex(s string) string{
    if len(s) < 2{
        return "0x"+s
    }

    if s[:2] != "0x"{
        return "0x"+s
    }
    
    return s
}

func stripHex(s string) string{
    if len(s) > 1{
        if s[:2] == "0x"{
            return s[2:]
        }
    }
    return s
}



// job switch
func (e *EPM) ExecuteJob(job Job){
    job.args = e.VarSub(job.args) // substitute vars 
    switch(job.cmd){
        case "deploy":
            e.Deploy(job.args)
        case "modify-deploy":
            e.ModifyDeploy(job.args)
        case "transact":
            e.Transact(job.args)
        case "query":
            e.Query(job.args)
        case "log":
            e.Log(job.args)
        case "set":
            e.Set(job.args)
        case "endow":
            e.Endow(job.args)
    }
}

/*
    The following are the "jobs" functions that EPM knows
    Interaction with BlockChain is strictly through Get() and Push() methods of ChainInterface
    Hides details of in-process vs. rpc
*/

func (e *EPM) Deploy(args []string){
    fmt.Println("deploy!")
    contract := args[0]
    key := args[1]
    
    // compile contract
    p := path.Join(ContractPath, contract)
    fmt.Println("path", p)
    b, err := lllcserver.CompileLLLWrapper(p)
    if err != nil{
        fmt.Println("error compiling!", err)
         return
    }

    addr, _ := e.eth.Push("create", []string{"0x"+ethutil.Bytes2Hex(b)})

    // assign contract addr to key (strip the {{}})
    e.vars[key[2:len(key)-2]] = "0x"+addr
}

func (e *EPM) ModifyDeploy(args []string){
    fmt.Println("modify-deploy!")
    contract := args[0]
    key := args[1]
    args = args[2:]
    newName := Modify(path.Join(ContractPath, contract), args) 

    e.Deploy([]string{newName, key})
}

func (e *EPM) Transact(args []string){
    to := args[0:1]
    dataS := args[1]
    data := strings.Split(dataS, " ")
    e.eth.Push("tx", append(to, data...))
}

func (e *EPM) Query(args []string){
    addr := args[0]
    storage := args[1]
    varName := args[2]

    fmt.Println("running query:", addr, storage)

    v, _ := e.eth.Get("get", []string{addr, storage})
    e.StoreVar(varName, v)
}

func (e *EPM) Log(args []string){
    k := args[0]
    v := args[1]

    f, err := os.OpenFile(e.log, os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    if _, err = f.WriteString(fmt.Sprintf("%s : %s", k, v)); err != nil {
        panic(err)
    }
}

func (e *EPM) Set(args []string){
    k := args[0]
    v := args[1]
    e.vars[k[2:len(k)-2]] = v
}

func (e *EPM) Endow(args []string){
    addr := args[0]
    value := args[1]
    e.eth.Push("endow", []string{addr, value})
}

// apply substitution/replace pairs from args to contract
// save in temp file
func Modify(contract string, args []string) string{
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

    newPath := path.Join(".tmp", ethutil.Bytes2Hex(ethcrypto.Sha3Bin([]byte(lll)))+".lll")
    err = ioutil.WriteFile(path.Join(ContractPath, newPath), []byte(lll), 0644)
    if err != nil{
        fmt.Println("could not write file", newPath, err)
        os.Exit(0)
    }
    return newPath
}

