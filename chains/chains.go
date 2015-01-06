package chains

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/eris-ltd/decerver-interfaces/dapps"
	"github.com/eris-ltd/epm-go/utils"
)

// Get ChainId from a reference name by reading the ref file
func ChainIdFromName(name string) string {
	refsPath := path.Join(utils.Blockchains, "refs", name)
	b, err := ioutil.ReadFile(refsPath)
	if err != nil {
		return ""
	}
	return string(b)
}

// Get ChainId from dapp name by reading package.json file
func ChainIdFromDapp(dapp string) (string, error) {
	p, err := CheckGetPackageFile(path.Join(utils.Apps, dapp))
	if err != nil {
		return "", err
	}

	var chainId string
	for _, dep := range p.ModuleDependencies {
		if dep.Name == "monk" {
			d := &dapps.MonkData{}
			if err := json.Unmarshal(dep.Data, d); err != nil {
				return "", err
			}
			chainId = d.ChainId
		}
	}
	if chainId == "" {
		return "", fmt.Errorf("Dapp is missing monk dependency or chainId!")
	}

	return chainId, nil
}

// Allow chain types to be specified in shorter form (ie. 'eth' for 'ethereum')
func ResolveChainType(chainType string) string {
	switch chainType {
	case "thel", "thelonious", "monk":
		return "thelonious"
	case "btc", "bitcoin":
		return "bitcoin"
	case "eth", "ethereum":
		return "ethereum"
	case "gen", "genesis":
		return "thelonious"
	}
	return ""
}

// Determines the chainId from a chainId prefix or from a ref, but not from a dapp.
func ResolveChainId(chainType, name, chainId string) (string, error) {
	chainType = ResolveChainType(chainType)
	if chainType == "" {
		return "", fmt.Errorf("Unknown chain type: %s", chainType)
	}

	var p string
	idFromName := ChainIdFromName(name)
	if idFromName != "" {
		chainId = idFromName
	}

	if chainId != "" {
		p = path.Join(utils.Blockchains, chainType, chainId)
		if _, err := os.Stat(p); err != nil {
			// see if its a prefix of a chainId
			id, err := findPrefixMatch(path.Join(utils.Blockchains, chainType), chainId)
			if err != nil {
				return "", err
			}
			p = path.Join(utils.Blockchains, chainType, id)
			chainId = id
		}
	}
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("Could not locate chain by name %s or by id %s", name, chainId)
	}

	return chainId, nil
}

// Return full path to a blockchain's directory
func ResolveChain(chainType, name, chainId string) (string, error) {
	id, err := ResolveChainId(chainType, name, chainId)
	if err != nil {
		return "", err
	}
	return path.Join(utils.Blockchains, chainType, id), nil
}

// lookup chainIds by prefix match
func findPrefixMatch(dirPath, prefix string) (string, error) {
	fs, _ := ioutil.ReadDir(dirPath)
	found := false
	var p string
	for _, f := range fs {
		if strings.HasPrefix(f.Name(), prefix) {
			if found {
				return "", fmt.Errorf("ChainId collision! Multiple chains begin with %s. Please be more specific", prefix)
			}
			p = f.Name() //path.Join(Blockchains, chainType, f.Name())
			found = true
		}
	}
	if !found {
		return "", fmt.Errorf("ChainId %s did not match any known chains", prefix)
	}
	return p, nil
}

// Maximum entries in the HEAD file
var MaxHead = 100

// Add a new entry to the top of the HEAD file
func changeHead(head string) error {
	b, err := ioutil.ReadFile(utils.HEAD)
	if err != nil {
		return err
	}
	bspl := strings.Split(string(b), "\n")
	var bsp string
	if len(bspl) >= MaxHead {
		bsp = strings.Join(bspl[:MaxHead-1], "\n")
	} else {
		bsp = string(b)
	}
	bsp = head + "\n" + bsp
	err = ioutil.WriteFile(utils.HEAD, []byte(bsp), 0666)
	if err != nil {
		return err
	}
	return nil
}

// Change the head to null (no head)
func NullHead() error {
	return changeHead("")
}

// Add a new entry to the top of the HEAD file.
// Argument is chainId or ref name
// The HEAD file is a running list of the latest head
// so we can go back if we mess up or forget.
func ChangeHead(head string) error {
	head, err := ResolveChainId("thelonious", head, head)
	if err != nil {
		return err
	}
	return changeHead(head)
}

// Add a reference name to a chainId
func AddRef(id, ref string) error {
	_, err := os.Stat(path.Join(utils.Refs, ref))
	if err == nil {
		return fmt.Errorf("Ref %s already exists", ref)
	}

	dataDir := path.Join(utils.Blockchains, "thelonious")
	_, err = os.Stat(path.Join(dataDir, id))
	if err != nil {
		id, err = findPrefixMatch(dataDir, id)
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(path.Join(utils.Refs, ref), []byte(id), 0644)
}

// Return a list of chain references
func GetRefs() (map[string]string, error) {
	fs, err := ioutil.ReadDir(utils.Refs)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, f := range fs {
		name := f.Name()
		b, err := ioutil.ReadFile(path.Join(utils.Refs, name))
		if err != nil {
			return nil, err
		}
		m[name] = string(b)
	}
	return m, nil
}

// Get the current active chain (top of the HEAD file)
func GetHead() (string, error) {
	// TODO: only read the one line!
	f, err := ioutil.ReadFile(utils.HEAD)
	if err != nil {
		return "", err
	}
	fspl := strings.Split(string(f), "\n")
	return fspl[0], nil
}

// Return a dapp package file
func CheckGetPackageFile(dappDir string) (*dapps.PackageFile, error) {
	if _, err := os.Stat(dappDir); err != nil {
		return nil, fmt.Errorf("Dapp %s not found", dappDir)
	}

	b, err := ioutil.ReadFile(path.Join(dappDir, "package.json"))
	if err != nil {
		return nil, err
	}

	p, err := dapps.NewPackageFileFromJson(b)
	if err != nil {
		return nil, err
	}
	return p, nil
}

