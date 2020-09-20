package source

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// GetData gets the data from requirements source
func GetData(uri *url.URL) ([]byte, error) {
	switch uri.Scheme {
	case "file":
		return getFileData(uri)
	case "http", "https":
		return getHTTPData(uri)
	case "ssh":
		return getSSHData(uri)
	default:
		return nil, fmt.Errorf("Unsupported uri scheme: %s", uri.Scheme)
	}
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
