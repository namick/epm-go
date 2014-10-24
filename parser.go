package epm

import (
    "fmt"
    "bufio"
    "os"
    "strings"
    "regexp"
    "strconv"
    "bytes"
    "encoding/binary"
    "encoding/hex"
    "github.com/project-douglas/lllc-server"
    "github.com/eris-ltd/thelonious/ethutil"
    "github.com/eris-ltd/thelonious/ethtest" // for ordered state map ... 
    "github.com/eris-ltd/thelonious/ethreact" // ...
    "log"
)

var (
    StateDiffOpen = "!{"
    StateDiffClose = "!}"
)

// an EPM Job
type Job struct{
   cmd string
   args []string  // args may contain unparsed math that will be handled by jobs.go
}


// EPM object. maintains list of jobs and a symbols table
type EPM struct{
    eth ChainInterface

    lllcURL string

    jobs []Job
    vars map[string]string

    pkgdef string // latest pkgdef we are parsing
    //states map[string]State// latest ethstates for taking diffs, by name
    Diff bool
    states map[string]ethtest.State //map[string]map[string]string // map from names (of diffs) to states
    diffName  map[int][]string //map job numbers to names of diffs invoked after that job
    diffSched map[int][]int //map job numbers to diff actions (save or diff ie 0 or 1)

    Ch chan ethreact.Event
    
    log string
}

// new empty epm
func NewEPM(eth ChainInterface, log string) *EPM{
    lllcserver.URL = "http://lllc.erisindustries.com/compile"
    e := &EPM{
        eth:  eth,
        jobs: []Job{},
        vars: make(map[string]string),
        log: ".epm-log",
        Diff: false, // off by default
        states: make(map[string]ethtest.State), //map[string]map[string]string),
        diffName: make(map[int][]string),
        diffSched: make(map[int][]int),
        Ch: nil, // set in main when created... make(chan ethreact.Event, 1), // this needs to be generalized
    }
    return e
}

// allowed commands
var CMDS = []string{"deploy", "modify-deploy", "transact", "query", "log", "set", "endow"}

// make sure command is valid
func checkCommand(cmd string) bool{
    r := false
    for _, c := range CMDS{
        if c == cmd{
            r = true
        }
    }
    return r
}

func parseStateDiff(lines *[]string, startLine int) string{
    for n, t := range *lines{
        tt := strings.TrimSpace(t)
        if len(tt) == 0 || tt[0:1] == "#" {
            continue
        }
        t = strings.Split(t, "#")[0]
        t = strings.TrimSpace(t)
        if len(t) == 0{
            continue
        } else if len(t) > 2 && (t[:2] == StateDiffOpen || t[:2] == StateDiffClose){
            // we found a match
            // shave previous lines
                *lines = (*lines)[n:]
            // see if there are other diff statements on this line
            i := strings.IndexAny(t, " \t")
            if i != -1{
                (*lines)[0] = (*lines)[0][i:]
            } else if len(*lines) >= 1{
                *lines = (*lines)[1:]
            }
            return t[2:]
        } else{
            *lines = (*lines)[n:]
            return ""
        }
    }
    return ""
}

func (e EPM) newDiffSched(i int){
    if e.diffSched[i] == nil{
        e.diffSched[i] = []int{}
        e.diffName[i] = []string{}
    }
}

func (e *EPM) parseStateDiffs(lines *[]string, startLine int, diffmap map[string]bool){
    // i is 0 for no jobs
    i := len(e.jobs)
    for {
        name := parseStateDiff(lines, startLine)
        if name != ""{
            e.newDiffSched(i)
            // if we've already seen the name, take diff
            // else, store state
            e.diffName[i] = append(e.diffName[i], name)
            if _, ok := diffmap[name]; ok{
                e.diffSched[i] = append(e.diffSched[i], 1)
            } else{
                e.diffSched[i] = append(e.diffSched[i], 0)
                diffmap[name] = true
            }
            /*if s, ok := e.states[name]; ok{
                fmt.Println("Name of Diff:", name)
                PrettyPrintAcctDiff(StorageDiff(s, e.CurrentState()))
            } else{
                e.states[name] = e.CurrentState()
            }*/
        } else{
            break
        }
    }
}

