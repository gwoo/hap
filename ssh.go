// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSHConfig holds the config for ssh connections
type SSHConfig struct {
	Addr         string
	Username     string
	Identity     string
	Password     string
	ClientConfig *ssh.ClientConfig
}

// NewClientConfig constructs a new client config
func NewClientConfig(config SSHConfig) (*ssh.ClientConfig, error) {
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	agent := agent.NewClient(sock)
	signers, err := agent.Signers()
	if err != nil {
		return nil, err
	}
	if config.Identity != "" {
		signer, err := NewSigner(config.Identity)
		if err != nil {
			return nil, err
		}
		signers = append(signers, signer)
	}
	auths := []ssh.AuthMethod{
		ssh.PublicKeys(signers...),
		ssh.Password(config.Password),
	}
	cfg := &ssh.ClientConfig{User: config.Username, Auth: auths}
	cfg.SetDefaults()
	return cfg, nil
}

// NewKeyFile takes a key and returns the key file
func NewKeyFile(key string) (string, error) {
	if string(key[0]) == "~" {
		u, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("[identity] %s", err)
		}
		key = strings.Replace(key, "~", u.HomeDir, 1)
	}
	return filepath.EvalSymlinks(key)
}

// NewKey parses and returns the interface for the key type (rsa, dss, etc)
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

// NewSigner creates a new ssh signer
func NewSigner(key string) (ssh.Signer, error) {
	pk, err := NewKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(pk)
}
