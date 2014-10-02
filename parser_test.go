package epm

import(
    "path"
    "testing"
)


var ExpectedParse = []Job{
    Job{"deploy", []string{"a.lll", "{{A}}"}},
    Job{"modify-deploy", []string{"b.lll", "{{B}}", "(def 'dougie 0x1313)", "(def 'dougie {{A}})"}},
    Job{"modify-deploy", []string{"b.lll", "{{D}}", "(def 'dougie 0x1313)", "(def 'dougie {{A}})", "[[0x42]]", "[[0x43]]"}},
    Job{"transact", []string{"{{B}}", "0x15 0x12"}},
    Job{"query", []string{"{{B}}", "0x15", "{{C}}"}},
    Job{"endow", []string{"{{A}}", "0x609"}},
}

func TestParse(t *testing.T){
    e := &EPM{nil, []Job{}, make(map[string]string), ".epm-log-test"}
    err := e.Parse(path.Join(TestPath, "test_parse.epm"))
    if err != nil{
        t.Error(err)
    }

    jobs := e.Jobs()
    for i, j := range jobs{
        expected := ExpectedParse[i]
        if !checkExpectedJobs(j, expected){
            t.Error("got:", j, "expected:", expected) 
        }
    }
}

var ExpectedVarSub_map= map[string]string{"A":"hello"}

var ExpectedVarSub_jobs = []Job{
    Job{"set", []string{"{{A}}", "hello"}},
    Job{"transact", []string{"hello", "0x15 0x12"}},
    Job{"modify-deploy", []string{"a.lll", "{{C}}", "(def 'dougie 0x1313)", "(def 'dougie hello)"}},
}

    
func TestVarSub(t *testing.T){
    e := &EPM{nil, []Job{}, make(map[string]string), ".epm-log-test"}
    err := e.Parse(path.Join(TestPath, "test_varsub.epm"))
    if err != nil{
        t.Error(err)
    }

    jobs := e.Jobs()
    for i, j := range jobs{
        if j.cmd == "set"{
            e.ExecuteJob(j)
        } else{
            e.VarSub(&j)
        }
        expected := ExpectedVarSub_jobs[i]
        if !checkExpectedJobs(j, expected){
            t.Error("got:", j, "expected:", expected) 
        }
    }
    if !checkExpectedMaps(e.Vars(), ExpectedVarSub_map){
        t.Error("got:", e.vars, "expected:", ExpectedVarSub_map)
    }
    
}

func checkExpectedMaps(m1 map[string]string, m2 map[string]string) bool{

    if len(m1) != len(m2){
        return false
    }

    for k, v := range m1{
        if v != m2[k]{
            return false
        }
    }
    return true
}

func checkExpectedJobs(j1 Job, j2 Job) bool{
    if j1.cmd != j2.cmd{
        return false
    }
    if len(j1.args) != len(j2.args){
        return false
    }
    for i, _ := range j1.args{
        if j1.args[i] != j2.args[i]{
            return false
        }
    }
    return true
}