// peel off the next command and its args
func peelCmd(lines *[]string, startLine int) (*Job, error){
    job := Job{"", []string{}}
    for line, t := range *lines{
        // ignore comments and blank lines
        //fmt.Println("next line:", line, t)
        tt := strings.TrimSpace(t)
        if len(tt) == 0 || tt[0:1] == "#" {
            continue
        }
        t = strings.Split(t, "#")[0]

        // if no cmd yet
        if job.cmd == ""{
            // cmd syntax check
            t = strings.TrimSpace(t)
            l := len(t)
            if t[l-1:] != ":"{
                return nil, fmt.Errorf("Syntax error: missing ':' on line %d", line+startLine)
            }
            cmd := t[:l-1]
            // ensure known cmd
            if !checkCommand(cmd){
                return nil, fmt.Errorf("Invalid command '%s' on line %d", cmd, line+startLine)
            }
            job.cmd = cmd
            continue
        }
        
        // if the line does not begin with white space, we are done 
        if !(t[0:1] == " " || t[0:1] == "\t"){ 
            // peel off lines we've read 
            *lines = (*lines)[line:]
            return &job, nil 
        } 
        t = strings.TrimSpace(t)
        // if there is a diff statement, we are done
        if t[:2] == StateDiffOpen || t[:2] == StateDiffClose{
            *lines = (*lines)[line:]
            return &job, nil
        }
        
        // the line is args. parse them
        // first, eliminate prefix whitespace/tabs
        args := strings.Split(t, "=>")
        // should be 'arg1 => arg2'
        if len(args) != 2 && len(args) != 3{
            return nil, fmt.Errorf("Syntax error: improper argument formatting on line %d", line+startLine)
        }
        for _, a := range args{
            shaven := strings.TrimSpace(a)
            
            job.args = append(job.args, shaven)
        }
    }
    // only gets here if we finish all the lines
    *lines = nil
    return &job, nil
}

