// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.
package hap

import (
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Config necessary for ssh connections
type SshConfig struct {
	Addr         string
	Username     string
	Identity     string
	Password     string
	ClientConfig *ssh.ClientConfig
}

// Construct a new client config
func NewClientConfig(config SshConfig) (*ssh.ClientConfig, error) {
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	agent := agent.NewClient(sock)
	signers, err := agent.Signers()
	if err != nil {
		return nil, err
	}
	key, err := NewKey(config.Identity)
	if err != nil {
		return nil, err
	}
	pk, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, err
	}
	signers = append(signers, pk)

	auths := []ssh.AuthMethod{
		ssh.PublicKeys(signers...),
		ssh.Password(config.Password),
	}
	cfg := &ssh.ClientConfig{User: config.Username, Auth: auths}
	cfg.SetDefaults()
	return cfg, nil
}

// Get the data in the key file
func NewKeyFile(key string) (string, error) {
	if string(key[0]) == "~" {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		key = strings.Replace(key, "~", u.HomeDir, 1)
	}
	return filepath.EvalSymlinks(key)
}

// Parse and return the interface for the key type (rsa, dss, etc)
func NewKey(key string) (interface{}, error) {
	file, err := NewKeyFile(key)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return ssh.ParseRawPrivateKey(b)
}
