package epm

import (
    "os"
    "os/user"
    "os/exec"
    "io/ioutil"
    "fmt"
    "strings"
    "path"
    "path/filepath"
    //"time"
    "github.com/project-douglas/lllc-server"
    "encoding/hex"
    "crypto/sha256"
)

var GOPATH = os.Getenv("GOPATH")
// should be set to the "current" directory if using epm-cli
var ContractPath = path.Join(GOPATH, "src", "github.com", "eris-ltd", "epm-go", "cmd", "tests", "contracts")
var TestPath = path.Join(GOPATH, "src", "github.com", "eris-ltd", "epm-go", "cmd", "tests", "definitions")
var EPMDIR = path.Join(usr(), ".epm-go")

func usr() string{
    u, _ := user.Current()
    return u.HomeDir
}

func (e *EPM) ExecuteJobs(){
    // so we can take a diff after
    e.state = e.eth.State()
    // set gendoug
    gendougaddr, _:= e.eth.Get("gendoug", nil)
    e.StoreVar("GENDOUG", gendougaddr)
    for _, j := range e.jobs{
        fmt.Println("job!", j.cmd, j.args)
        e.ExecuteJob(j)
       // time.Sleep(time.Second) // this was not necessary for epm but was when called from CI. not sure why :(
        // otherwise, tx reactors get blocked;
    }
}
// job switch
// args are still raw input from user (but only 2 or 3)
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
    //fmt.Println("deploy!")
    contract := args[0]
    key := args[1]
    contract = strings.Trim(contract, "\"")
    var p string 
    // compile contract
    if filepath.IsAbs(contract){
        p = contract
    } else {
        p = path.Join(ContractPath, contract)
    }
    b, err := lllcserver.Compile(p)
    if err != nil{
        fmt.Println("error compiling!", err)
         return
    }
    // deploy contract
    addr, _ := e.eth.Push("create", []string{"0x"+hex.EncodeToString(b)})
    // save contract address
    e.StoreVar(key, addr)
}

func (e *EPM) ModifyDeploy(args []string){
    //fmt.Println("modify-deploy!")
    contract := args[0]
    key := args[1]
    args = args[2:]

    contract = strings.Trim(contract, "\"")
    newName := Modify(path.Join(ContractPath, contract), args) 
    e.Deploy([]string{newName, key})
}

func (e *EPM) Transact(args []string){
    to := args[0:1]
    dataS := args[1]
    data := strings.Split(dataS, " ")
    data = DoMath(data)
    e.eth.Push("tx", append(to, data...))
}

func (e *EPM) Query(args []string){
    addr := args[0]
    storage := args[1]
    varName := args[2]

    //fmt.Println("running query:", addr, storage)

    v, _ := e.eth.Get("get", []string{addr, storage})
    e.StoreVar(varName, v)
}

func (e *EPM) Log(args []string){
    k := args[0]
    v := args[1]

    _, err := os.Stat(e.log)
    var f *os.File
    if err != nil{
        f, err = os.Create(e.log)
        if err != nil{
            panic(err)    
        }
    } else {
        f, err = os.OpenFile(e.log, os.O_APPEND|os.O_WRONLY, 0600)
        if err != nil {
            panic(err)
        }
    }
    defer f.Close()

    if _, err = f.WriteString(fmt.Sprintf("%s : %s", k, v)); err != nil {
        panic(err)
    }
}

func (e *EPM) Set(args []string){
    k := args[0]
    v := args[1]
    e.StoreVar(k, v)
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

    // when we modify a contract, we save it in the .tmp dir in the same relative path as its original root.
    // eg. if ContractPath is ~/ponos and we modify ponos/projects/issue.lll then the modified version will be found in
    //  EPMDIR/.tmp/ponos/projects/somehash.lll
    dirC := path.Dir(contract) // absolute path
    l := len(ContractPath)
    dir := dirC[l+1:] //this is relative to the contract root (ie. projects/issue.lll)
    root := path.Base(ContractPath) // base of the ContractPath should be the root dir of the contracts
    dir = path.Join(root, dir) // add in the root (ie. ponos/projects/issue.lll)
    
    for len(args) > 0 {
        sub := args[0]
        rep := args[1]

        lll = strings.Replace(lll, sub, rep, -1)
        args = args[2:]
    }
    
    hash := sha256.Sum256([]byte(lll))
    newPath := path.Join(EPMDIR, ".tmp", dir, hex.EncodeToString(hash[:])+".lll")
    err = ioutil.WriteFile(newPath, []byte(lll), 0644)
    if err != nil{
        fmt.Println("could not write file", newPath, err)
        os.Exit(0)
    }
    return newPath
}


func CheckMakeTmp(){
    _, err := os.Stat(path.Join(EPMDIR, ".tmp"))
    if err != nil{
       err := os.MkdirAll(path.Join(EPMDIR, ".tmp"), 0777)  //wtf!
       if err != nil{
            fmt.Println("Could not make directory. Exiting", err)
            os.Exit(0)
       }
       // copy the current dir into .tmp. Necessary for finding include files after a modify. :sigh:
       cmd := exec.Command("cp", "-r", ContractPath, path.Join(EPMDIR, ".tmp"))
       err = cmd.Run()
       if err != nil{
            fmt.Println("error copying working dir into tmp:", err)
            os.Exit(0)
       }

       }
}
