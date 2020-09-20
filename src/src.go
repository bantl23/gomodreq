package src

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

// SourceI is the inteface to get data
type SourceI interface {
	GetData(*url.URL) ([]byte, error)
}

// Source is the structured used to get data
type Source struct {
	URI *url.URL
}

// GetData gets the data from requirements source
func (s Source) GetData() ([]byte, error) {
	switch s.URI.Scheme {
	case "file":
		return s.getFileData()
	case "http", "https":
		return s.getHTTPData()
	case "ssh":
		return s.getSSHData()
	default:
		return nil, fmt.Errorf("Unsupported uri scheme: %s", s.URI.Scheme)
	}
}

// getFileData gets data from a file
func (s Source) getFileData() ([]byte, error) {
	_, err := os.Stat(s.URI.Path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist %s [%+v]", s.URI.Path, err)
	}
	data, err := ioutil.ReadFile(s.URI.Path)
	if err != nil {
		return nil, fmt.Errorf("problem reading file %s [%+v]", s.URI.Path, err)
	}
	return data, nil
}

// getHTTPData gets data from a http resource
func (s Source) getHTTPData() ([]byte, error) {
	client := http.Client{}
	resp, err := client.Get(s.URI.String())
	if err != nil {
		return nil, fmt.Errorf("unable to get webpage %s [%+v]", s.URI.String(), err)
	}
	defer resp.Body.Close()
	b := &bytes.Buffer{}
	_, err = b.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from %s [%+v]", s.URI.String(), err)
	}
	return b.Bytes(), nil
}

// getSSHData gets data from ssh resource
func (s Source) getSSHData() ([]byte, error) {
	username := s.URI.User.Username()
	password, pwOk := s.URI.User.Password()
	hostname := s.URI.Host
	if !strings.Contains(s.URI.Host, ":") {
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

	b, err := session.CombinedOutput("cat " + s.URI.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to get remote file contents %s [%+v]", s.URI.Path, err)
	}
	return b, nil
}
