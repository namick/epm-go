package main

import (
	"bytes"
	"github.com/eris-ltd/epm-go"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

func cleanupEPM() {
	dirs := []string{epm.EPMDIR, *database}
	for _, d := range dirs {
		err := os.RemoveAll(d)
		if err != nil {
			logger.Errorln("Error removing dir", d, err)
		}
	}
}

func installEPM() {
	cur, _ := os.Getwd()
	os.Chdir(path.Join(GoPath, "src", "github.com", "eris-ltd", "epm-go", "cmd", "epm"))
	cmd := exec.Command("go", "install")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	logger.Infoln(out.String())
	os.Chdir(cur)
}

func updateEPM() {
	cur, _ := os.Getwd()

	// pull changes
	os.Chdir(path.Join(GoPath, "src", "github.com", "eris-ltd", "epm-go"))
	cmd := exec.Command("git", "pull", "origin", "master")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	res := out.String()
	logger.Infoln(res)

	if strings.Contains(res, "up-to-date") {
		// return to original dir
		os.Chdir(cur)
		return
	}

	// install
	installEPM()

	// return to original dir
	os.Chdir(cur)
}

func cleanUpdateInstall() {
	if *clean && *update {
		cleanupEPM()
		updateEPM()
	} else if *clean {
		cleanupEPM()
		if *install {
			installEPM()
		}
	} else if *update {
		updateEPM()
	} else if *install {
		installEPM()
	}
}

// looks for pkg-def file
// exits if error (none or more than 1)
// returns dir of pkg, name of pkg (no extension) and whether or not there's a test file
func getPkgDefFile(pkgPath string) (string, string, bool) {
	logger.Infoln("Pkg path:", pkgPath)
	var pkgName string
	var test_ bool

	// if its not a directory, look for a corresponding test file
	f, err := os.Stat(pkgPath)
	if err != nil {
		logger.Errorln(err)
		os.Exit(0)
	}
	if !f.IsDir() {
		dir, fil := path.Split(pkgPath)
		spl := strings.Split(fil, ".")
		pkgName = spl[0]
		ext := spl[1]
		if ext != PkgExt {
			logger.Errorln("Did not understand extension. Got %s, expected %s\n", ext, PkgExt)
			os.Exit(0)
		}

		_, err := os.Stat(path.Join(dir, pkgName) + "." + TestExt)
		if err != nil {
			logger.Errorln("There was no test found for package-definition %s. Deploying without test ...\n", pkgName)
			test_ = false
		} else {
			test_ = true
		}
		return dir, pkgName, test_
	}

	// read dir for files
	files, err := ioutil.ReadDir(pkgPath)
	if err != nil {
		logger.Errorln("Could not read directory:", err)
		os.Exit(0)
	}
	// find all package-defintion and package-definition-test files
	candidates := make(map[string]int)
	candidates_test := make(map[string]int)
	for _, f := range files {
		name := f.Name()
		spl := strings.Split(name, ".")
		if len(spl) < 2 {
			continue
		}
		name = spl[0]
		ext := spl[1]
		if ext == PkgExt {
			candidates[name] = 1
		} else if ext == TestExt {
			candidates_test[name] = 1
		}
	}
	// exit if too many or no options
	if len(candidates) > 1 {
		logger.Errorln("More than one package-definition file available. Please select with the '-p' flag")
		os.Exit(0)
	} else if len(candidates) == 0 {
		logger.Errorln("No package-definition files found for extensions", PkgExt, TestExt)
		os.Exit(0)
	}
	// this should run once (there's only one candidate)
	for k, _ := range candidates {
		pkgName = k
		if candidates_test[pkgName] == 1 {
			test_ = true
		} else {
			logger.Infoln("There was no test found for package-definition %s. Deploying without test ...\n", pkgName)
			test_ = false
		}
	}
	return pkgPath, pkgName, test_
}
