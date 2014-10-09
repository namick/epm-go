package epm

import (
    "os"
    "os/user"
    "io/ioutil"
    "fmt"
    "strings"
    "path"
    "path/filepath"
    "bufio"
    //"time"
    "github.com/project-douglas/lllc-server"
    "github.com/eris-ltd/eth-go-mods/ethutil"
    "github.com/eris-ltd/eth-go-mods/ethcrypto"
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
    for _, j := range e.jobs{
        fmt.Println("job!", j.cmd, j.args)
        e.ExecuteJob(j)
       // time.Sleep(time.Second) // this was not necessary for epm but was when called from CI. not sure why :(
        // otherwise, tx reactors get blocked;
    }
}

type TestResults struct{
    Tests []string
    Errors []string // go can't marshal/unmarshal errors

    FailedTests []int
    Failed int

    Err string // if we suffered a non-epm-test error

    PkgDefFile string
    PkgTestFile string
}

func (t *TestResults) String() string{
    result := ""

    result += fmt.Sprintf("PkgDefFile: %s\\n", t.PkgDefFile)
    result += fmt.Sprintf("PkgTestFile: %s\\n", t.PkgTestFile)

    if t.Err != ""{
        result += fmt.Sprintf("Fail due to error: %s", t.Err)
        return result
    }

    if t.Failed > 0{
        for _, testN := range t.FailedTests{
            result += fmt.Sprintf("Test %d failed.\\n\\tQuery: %s\\n\\tError: %s", testN, t.Tests[testN], t.Errors[testN])
            if result[len(result)-1:] != "\n"{
                result += "\\n"
            }
        }
        return result
    }
    result += "\\nAll Tests Passed"
    return result
}

func (e *EPM) Test(filename string) (*TestResults, error){
    lines := []string{}
    f, _ := os.Open(filename)
    scanner := bufio.NewScanner(f)
    for scanner.Scan(){
        t := scanner.Text()
        lines = append(lines, t)
    }
    if len(lines) == 0{
        return nil, fmt.Errorf("No tests to run...")
    }

    results := TestResults{
        Tests: lines,
        Errors: []string{},
        FailedTests: []int{},
        Failed: 0,
        Err: "",
        PkgDefFile: e.pkgdef, 
        PkgTestFile: filename,
    }
    
    for i, line := range lines{
        fmt.Println("vars:", e.Vars())
        err := e.ExecuteTest(line, i)
        if err != nil{
            results.Errors = append(results.Errors, err.Error())
        } else{
            results.Errors = append(results.Errors, "")
        }

        if err != nil{
            results.Failed += 1
            results.FailedTests = append(results.FailedTests, i)
            fmt.Println(err)
        }
    }
    var err error
    if results.Failed == 0{
        err = nil
        fmt.Println("passed all tests")
    } else {
        err = fmt.Errorf("failed %d/%d tests", results.Failed, len(lines))
    }
    return &results, err
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
    expected := Coerce2Hex(args[2])
    //expected := args[2]

    // retrieve the value
    val, _ := e.eth.Get("get", []string{addHex(addr), addHex(storage)})
    val = addHex(val)
    //val, _ := e.eth.Get("get", []string{addr, storage})

    if stripHex(expected) != stripHex(val){
        return fmt.Errorf("Test %d failed. Got: %s, expected %s", i, val, expected)
    }

    // store the value
    if len(args) == 4{
        e.StoreVar(args[3], val)
    }
    return nil
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
   
    var p string 
    // compile contract
    if filepath.IsAbs(contract){
        p = contract
    } else {
        p = path.Join(ContractPath, contract)
    }
    fmt.Println("path", p)
    // this needs a better solution ...
    lllcserver.URL = "http://162.218.65.211:9999/compile"
    b, err := lllcserver.Compile(p)
    if err != nil{
        fmt.Println("error compiling!", err)
         return
    }
    // deploy contract
    addr, _ := e.eth.Push("create", []string{"0x"+ethutil.Bytes2Hex(b)})
    // save contract address
    e.StoreVar(key, addr)
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

    for len(args) > 0 {
        sub := args[0]
        rep := args[1]

        lll = strings.Replace(lll, sub, rep, -1)
        args = args[2:]
    }
    
    newPath := path.Join(EPMDIR, ".tmp", ethutil.Bytes2Hex(ethcrypto.Sha3Bin([]byte(lll)))+".lll")
    err = ioutil.WriteFile(newPath, []byte(lll), 0644)
    if err != nil{
        fmt.Println("could not write file", newPath, err)
        os.Exit(0)
    }
    return newPath
}


func CheckMakeTmp(){
    _, err := os.Stat(path.Join(EPMDIR))
    if err != nil{
       err := os.MkdirAll(path.Join(EPMDIR, ".tmp"), 0777)  //wtf!
       if err != nil{
            fmt.Println("Could not make directory. Exiting", err)
            os.Exit(0)
       }
    }
}
