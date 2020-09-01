package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v2"
)

// set by build flags
var (
	version = "v0.0.0"
	commit  = "dev"
)

// Requirements contains the golang code requirements
type Requirements struct {
	Required map[string]string   `yaml:"required,omitempty"`
	Banned   map[string][]string `yaml:"banned,omitempty"`
}

// getFileData gets data from a file
func getFileData(uri *url.URL) ([]byte, error) {
	_, err := os.Stat(uri.Path)
	if os.IsNotExist(err) {
		return nil, err
	}
	data, err := ioutil.ReadFile(uri.Path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// getHTTPData gets data from a http resource
func getHTTPData(uri *url.URL) ([]byte, error) {
	client := http.Client{}
	resp, err := client.Get(uri.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b := &bytes.Buffer{}
	_, err = b.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// getSSHData gets data from ssh resource
func getSSHData(uri *url.URL) ([]byte, error) {
	username := uri.User.Username()
	password, pwOk := uri.User.Password()
	hostname := uri.Host
	if !strings.Contains(uri.Host, ":") {
		hostname = hostname + ":22"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadFile(filepath.Join(homeDir, ".ssh", "id_rsa"))
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	auth := []ssh.AuthMethod{
		ssh.PublicKeys(signer),
	}
	if pwOk == true {
		auth = append(auth, ssh.Password(password))
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", hostname, config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	b, err := session.CombinedOutput("cat " + uri.Path)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func getData(uri *url.URL) ([]byte, error) {
	var getDataFunc func(uri *url.URL) ([]byte, error)

	switch uri.Scheme {
	case "file":
		getDataFunc = getFileData
	case "http", "https":
		getDataFunc = getHTTPData
	case "ssh":
		getDataFunc = getSSHData
	default:
		return nil, fmt.Errorf("Unsupported scheme: %s", uri.Scheme)
	}

	return getDataFunc(uri)
}

func getReq(loc string) (*Requirements, error) {
	req := new(Requirements)

	uri, err := url.ParseRequestURI(loc)
	if err != nil {
		return nil, err
	}
	data, err := getData(uri)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func getMod(loc string) (*modfile.File, error) {
	uri, err := url.ParseRequestURI(loc)
	if err != nil {
		return nil, err
	}
	data, err := getData(uri)
	if err != nil {
		return nil, err
	}
	mod, err := modfile.Parse(uri.Path, data, nil)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func checkRequired(req *Requirements, mod *modfile.File) (int, error) {
	exitCode := 0
	for i := range mod.Require {
		path := mod.Require[i].Mod.Path
		modvers := mod.Require[i].Mod.Version
		_, ok := req.Required[path]
		if ok == true {
			re, err := regexp.Compile(req.Required[path])
			if err != nil {
				return exitCode, err
			}
			if !re.Match([]byte(modvers)) {
				fmt.Printf("Error: package %s version %s does not met requirements [regex is %s]\n", path, modvers, req.Required[path])
				exitCode = 1
			}
		}
	}
	return exitCode, nil
}

func checkBanned(req *Requirements, mod *modfile.File) (int, error) {
	exitCode := 0
	for i := range mod.Require {
		path := mod.Require[i].Mod.Path
		modvers := mod.Require[i].Mod.Version
		_, ok := req.Banned[path]
		if ok == true {
			for j := range req.Banned[path] {
				re, err := regexp.Compile(req.Banned[path][j])
				if err != nil {
					return exitCode, err
				}
				if re.Match([]byte(modvers)) {
					fmt.Printf("Error: package %s version %s is banned [regex is %s]\n", path, modvers, req.Banned[path][j])
					exitCode = 2
				}
			}
		}
	}
	return exitCode, nil
}

func mainWithExit() int {
	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return -1
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage:\n%s <mod_req_uri> (default: file://%s/.gomodreq.yml):\n",
			filepath.Base(os.Args[0]),
			path,
		)
		flag.PrintDefaults()
	}

	versFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versFlag == true {
		fmt.Printf("%s (%s)\n", version, commit)
		return 0
	}

	modLoc := "file://" + path + "/go.mod"
	reqLoc := "file://" + path + "/.gomodreq.yml"
	if len(os.Args) == 2 {
		reqLoc = os.Args[1]
	} else if len(os.Args) > 2 {
		fmt.Println("only one mod req file")
		return -1
	}

	req, err := getReq(reqLoc)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	mod, err := getMod(modLoc)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	reqEC, err := checkRequired(req, mod)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	banEC, err := checkBanned(req, mod)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	if reqEC == 0 && banEC == 0 {
		fmt.Println("All module requirements met")
	}

	return reqEC + banEC
}

func main() {
	os.Exit(mainWithExit())
}
