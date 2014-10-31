package epm

import (
    "fmt"
    "github.com/eris-ltd/decerver-interfaces/core"
)


func (e *EPM) CurrentState() core.State{ //map[string]string{
    if e.eth == nil{
        return core.State{}
    }
    return e.eth.State()
}

func (e *EPM) checkTakeStateDiff(i int){
    if _, ok := e.diffSched[i]; !ok{
       return 
    }
    e.WaitForBlock()
    scheds := e.diffSched[i]
    names := e.diffName[i]
    for j, sched := range scheds{
        name := names[j]
        if sched == 0{
            // store state
            e.states[name] = e.CurrentState()
        } else{
            // take diff
            e.WaitForBlock()
            PrintDiff(name, e.states[name], e.CurrentState())
        }
    }
}

func StorageDiff(pre, post core.State) core.State{ //map[string]string) map[string]map[string]string{
    diff := core.State{make(map[string]core.Storage), []string{}}
    // for each account in post, compare all elements. 
    for _, addr := range post.Order{
        acct := post.State[addr]
        diff.State[addr] = core.Storage{make(map[string]interface{}), []string{}}
        diff.Order = append(diff.Order, addr)
        acct2 := pre.State[addr]
        // for each storage in the post acct, check for diff in 2. 
        for _, k := range acct.Order{
            v := acct.Storage[k]
            v2, ok := acct2.Storage[k]
            //fmt.Println(v, v2)
            // if its not in the pre-state or its different, add to diff
            if !ok || v2 != v{
                diff.State[addr].Storage[k] = v
                st := diff.State[addr]
                st.Order = append(diff.State[addr].Order, k)
                diff.State[addr] = st
            }
        }
   }
   return diff
}

func PrettyPrintAcctDiff(dif core.State) string{ //map[string]string) string{
    result := ""
    for _, addr := range dif.Order{
        acct := dif.State[addr]
        if len(acct.Order) == 0{
            continue
        }
        result += addr + ":\n"
        for _, store := range acct.Order{
            v := acct.Storage[store]
            val := v.(string)
            result += "\t"+store+": "+val+"\n"
        }
    }
    return result
}

func PrintDiff(name string, pre, post core.State){  //map[string]string) {
    /*
    fmt.Println("pre")
    fmt.Println(PrettyPrintAcctDiff(pre))
    fmt.Println("\n\n")
    fmt.Println("post")
    fmt.Println(PrettyPrintAcctDiff(post))
    fmt.Println("\n\n")
    */
    fmt.Println("diff:", name)
    diff := StorageDiff(pre, post)
    fmt.Println(PrettyPrintAcctDiff(diff))
}

