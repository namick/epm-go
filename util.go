package epm

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func usr() string {
	u, _ := user.Current()
	return u.HomeDir
}

// make sure command is valid
func checkCommand(cmd string) bool {
	r := false
	for _, c := range CMDS {
		if c == cmd {
			r = true
		}
	}
	return r
}

// peel off the next command and its args
func peelCmd(lines *[]string, startLine int) (*Job, error) {
	job := Job{"", []string{}}
	for line, t := range *lines {
		// ignore comments and blank lines
		tt := strings.TrimSpace(t)
		if len(tt) == 0 || tt[0:1] == "#" {
			continue
		}
		// ignore comments at end of the line
		t = strings.Split(t, "#")[0]

		// if no cmd yet, this should be a cmd
		if job.cmd == "" {
			// cmd syntax check
			t = strings.TrimSpace(t)
			l := len(t)
			if t[l-1:] != ":" {
				return nil, fmt.Errorf("Syntax error: missing ':' on line %d", line+startLine)
			}
			cmd := t[:l-1]
			// ensure known cmd
			if !checkCommand(cmd) {
				return nil, fmt.Errorf("Invalid command '%s' on line %d", cmd, line+startLine)
			}
			job.cmd = cmd
			continue
		}

		// if the line does not begin with white space, we are done
		if !(t[0:1] == " " || t[0:1] == "\t") {
			// peel off lines we've read
			*lines = (*lines)[line:]
			return &job, nil
		}
		t = strings.TrimSpace(t)
		// if there is a diff statement, we are done
		if t[:2] == StateDiffOpen || t[:2] == StateDiffClose {
			*lines = (*lines)[line:]
			return &job, nil
		}

		// the line is args. parse them
		// first, eliminate prefix whitespace/tabs
		args := strings.Split(t, "=>")
		// should be 'arg1 => arg2'
		// TODO: tailor lengths to the job cmd
		if len(args) > 3 {
			return nil, fmt.Errorf("Syntax error: improper argument formatting on line %d", line+startLine)
		}
		for _, a := range args {
			shaven := strings.TrimSpace(a)

			job.args = append(job.args, shaven)
		}
	}
	// only gets here if we finish all the lines
	*lines = nil
	return &job, nil
}

func parseStateDiff(lines *[]string, startLine int) string {
	for n, t := range *lines {
		tt := strings.TrimSpace(t)
		if len(tt) == 0 || tt[0:1] == "#" {
			continue
		}
		t = strings.Split(t, "#")[0]
		t = strings.TrimSpace(t)
		if len(t) == 0 {
			continue
		} else if len(t) > 2 && (t[:2] == StateDiffOpen || t[:2] == StateDiffClose) {
			// we found a match
			// shave previous lines
			*lines = (*lines)[n:]
			// see if there are other diff statements on this line
			i := strings.IndexAny(t, " \t")
			if i != -1 {
				(*lines)[0] = (*lines)[0][i:]
			} else if len(*lines) >= 1 {
				*lines = (*lines)[1:]
			}
			return t[2:]
		} else {
			*lines = (*lines)[n:]
			return ""
		}
	}
	return ""
}

