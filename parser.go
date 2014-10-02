package epm

import (
    "fmt"
    "bufio"
    "os"
    "strings"
    "regexp"
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
        if len(args) != 2 {
            return nil, fmt.Errorf("Syntax error: improper argument formatting on line %d", line+startLine)
        }
        a0 := shaveWhitespace(args[0])
        a1 := shaveWhitespace(args[1])
        job.args = append(job.args, a0, a1)
        
    }
    // only gets here if we finish all the lines
    *lines = nil
    return &job, nil
}

//parse should open a file, read all lines, peel commands into jobs
func (e *EPM) Parse(filename string) error{
    lines := []string{}
    f, _ := os.Open("hi.txt")
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

// job switch
func (e *EPM) ExecuteJob(job Job){
    fmt.Println(job)
    e.VarsSub(&job) // substitute vars 
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
    fmt.Println("vars:", e.vars)
}

// replaces any {{varname}} args with the variable value
func (e *EPM) VarsSub(job *Job){
    r, _ := regexp.Compile(`\{\{(.+?)\}\}`)
    fmt.Println("vars subbing")
    fmt.Println("vars map:", e.vars)
    fmt.Println("args", job.args)
    for i, a := range job.args{
        //l := len(a)
        fmt.Println(a)
        // if it already exists, replace it
        // else, leave alone
        job.args[i] = r.ReplaceAllStringFunc(a, func(s string) string{
            k := s[2:len(s)-2] // shave the brackets
            v, ok := e.vars[k]
            fmt.Println("in replace all s/k/v", s, k, v)
            if ok{
                return v
            } else{
                return s
            }
        })
    }
    fmt.Println("args after sub", job.args)
}

func (e *EPM) Vars() map[string]string{
    return e.vars
}
