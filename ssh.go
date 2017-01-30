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

// NewClientConfig constructs a new *ssh.ClientConfig
func NewClientConfig(config SSHConfig) (*ssh.ClientConfig, error) {
	methods := make([]ssh.AuthMethod, 0)
	if sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		method := ssh.PublicKeysCallback(agent.NewClient(sock).Signers)
		methods = append(methods, method)
	}
	if config.Username == "" {
		u, _ := user.Current()
		config.Username = u.Name
	}
	if config.Identity != "" {
		if method := NewPublicKeyMethod(config.Identity); method != nil {
			methods = append(methods, method)
		}
	} else {
		home := os.Getenv("HOME")
		keys := []string{home + "/.ssh/id_rsa", home + "/.ssh/id_dsa"}
		for _, key := range keys {
			if method := NewPublicKeyMethod(key); method != nil {
				methods = append(methods, method)
			}
		}
	}
	if config.Password != "" {
		methods = append(methods, ssh.Password(config.Password))
	}
	cfg := &ssh.ClientConfig{User: config.Username, Auth: methods}
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
func NewKey(key string) (ssh.Signer, error) {
	file, err := NewKeyFile(key)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(b)
}

// NewPublicKeyMethod creates a new auth method for public keys
func NewPublicKeyMethod(key string) ssh.AuthMethod {
	pk, err := NewKey(key)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(pk)
}
