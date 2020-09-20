package src

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// SSHSource is the structure to get data
type SSHSource struct {
	URI *url.URL
}

// GetData gets data from ssh resource
func (s SSHSource) GetData() ([]byte, error) {
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
