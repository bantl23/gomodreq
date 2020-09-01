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
		return nil, fmt.Errorf("file does not exist %s [%+v]", uri.Path, err)
	}
	data, err := ioutil.ReadFile(uri.Path)
	if err != nil {
		return nil, fmt.Errorf("problem reading file %s [%+v]", uri.Path, err)
	}
	return data, nil
}

// getHTTPData gets data from a http resource
func getHTTPData(uri *url.URL) ([]byte, error) {
	client := http.Client{}
	resp, err := client.Get(uri.String())
	if err != nil {
		return nil, fmt.Errorf("unable to get webpage %s [%+v]", uri.String(), err)
	}
	defer resp.Body.Close()
	b := &bytes.Buffer{}
	_, err = b.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from %s [%+v]", uri.String(), err)
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
		return nil, fmt.Errorf("unable to get home directory for %s [%+v]", username, err)
	}

	identityFile := filepath.Join(homeDir, ".ssh", "id_rsa")
	key, err := ioutil.ReadFile(identityFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read identity file %s [%+v]", identityFile, err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse identity file %s [%+v]", identityFile, err)
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
		return nil, fmt.Errorf("unable to connect to host %s [%+v]", hostname, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("unable to create new ssh session [%+v]", err)
	}
	defer session.Close()

	b, err := session.CombinedOutput("cat " + uri.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to get remote file contents %s [%+v]", uri.Path, err)
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
		return nil, fmt.Errorf("Unsupported uri scheme: %s", uri.Scheme)
	}

	return getDataFunc(uri)
}

func getReq(loc string) (*Requirements, error) {
	req := new(Requirements)

	uri, err := url.ParseRequestURI(loc)
	if err != nil {
		return nil, fmt.Errorf("unable to parse uri %s [%+v]", uri, err)
	}
	data, err := getData(uri)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, req)
	if err != nil {
		return nil, fmt.Errorf("unable to parse requirements file %s [%+v]", uri, err)
	}

	return req, nil
}

func getMod(loc string) (*modfile.File, error) {
	uri, err := url.ParseRequestURI(loc)
	if err != nil {
		return nil, fmt.Errorf("unable to parse uri %s [%+v]", uri, err)
	}
	data, err := getData(uri)
	if err != nil {
		return nil, err
	}
	mod, err := modfile.Parse(uri.Path, data, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to parse mod file %s [%+v]", uri, err)
	}

	return mod, nil
}

func checkRequired(req *Requirements, mod *modfile.File, exitCode int) (int, error) {
	for i := range mod.Require {
		path := mod.Require[i].Mod.Path
		modvers := mod.Require[i].Mod.Version
		_, ok := req.Required[path]
		if ok == true {
			re, err := regexp.Compile(req.Required[path])
			if err != nil {
				return exitCode, fmt.Errorf("unable to compile regex %s [%+v]", req.Required[path], err)
			}
			if !re.Match([]byte(modvers)) {
				fmt.Printf("Error: package %s version %s does not met requirements [regex is %s]\n", path, modvers, req.Required[path])
				exitCode = 1
			}
		}
	}
	return exitCode, nil
}

func checkBanned(req *Requirements, mod *modfile.File, exitCode int) (int, error) {
	for i := range mod.Require {
		path := mod.Require[i].Mod.Path
		modvers := mod.Require[i].Mod.Version
		_, ok := req.Banned[path]
		if ok == true {
			for j := range req.Banned[path] {
				re, err := regexp.Compile(req.Banned[path][j])
				if err != nil {
					return exitCode, fmt.Errorf("unable to compile regex %s [%+v]", req.Banned[path][j], err)
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
		fmt.Println("Fatal: unable to get current directory")
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
	reqLoc := os.Args[1:]
	if len(reqLoc) == 0 {
		reqLoc = []string{"file://" + path + "/.gomodreq.yml"}
	}

	mod, err := getMod(modLoc)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	reqExitCode := 0
	banExitCode := 0
	for i := range reqLoc {
		fmt.Printf("Checking %s\n", reqLoc[i])
		req, err := getReq(reqLoc[i])
		if err != nil {
			fmt.Println(err)
			return -1
		}
		reqExitCode, err = checkRequired(req, mod, reqExitCode)
		if err != nil {
			fmt.Println(err)
			return -1
		}
		banExitCode, err = checkBanned(req, mod, banExitCode)
		if err != nil {
			fmt.Println(err)
			return -1
		}
	}

	if reqExitCode == 0 && banExitCode == 0 {
		fmt.Println("All module requirements met")
	}

	return reqExitCode + banExitCode
}

func main() {
	os.Exit(mainWithExit())
}
