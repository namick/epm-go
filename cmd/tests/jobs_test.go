package main

import (
    "fmt"
    "testing"
    "path"
    "github.com/eris-ltd/epm-go"
    "github.com/eris-ltd/eth-go-mods/ethreact"
)

/*
    For direct coding of hardcoded contracts and test results.   
    See definitions and contracts for context
*/

func TestDeploy(t *testing.T){
    eth := NewEthNode()
    e := epm.NewEPM(epm.NewEthD(eth), ".epm-log-test")
    e.Ch = Subscribe(eth, "newBlock")

    err := e.Parse(path.Join(epm.TestPath, "test_deploy.epm"))
    if err != nil{
        t.Error(err)
    }
    fmt.Println(e.Jobs())
    e.ExecuteJobs()

    addr := e.Vars()["addr"]
    fmt.Println("addr", addr)
    //0x60, 5050

    e.WaitForBlock()
    got := eth.GetStorageAt(addr, "0x60")
    if got != "5050"{
        t.Error("got:", got, "expected:", "0x5050")
    }
    eth.Stop()
}

func TestModifyDeploy(t *testing.T){
    eth := NewEthNode()
    e := epm.NewEPM(epm.NewEthD(eth), ".epm-log-test")
    e.Ch = Subscribe(eth, "newBlock")

    err := e.Parse(path.Join(epm.TestPath, "test_modify_deploy.epm"))
    if err != nil{
        t.Error(err)
    }
    e.ExecuteJobs()

    addr := e.Vars()["doug"]
    addr2 := e.Vars()["doug2"]
    fmt.Println("doug addr", addr)
    fmt.Println("doug addr2", addr2)
    //0x60, 0x5050

    e.WaitForBlock()
    got1 := eth.GetStorageAt(addr, "0x60")
    if got1 != "5050"{
        t.Error("got:", got1, "expected:", "0x5050")
    }
    got2 := eth.GetStorageAt(addr2, "0x60")
    if got2 != addr[2:]{
        t.Error("got:", got2, "expected:", addr)
    }
    eth.Stop()
}

// doesn't work unless we wait a block until actually making the query
// not going to fly here
func iTestQuery(t *testing.T){
    eth := NewEthNode()
    e := epm.NewEPM(epm.NewEthD(eth), ".epm-log-test")

    err := e.Parse(path.Join(epm.TestPath, "test_query.epm"))
    if err != nil{
        t.Error(err)
    }
    e.ExecuteJobs()

    ch := make(chan ethreact.Event, 1) 
    eth.Ethereum.Reactor().Subscribe("newBlock", ch)
    _ = <- ch
    a := e.Vars()["B"]
    if a != "0x5050"{
        t.Error("got:", a, "expecxted:", "0x5050")
    }
}

func TestStack(t *testing.T){
    eth := NewEthNode()
    e := epm.NewEPM(epm.NewEthD(eth), ".epm-log-test")
    e.Ch = Subscribe(eth, "newBlock")

    err := e.Parse(path.Join(epm.TestPath, "test_parse.epm"))
    if err != nil{
        t.Error(err)
    }
    e.ExecuteJobs()

    addr1 := e.Vars()["A"]
    addr2 := e.Vars()["B"]
    addr3 := e.Vars()["D"]
    fmt.Println("addr", addr1)
    fmt.Println("addr2", addr2)
    fmt.Println("addr3", addr3)
    //0x60, 0x5050

    e.WaitForBlock()
    got := eth.GetStorageAt(addr2, addr1)
    if got != "15"{
        t.Error("got:", got, "expected:", "0x15")
    }
    got = eth.GetStorageAt(addr3, "0x43")
    if got != "8080"{
        t.Error("got:", got, "expected:", "0x8080")
    }
    got = eth.GetStorageAt(addr3, addr1)
    if got != "15"{
        t.Error("got:", got, "expected:", "0x15")
    }
    got = eth.GetStorageAt(addr2, "0x12")
    if "0x"+got != epm.Coerce2Hex("ethan"){
        t.Error("got:", got, "expected:", epm.Coerce2Hex("ethan"))
    }
    eth.Stop()
}

// not a real test since the diffs just print we don't have access to them programmatically yet
// TODO>..
func TestDiff(t *testing.T){
    eth := NewEthNode()
    e := epm.NewEPM(epm.NewEthD(eth), ".epm-log-test")
    e.Ch = Subscribe(eth, "newBlock")

    err := e.Parse(path.Join(epm.TestPath, "test_diff.epm"))
    if err != nil{
        t.Error(err)
    }

    e.Diff = true
    e.ExecuteJobs()

    e.WaitForBlock()
}


