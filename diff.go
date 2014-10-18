package epm

import (
    "fmt"
)


func (e *EPM) PrevState() map[string]map[string]string{
    return e.states["begin"]
}

func (e *EPM) CurrentState() map[string]map[string]string{
    if e.eth == nil{
        return nil
    }
    return e.eth.State()
}

func StorageDiff(pre, post map[string]map[string]string) map[string]map[string]string{
    diff := make(map[string]map[string]string)
    // for each account in pre, compare all elements. remove accounts from post
    for addr, acct := range pre{
        diff[addr] = make(map[string]string)
        acct2 := post[addr]
        // for each storage in the pre acct, check for diff in 2. delete from 2
        for k, v := range acct{
            v2, ok := acct2[k]
            if !ok{
                diff[addr][k] = "0x"
            } else if v2 != v{
                diff[addr][k] = v2
                delete(acct2, k) // delete this entry from post account
            } else if v2 == v{
                delete(acct2, k)
            }
        }
        // what's left are new entries
        for k, v := range acct2{
            diff[addr][k] = v
        }
        delete(post, addr) // delete this address from post accounts
   }
   // whats left are new accts
   for addr, acct := range post{
       diff[addr] = make(map[string]string)
       for k, v := range acct{
            diff[addr][k] = v
       }

   }
   return diff
}

func PrettyPrintAcctDiff(dif map[string]map[string]string) string{
    result := ""
    for addr, acct := range dif{
        if len(acct) == 0{
            continue
        }
        result += addr + ":\n"
        for store, val := range acct{
            result += "\t"+store+": "+val+"\n"
        }
    }
    return result
}

func PrintDiff(name string, pre, post map[string]map[string]string) {
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