// takes a simple string of nums/hex and ops
// all strings should have been removed
func tokenize(s string) []string {
	tokens := []string{}
	r_opMatch := regexp.MustCompile(`\+|\-|\*`)
	m_inds := r_opMatch.FindAllStringSubmatchIndex(s, -1)
	// if no ops, just return the string
	if len(m_inds) == 0 {
		return []string{s}
	}
	// for each theres a symbol and hex/num after it
	l := 0
	for i, matchI := range m_inds {
		i0 := matchI[0]
		i1 := matchI[1]
		ss := s[l:i0]
		l = i1
		if len(ss) != 0 {
			tokens = append(tokens, ss)
		}
		tokens = append(tokens, s[i0:i1])
		if i == len(m_inds)-1 {
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
func DoMath(args []string) []string {
	margs := []string{} // return
	//fmt.Println("domath:", args)
	r_stringMatch := regexp.MustCompile(`\"(.*?)\"`) //"

	for _, a := range args {
		//fmt.Println("time to tokenize:", a)
		tokens := []string{}
		// find all strings (between "")
		strMatches := r_stringMatch.FindAllStringSubmatchIndex(a, -1)
		// grab the expression before the first string
		if len(strMatches) > 0 {
			// loop through every interval between strMatches
			// tokenize, append to tokens
			l := 0
			for j, matchI := range strMatches {
				i0 := matchI[0]
				i1 := matchI[1]
				// get everything before this token
				s := a[l:i0]
				l = i1
				// if s is empty, add the string to tokens, move on
				if len(s) == 0 {
					tokens = append(tokens, a[i0+1:i1-1])
				} else {
					t := tokenize(s)
					tokens = append(tokens, t...)
					tokens = append(tokens, a[i0+1:i1-1])
				}
				// if we're on the last one, get anything that comes after
				if j == len(strMatches)-1 {
					s := a[l:]
					if len(s) > 0 {
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
		if err != nil {
			log.Fatal(err)
		}
		result := new(big.Int).SetBytes(tokenBigBytes)
		// start with the second token, and go up in twos (should be odd num of tokens)
		for j := 0; j < (len(tokens)-1)/2; j++ {
			op := tokens[2*j+1]
			n := tokens[2*j+2]
			nBigBytes, err := hex.DecodeString(stripHex(Coerce2Hex(n)))
			if err != nil {
				log.Fatal(err)
			}
			tokenBigInt := new(big.Int).SetBytes(nBigBytes)
			switch op {
			case "+":
				result.Add(result, tokenBigInt)
			case "-":
				result.Sub(result, tokenBigInt)
			case "*":
				result.Mul(result, tokenBigInt)
			}
		}
		// TODO: deal with 32-byte overflow
		resultHex := "0x" + hex.EncodeToString(result.Bytes())
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
	if buf.Len() > N {
		return buf.Bytes()[buf.Len()-N:]
	}
	return buf.Bytes()
}

// s can be string, hex, or int.
// returns properly formatted 32byte hex value
func Coerce2Hex(s string) string {
	//fmt.Println("coercing to hex:", s)
	// is int?
	i, err := strconv.Atoi(s)
	if err == nil {
		return "0x" + hex.EncodeToString(NumberToBytes(int32(i), i/256+1))
	}
	// is already prefixed hex?
	if len(s) > 1 && s[:2] == "0x" {
		if len(s)%2 == 0 {
			return s
		}
		return "0x0" + s[2:]
	}
	// is unprefixed hex?
	if len(s) > 32 {
		return "0x" + s
	}
	pad := strings.Repeat("\x00", (32-len(s))) + s
	ret := "0x" + hex.EncodeToString([]byte(pad))
	//fmt.Println("result:", ret)
	return ret
}

func addHex(s string) string {
	if len(s) < 2 {
		return "0x" + s
	}

	if s[:2] != "0x" {
		return "0x" + s
	}

	return s
}

func stripHex(s string) string {
	if len(s) > 1 {
		if s[:2] == "0x" {
			return s[2:]
		}
	}
	return s
}

// split line and trim space
func parseLine(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimRight(line, ";")

	args := strings.Split(line, ";")
	for i, a := range args {
		args[i] = strings.TrimSpace(a)
	}
	return args
}

func CheckMakeTmp() {
	_, err := os.Stat(path.Join(EPMDIR, ".tmp"))
	if err != nil {
		err := os.MkdirAll(path.Join(EPMDIR, ".tmp"), 0777) //wtf!
		if err != nil {
			fmt.Println("Could not make directory. Exiting", err)
			os.Exit(0)
		}
	}
	// copy the current dir into .tmp. Necessary for finding include files after a modify. :sigh:
	root := path.Base(ContractPath)
	if _, err = os.Stat(path.Join(EPMDIR, ".tmp", root)); err != nil {
		cmd := exec.Command("cp", "-r", ContractPath, path.Join(EPMDIR, ".tmp"))
		err = cmd.Run()
		if err != nil {
			fmt.Println("error copying working dir into tmp:", err)
			os.Exit(0)
		}
	}
}
