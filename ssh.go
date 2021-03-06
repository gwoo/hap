// Hap - the simple and effective provisioner
// Copyright (c) 2019 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"io/ioutil"
	"log"
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

// NewSSHConfig converts a Host into the SSHConfig
func NewSSHConfig(host *Host) SSHConfig {
	if host.Username == "" {
		u, _ := user.Current()
		host.Username = u.Name
	}
	if strings.HasPrefix(host.Identity, ".") {
		dir, _ := os.Getwd()
		host.Identity = dir + host.Identity[1:]
	}
	config := SSHConfig{
		Addr:     host.Addr,
		Username: host.Username,
		Identity: host.Identity,
		Password: host.Password,
	}
	return config
}

// NewClientConfig constructs a new *ssh.ClientConfig
func NewClientConfig(config SSHConfig) (*ssh.ClientConfig, error) {
	methods := make([]ssh.AuthMethod, 0)
	if sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		method := ssh.PublicKeysCallback(agent.NewClient(sock).Signers)
		methods = append(methods, method)
	}
	if config.Identity != "" {
		method, err := NewPublicKeyMethod(config.Identity)
		if err == nil {
			methods = append(methods, method)
		} else {
			log.Println("[identity]", err)
		}
	} else {
		home := os.Getenv("HOME")
		keys := []string{home + "/.ssh/id_rsa", home + "/.ssh/id_dsa"}
		for _, key := range keys {
			if method, err := NewPublicKeyMethod(key); err == nil {
				methods = append(methods, method)
			}
		}
	}
	if config.Password != "" {
		methods = append(methods, ssh.Password(config.Password))
	}
	cfg := &ssh.ClientConfig{
		Auth:            methods,
		User:            config.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	cfg.SetDefaults()
	return cfg, nil
}

// NewKeyFile takes a key and returns the key file
func NewKeyFile(key string) (string, error) {
	if string(key[0]) == "~" {
		var homeDir string
		u, err := user.Current()
		if err != nil {
			homeDir = os.Getenv("HOME")
		} else {
			homeDir = u.HomeDir
		}
		key = strings.Replace(key, "~", homeDir, 1)
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
func NewPublicKeyMethod(key string) (ssh.AuthMethod, error) {
	pk, err := NewKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(pk), nil
}
