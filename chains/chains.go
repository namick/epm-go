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
	s := string(b)
	_, id, _ := SplitRef(s)
	return id

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
func ResolveChainType(chainType string) (string, error) {
	switch chainType {
	case "thel", "thelonious", "monk":
		return "thelonious", nil
	case "btc", "bitcoin":
		return "bitcoin", nil
	case "eth", "ethereum":
		return "ethereum", nil
	case "gen", "genesis":
		return "thelonious", nil
	}
	return "", fmt.Errorf("Unknown chain type: %s", chainType)
}

// Determines the chainId from a chainId prefix or from a ref, but not from a dapp.
func ResolveChainId(chainType, name, chainId string) (string, error) {
	var err error
	chainType, err = ResolveChainType(chainType)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("Could not locate %s chain by name %s or by id %s", chainType, name, chainId)
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
		return "", fmt.Errorf("ChainId %s did not match any known chains. Did you specify the type correctly (-type)?", prefix)
	}
	return p, nil
}

// Maximum entries in the HEAD file
var MaxHead = 100

// Add a new entry (type/chainId) to the top of the HEAD file
// Expects the chain type to have already been checked
func changeHead(typ, head string) error {
	// read in the entire head file and clip
	// if we have reached the max length
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

	// add the new head
	var s string
	if head != "" {
		s = typ + "/"
	}
	s = s + head + "\n" + bsp
	err = ioutil.WriteFile(utils.HEAD, []byte(s), 0666)
	if err != nil {
		return err
	}
	return nil
}

// Change the head to null (no head)
func NullHead() error {
	return changeHead("", "")
}

// Add a new entry to the top of the HEAD file.
// Arguments are chain type and new head (chainId or ref name)
// The HEAD file is a running list of the latest head
// so we can go back if we mess up or forget.
func ChangeHead(typ, head string) error {
	var err error
	typ, err = ResolveChainType(typ)
	if err != nil {
		return err
	}

	head, err = ResolveChainId(typ, head, head)
	if err != nil {
		return err
	}
	return changeHead(typ, head)
}

// Add a reference name to a chainId
func AddRef(typ, id, ref string) error {
	_, err := os.Stat(path.Join(utils.Refs, ref))
	if err == nil {
		return fmt.Errorf("Ref %s already exists", ref)
	}

	typ, err = ResolveChainType(typ)
	if err != nil {
		return err
	}

	dataDir := path.Join(utils.Blockchains, typ)
	_, err = os.Stat(path.Join(dataDir, id))
	if err != nil {
		id, err = findPrefixMatch(dataDir, id)
		if err != nil {
			return err
		}
	}

	refid := path.Join(typ, id)
	return ioutil.WriteFile(path.Join(utils.Refs, ref), []byte(refid), 0644)
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
// Returns chain type and chain id
func GetHead() (string, string, error) {
	// TODO: only read the one line!
	f, err := ioutil.ReadFile(utils.HEAD)
	if err != nil {
		return "", "", err
	}
	fspl := strings.Split(string(f), "\n")
	head := fspl[0]
	if head == "" {
		return "", "", nil
	}
	return SplitRef(head)
}

func SplitRef(ref string) (string, string, error) {
	sp := strings.Split(ref, "/")
	if len(sp) != 2 {
		return "", "", fmt.Errorf("Improperly formatted ref:", ref)
	}
	return sp[0], sp[1], nil
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
