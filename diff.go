package epm

import (
    "fmt"
    "github.com/eris-ltd/eth-go-mods/ethtest"
)


func (e *EPM) CurrentState() ethtest.State{ //map[string]string{
    if e.eth == nil{
        return ethtest.State{}
    }
    return e.eth.State()
}

func StorageDiff(pre, post ethtest.State) ethtest.State{ //map[string]string) map[string]map[string]string{
    diff := ethtest.State{make(map[string]ethtest.Storage), []string{}}
    // for each account in post, compare all elements. 
    for _, addr := range post.Order{
        acct := post.State[addr]
        diff.State[addr] = ethtest.Storage{make(map[string]string), []string{}}
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

func PrettyPrintAcctDiff(dif ethtest.State) string{ //map[string]string) string{
    result := ""
    for _, addr := range dif.Order{
        acct := dif.State[addr]
        if len(acct.Order) == 0{
            continue
        }
        result += addr + ":\n"
        for _, store := range acct.Order{
            val := acct.Storage[store]
            result += "\t"+store+": "+val+"\n"
        }
    }
    return result
}

func PrintDiff(name string, pre, post ethtest.State){  //map[string]string) {
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

