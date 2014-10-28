package epm

import (
    "os"
    "bufio"
    "fmt"
)

func (e *EPM) Repl(){
    fmt.Println("#######################################################")
    fmt.Println("##  Welcome to the interactive EPM shell ##############")
    fmt.Println("#######################################################")

    reader := bufio.NewReader(os.Stdin)
    for {
        lines := []string{}
        fmt.Print(">>")
        for {
            text, _ := reader.ReadString('\n')   
            if text == "\n"{
                break
            }
            lines = append(lines, text)
        }
        // check lines for special syntax things

        // else parse for normal cmds
        err := e.parse(lines)
        if err != nil{
            fmt.Println("!>> Parse error:", err)
            continue
        }
        // epm execute jobs
        e.ExecuteJobs()
        // wait for a block
        e.WaitForBlock()
        /*
        if test_{
            results, err := e.Test(path.Join(dir, pkg+"."+TestExt))
            if err != nil{
                fmt.Println(err)
                fmt.Println("failed tests:", results.FailedTests)
            }
        }*/
        // remove jobs for next run
        e.jobs = []Job{}
    }
}
