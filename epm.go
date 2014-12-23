package epm

import (
	"bufio"
	"github.com/eris-ltd/decerver-interfaces/events"
	"github.com/eris-ltd/decerver-interfaces/modules"
	"github.com/eris-ltd/thelonious/monklog"
	"github.com/project-douglas/lllc-server"
	"os"
	"regexp"
)

var logger *monklog.Logger = monklog.NewLogger("EPM")

var (
	StateDiffOpen  = "!{"
	StateDiffClose = "!}"
	LLLURL         = "http://lllc.erisindustries.com/compile"
)

// an EPM Job
type Job struct {
	cmd  string
	args []string // args may contain unparsed math that will be handled by jobs.go
}

type KeyManager interface {
	ActiveAddress() string
	Address(n int) (string, error)
	SetAddress(addr string) error
	SetAddressN(n int) error
	NewAddress(set bool) string
	AddressCount() int
}

type DecerverModule interface{
    Init() error
    Start() error

    ReadConfig(string)
    SetProperty(string, interface{}) error
    Property(string) interface{}
}

type Blockchain interface {
	KeyManager
    DecerverModule
	WorldState() *modules.WorldState
	State() *modules.State
	Storage(target string) *modules.Storage
	Account(target string) *modules.Account
	StorageAt(target, storage string) string
	BlockCount() int
	LatestBlock() string
	Block(hash string) *modules.Block
	IsScript(target string) bool
	Tx(addr, amt string) (string, error)
	Msg(addr string, data []string) (string, error)
	Script(file, lang string) (string, error)
	Subscribe(name, event, target string) chan events.Event
	UnSubscribe(name string)
	Commit()
	AutoCommit(toggle bool)
	IsAutocommit() bool

	Shutdown() error
	WaitForShutdown()
}

// EPM object. Maintains list of jobs and a symbols table
type EPM struct {
	chain Blockchain

	lllcURL string

	jobs []Job
	vars map[string]string

	pkgdef string
	Diff   bool
	states map[string]modules.State

	//map job numbers to names of diffs invoked after that job
	diffName map[int][]string
	//map job numbers to diff actions (save or diff ie 0 or 1)
	diffSched map[int][]int

	log string
}

// new empty epm
func NewEPM(chain Blockchain, log string) (*EPM, error) {
	lllcserver.URL = LLLURL
	//logger.Infoln("url: ", LLLURL)
	e := &EPM{
		chain:     chain,
		jobs:      []Job{},
		vars:      make(map[string]string),
		log:       ".epm-log",
		Diff:      false, // off by default
		states:    make(map[string]modules.State),
		diffName:  make(map[int][]string),
		diffSched: make(map[int][]int),
	}
	// temp dir
	err := CopyContractPath()
	return e, err
}

// allowed commands
var CMDS = []string{"deploy", "modify-deploy", "transact", "query", "log", "set", "endow", "test", "epm"}

func (e EPM) newDiffSched(i int) {
	if e.diffSched[i] == nil {
		e.diffSched[i] = []int{}
		e.diffName[i] = []string{}
	}
}

func (e *EPM) parseStateDiffs(lines *[]string, startLine int, diffmap map[string]bool) {
	// i is 0 for no jobs
	i := len(e.jobs)
	for {
		name := parseStateDiff(lines, startLine)
		if name != "" {
			e.newDiffSched(i)
			// if we've already seen the name, take diff
			// else, store state
			e.diffName[i] = append(e.diffName[i], name)
			if _, ok := diffmap[name]; ok {
				e.diffSched[i] = append(e.diffSched[i], 1)
			} else {
				e.diffSched[i] = append(e.diffSched[i], 0)
				diffmap[name] = true
			}
			/*if s, ok := e.states[name]; ok{
			      fmt.Println("Name of Diff:", name)
			      PrettyPrintAcctDiff(StorageDiff(s, e.CurrentState()))
			  } else{
			      e.states[name] = e.CurrentState()
			  }*/
		} else {
			break
		}
	}
}

func (e *EPM) Parse(filename string) error {
	logger.Infoln("Parsing ", filename)
	// set current file to parse
	e.pkgdef = filename

	lines := []string{}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	// read in all lines
	for scanner.Scan() {
		t := scanner.Text()
		lines = append(lines, t)
	}
	return e.parse(lines)
}

// parse should take a list of lines, peel commands into jobs
// lines either come from a file or from iepm
func (e *EPM) parse(lines []string) error {

	diffmap := make(map[string]bool)

	l := 0
	startLength := len(lines)
	// check if we need to start diffs before the jobs
	e.parseStateDiffs(&lines, l, diffmap)
	for lines != nil {
		// peel off a job and append
		job, err := peelCmd(&lines, l)
		if err != nil {
			return err
		}
		e.jobs = append(e.jobs, *job)
		// check if we need to take or diff state after this job
		// if diff is disabled they will not run, but we need to parse them out
		e.parseStateDiffs(&lines, l, diffmap)
		l = startLength - len(lines)
	}
	return nil
}

// replaces any {{varname}} args with the variable value
func (e *EPM) VarSub(args []string) []string {
	r, _ := regexp.Compile(`\{\{(.+?)\}\}`)
	for i, a := range args {
		// if its a known var, replace it
		// else, leave alone
		args[i] = r.ReplaceAllStringFunc(a, func(s string) string {
			k := s[2 : len(s)-2] // shave the brackets
			v, ok := e.vars[k]
			if ok {
				return v
			} else {
				return s
			}
		})
	}
	return args
}

func (e *EPM) Vars() map[string]string {
	return e.vars
}

func (e *EPM) Jobs() []Job {
	return e.jobs
}

func (e *EPM) StoreVar(key, val string) {
	if key[:2] == "{{" && key[len(key)-2:] == "}}" {
		key = key[2 : len(key)-2]
	}
	e.vars[key] = Coerce2Hex(val)
}