//parse should open a file, read all lines, peel commands into jobs
func (e *EPM) Parse(filename string) error{
    fmt.Println("Parsing", filename)
    // set current file to parse
    e.pkgdef = filename

    // temp dir
    // TODO: move it!
    CheckMakeTmp()

    lines := []string{}
    f, err := os.Open(filename)
    if err != nil{
        return err
    }
    scanner := bufio.NewScanner(f)
    // read in all lines
    for scanner.Scan(){
        t := scanner.Text()
        lines = append(lines, t)
    }

    diffmap := make(map[string]bool)

    l := 0
    startLength := len(lines)
    // check if we need to start diffs before the jobs
    e.parseStateDiffs(&lines, l, diffmap)
    for lines != nil{
        // peel off a job and append
        job, err := peelCmd(&lines, l)
        if err != nil{
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
func (e *EPM) VarSub(args []string) []string{
    r, _ := regexp.Compile(`\{\{(.+?)\}\}`)
    for i, a := range args{
        // if its a known var, replace it
        // else, leave alone
        args[i] = r.ReplaceAllStringFunc(a, func(s string) string{
            k := s[2:len(s)-2] // shave the brackets
            v, ok := e.vars[k]
            if ok{
                return v
            } else{
                return s
            }
        })
    }
    return args
}

func (e *EPM) Vars() map[string]string{
    return e.vars
}

func (e *EPM) Jobs() []Job{
    return e.jobs
}

func (e *EPM) StoreVar(key, val string){
    //fmt.Println("storing var:", key, val)
    if key[:2] == "{{" && key[len(key)-2:] == "}}"{
        key = key[2:len(key)-2]
    }
    e.vars[key] = Coerce2Hex(val)
    //fmt.Println("stored result:", e.vars[key])
}

// takes a simple string of nums/hex and ops
// all strings should have been removed
func tokenize(s string) []string{
    tokens := []string{}
    r_opMatch := regexp.MustCompile(`\+|\-|\*`)
    m_inds:= r_opMatch.FindAllStringSubmatchIndex(s, -1)
    // if no ops, just return the string
    if len(m_inds) == 0{
        return []string{s}
    }
    // for each theres a symbol and hex/num after it
    l := 0
    for i, matchI := range m_inds{
        i0 := matchI[0]
        i1 := matchI[1]
        ss := s[l:i0]
        l = i1
        if len(ss) != 0{
            tokens = append(tokens, ss)
        }
        tokens = append(tokens, s[i0:i1]) 
        if i == len(m_inds)-1{
            tokens = append(tokens, s[i1:])
            break
        }
        //tokens = append(tokens, s[i1:m_inds[i+1][0]])
    }
    return tokens
}

// applies any math within an arg
// splits each arg into tokens first by pulling out strings, then by parsing ops/nums/hex between strings
// finally, run through all the tokens doing the math
func DoMath(args []string) []string{
    margs := []string{} // return
    //fmt.Println("domath:", args)
    r_stringMatch := regexp.MustCompile(`\"(.*?)\"`) //"

    for _, a := range args{
        //fmt.Println("time to tokenize:", a)
        tokens := []string{}
        // find all strings (between "")
        strMatches := r_stringMatch.FindAllStringSubmatchIndex(a, -1)
        // grab the expression before the first string
        if len(strMatches) > 0{
            // loop through every interval between strMatches
            // tokenize, append to tokens
            l := 0
            for j, matchI := range strMatches{
                i0 := matchI[0]
                i1 := matchI[1]
                // get everything before this token
                s := a[l:i0]
                l = i1
                // if s is empty, add the string to tokens, move on
                if len(s) == 0{
                    tokens = append(tokens, a[i0+1:i1-1])
                } else{
                    t := tokenize(s)
                    tokens = append(tokens, t...)
                    tokens = append(tokens, a[i0+1:i1-1])
                }
                // if we're on the last one, get anything that comes after
                if j == len(strMatches)-1{
                    s := a[l:]
                    if len(s) > 0{
                        t := tokenize(s)
                        tokens = append(tokens, t...)
                    }
                }
            }
        } else {
           // just tokenize the args 
           tokens = tokenize(a)
        }
        //fmt.Println("tokens:", tokens)

        // now run through the tokens doin the math
        // initialize the first value
        tokenBigBytes, err := hex.DecodeString(stripHex(Coerce2Hex(tokens[0])))
        if err != nil{
            log.Fatal(err)
        }
        result := ethutil.BigD(tokenBigBytes)
        // start with the second token, and go up in twos (should be odd num of tokens)
        for j := 0; j < (len(tokens)-1)/2; j ++{
            op := tokens[2*j+1]
            n := tokens[2*j+2]
            nBigBytes, err := hex.DecodeString(stripHex(Coerce2Hex(n)))
            if err != nil{
                log.Fatal(err)
            }
            tokenBigInt := ethutil.BigD(nBigBytes)
            switch (op){
                case "+":
                    result.Add(result, tokenBigInt)
                case "-":
                    result.Sub(result, tokenBigInt)
                case "*":
                    result.Mul(result, tokenBigInt)
            }
        }
        // TODO: deal with 32-byte overflow
        resultHex := "0x"+hex.EncodeToString(result.Bytes())
        //fmt.Println("resultHex:", resultHex)
        margs = append(margs, resultHex)
    }
    return margs
}

// keeps N bytes of the conversion
func NumberToBytes(num interface{}, N int) []byte {
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, num)
    if err != nil {
        fmt.Println("NumberToBytes failed:", err)
    }
    //fmt.Println("btyes!", buf.Bytes())
    if buf.Len() > N{
        return buf.Bytes()[buf.Len()-N:]
    }
    return buf.Bytes()
}

// s can be string, hex, or int.
// returns properly formatted 32byte hex value
func Coerce2Hex(s string) string{
    //fmt.Println("coercing to hex:", s)
    // is int?
    i, err := strconv.Atoi(s)
    if err == nil{
        return "0x"+hex.EncodeToString(NumberToBytes(int32(i), i/256+1))
    }
    // is already prefixed hex?
    if len(s) > 1 && s[:2] == "0x"{
        if len(s) % 2 == 0{
            return s
        }
        return "0x0"+s[2:]
    }
    // is unprefixed hex?
    if len(s) > 32{
        return "0x"+s
    }
    pad := s + strings.Repeat("\x00", (32-len(s)))
    ret := "0x"+hex.EncodeToString([]byte(pad))
    //fmt.Println("result:", ret)
    return ret
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

// split line and trim space
func parseLine(line string) []string{
    line = strings.TrimSpace(line)
    line = strings.TrimRight(line, ";")

    args := strings.Split(line, ";")
    for i, a := range args{
        args[i] = strings.TrimSpace(a)
    }
    return args
}
