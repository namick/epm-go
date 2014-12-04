package epm

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	//"time"
	"crypto/sha256"
	"encoding/hex"
	"github.com/project-douglas/lllc-server"
)

var GOPATH = os.Getenv("GOPATH")

// should be set to the "current" directory if using epm-cli
var ContractPath = path.Join(GOPATH, "src", "github.com", "eris-ltd", "epm-go", "cmd", "tests", "contracts")
var TestPath = path.Join(GOPATH, "src", "github.com", "eris-ltd", "epm-go", "cmd", "tests", "definitions")
var EPMDIR = path.Join(usr(), ".epm-go")

func usr() string {
	u, _ := user.Current()
	return u.HomeDir
}

func (e *EPM) Commit() {
	e.chain.Commit()
}

func (e *EPM) ExecuteJobs() {
	if e.Diff {
		e.checkTakeStateDiff(0)
	}
	// TODO: set gendoug...
	//gendougaddr, _:= e.eth.Get("gendoug", nil)
	//e.StoreVar("GENDOUG", gendougaddr)

	for i, j := range e.jobs {
		//fmt.Println("job!", j.cmd, j.args)
		e.ExecuteJob(j)
		if e.Diff {
			e.checkTakeStateDiff(i + 1)
		}

		// time.Sleep(time.Second) // this was not necessary for epm but was when called from CI. not sure why :(
		// otherwise, tx reactors get blocked;
	}
	if e.Diff {
		e.checkTakeStateDiff(len(e.jobs))
	}
}

// job switch
// args are still raw input from user (but only 2 or 3)
func (e *EPM) ExecuteJob(job Job) {
	job.args = e.VarSub(job.args) // substitute vars
	//fmt.Println("job!", job.cmd, job.args)
	switch job.cmd {
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
	case "test":
		e.chain.Commit()
		err := e.ExecuteTest(job.args[0], 0)
		if err != nil {
			fmt.Println(err)
		}
	case "epm":
		e.EPMx(job.args[0])

	}
	//fmt.Println(e.vars)
}

/*
   The following are the "jobs" functions that EPM knows
   Interaction with BlockChain is strictly through Get() and Push() methods of ChainInterface
   Hides details of in-process vs. rpc
*/

func (e *EPM) EPMx(filename string) {
	// save the old jobs, empty the job list
	oldjobs := e.jobs
	e.jobs = []Job{}

	if err := e.Parse(filename); err != nil {
		fmt.Println("failed to parse pdx file:", filename)
		fmt.Println(err)
		os.Exit(0)
	}

	e.ExecuteJobs()
	e.jobs = oldjobs
}

func (e *EPM) Deploy(args []string) {
	//fmt.Println("deploy!")
	contract := args[0]
	key := args[1]
	contract = strings.Trim(contract, "\"")
	var p string
	// compile contract
	if filepath.IsAbs(contract) {
		p = contract
	} else {
		p = path.Join(ContractPath, contract)
	}
	b, err := lllcserver.Compile(p, false)
	if err != nil {
		fmt.Println("error compiling!", err)
		return
	}
	// deploy contract
	addr, _ := e.chain.Script("0x"+hex.EncodeToString(b), "bytes")
	// save contract address
	e.StoreVar(key, addr)
}

func (e *EPM) ModifyDeploy(args []string) {
	//fmt.Println("modify-deploy!")
	contract := args[0]
	key := args[1]
	args = args[2:]

	contract = strings.Trim(contract, "\"")
	newName := Modify(path.Join(ContractPath, contract), args)
	e.Deploy([]string{newName, key})
}

func (e *EPM) Transact(args []string) {
	to := args[0]
	dataS := args[1]
	data := strings.Split(dataS, " ")
	data = DoMath(data)
	e.chain.Msg(to, data)
}

func (e *EPM) Query(args []string) {
	addr := args[0]
	storage := args[1]
	varName := args[2]

	//fmt.Println("running query:", addr, storage)

	v := e.chain.StorageAt(addr, storage)
	e.StoreVar(varName, v)
	fmt.Printf("\tresult: %s = %s\n", varName, v)
}

func (e *EPM) Log(args []string) {
	k := args[0]
	v := args[1]

	_, err := os.Stat(e.log)
	var f *os.File
	if err != nil {
		f, err = os.Create(e.log)
		if err != nil {
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

func (e *EPM) Set(args []string) {
	k := args[0]
	v := args[1]
	e.StoreVar(k, v)
}

func (e *EPM) Endow(args []string) {
	addr := args[0]
	value := args[1]
	e.chain.Tx(addr, value)
}

// apply substitution/replace pairs from args to contract
// save in temp file
func Modify(contract string, args []string) string {
	b, err := ioutil.ReadFile(contract)
	if err != nil {
		fmt.Println("could not open file", contract)
		fmt.Println(err)
		os.Exit(0)
	}

	lll := string(b)

	fmt.Println("contract", contract)
	fmt.Println("contract path", ContractPath)
	// when we modify a contract, we save it in the .tmp dir in the same relative path as its original root.
	// eg. if ContractPath is ~/ponos and we modify ponos/projects/issue.lll then the modified version will be found in
	//  EPMDIR/.tmp/ponos/projects/somehash.lll
	dirC := path.Dir(contract) // absolute path
	l := len(ContractPath)
	var dir string
	if dirC != ContractPath {
		dir = dirC[l+1:] //this is relative to the contract root (ie. projects/issue.lll)
	} else {
		dir = ""
	}
	root := path.Base(ContractPath) // base of the ContractPath should be the root dir of the contracts
	dir = path.Join(root, dir)      // add in the root (ie. ponos/projects/issue.lll)

	for len(args) > 0 {
		sub := args[0]
		rep := args[1]

		lll = strings.Replace(lll, sub, rep, -1)
		args = args[2:]
	}

	hash := sha256.Sum256([]byte(lll))
	newPath := path.Join(EPMDIR, ".tmp", dir, hex.EncodeToString(hash[:])+".lll")
	err = ioutil.WriteFile(newPath, []byte(lll), 0644)
	if err != nil {
		fmt.Println("could not write file", newPath, err)
		os.Exit(0)
	}
	return newPath
}

func CheckMakeTmp() {
	_, err := os.Stat(path.Join(EPMDIR, ".tmp"))
	if err != nil {
		err := os.MkdirAll(path.Join(EPMDIR, ".tmp"), 0777) //wtf!
		if err != nil {
			fmt.Println("Could not make directory. Exiting", err)
			os.Exit(0)
		}
	}
	// copy the current dir into .tmp. Necessary for finding include files after a modify. :sigh:
	root := path.Base(ContractPath)
	if _, err = os.Stat(path.Join(EPMDIR, ".tmp", root)); err != nil {
		cmd := exec.Command("cp", "-r", ContractPath, path.Join(EPMDIR, ".tmp"))
		err = cmd.Run()
		if err != nil {
			fmt.Println("error copying working dir into tmp:", err)
			os.Exit(0)
		}
	}
}
