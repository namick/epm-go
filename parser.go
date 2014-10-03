package epm

import (
    "fmt"
    "bufio"
    "os"
    "strings"
    "regexp"
    "strconv"
    "github.com/eris-ltd/eth-go-mods/ethutil"
)

// an EPM Job
type Job struct{
   cmd string
   args []string 
}

// EPM object. maintains list of jobs and a symbols table
type EPM struct{
    eth ChainInterface

    jobs []Job
    vars map[string]string
    
    log string
}

// new empty epm
func NewEPM(eth ChainInterface) *EPM{
    return &EPM{
        eth:  eth,
        jobs: []Job{},
        vars: make(map[string]string),
        log: ".epm-log",
    }
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

func shaveWhitespace(t string) string{
    // shave whitespace from front
    for ; t[0:1] == " " || t[0:1] == "\t"; t = t[1:]{
    }
    // shave whitespace from back...
    l := len(t)
    for ; t[l-1:] == " "; t = t[:l-1]{
    }
    return t
}

// peel off the next command and its args
func peelCmd(lines *[]string, startLine int) (*Job, error){
    job := Job{"", []string{}}
    for line, t := range *lines{
        // ignore comments and blank lines
        //fmt.Println("next line:", line, t)
        if len(t) == 0 || t[0:1] == "#" {
            continue
        }
        // if no cmd yet
        if job.cmd == ""{
            // cmd syntax check
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
        // if the line does not begin with white space, we're done
        if !(t[0:1] == " " || t[0:1] == "\t"){
            // peel off lines we've read
            *lines = (*lines)[line:]
            return &job, nil 
        } 
        
        // the line is args. parse them
        // first, eliminate prefix whitespace/tabs
        t = shaveWhitespace(t)

        args := strings.Split(t, "=>")
        // should be 'arg1 => arg2'
        if len(args) != 2 && len(args) != 3{
            return nil, fmt.Errorf("Syntax error: improper argument formatting on line %d", line+startLine)
        }
        for _, a := range args{
            shaven := shaveWhitespace(a)
            job.args = append(job.args, shaven)
        }
    }
    // only gets here if we finish all the lines
    *lines = nil
    return &job, nil
}

//parse should open a file, read all lines, peel commands into jobs
func (e *EPM) Parse(filename string) error{
    lines := []string{}
    f, _ := os.Open(filename)
    scanner := bufio.NewScanner(f)
    // read in all lines
    for scanner.Scan(){
        t := scanner.Text()
        lines = append(lines, t)
    }

    l := 0
    startLength := len(lines)
    for lines != nil{
        job, err := peelCmd(&lines, l)
        if err != nil{
            return err
        }
        e.jobs = append(e.jobs, *job)
        l = startLength - len(lines)
    }
    return nil
}

// replaces any {{varname}} args with the variable value
func (e *EPM) VarSub(args []string) []string{
    r, _ := regexp.Compile(`\{\{(.+?)\}\}`)
    for i, a := range args{
        // if it already exists, replace it
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
    if key[:2] == "{{" && key[len(key)-2:] == "}}"{
        key = key[2:len(key)-2]
    }
    e.vars[key] = addHex(val)
}

// s can be string, hex, or int.
// returns properly formatted 32byte hex value
func Coerce2Hex(s string) string{
    _, err := strconv.Atoi(s)
    if err == nil{
        pad := strings.Repeat("\x00", (32-len(s)))+s
        return "0x"+ethutil.Bytes2Hex([]byte(pad))
    }
    if len(s) > 1 && s[:2] == "0x"{
        return s
    }
    pad := s + strings.Repeat("\x00", (32-len(s)))
    return "0x"+ethutil.Bytes2Hex([]byte(pad))
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

