// Hap - the simple and effective provisioner
// Copyright (c) 2017 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"testing"
)

func TestNewClientConfig(t *testing.T) {
	sshConfig := SSHConfig{
		Addr:     "10.0.0.1:22",
		Username: "bob",
		Identity: "~/.ssh/id_rsa",
		Password: "password",
	}
	clientConfig, err := NewClientConfig(sshConfig)
	if err != nil {
		t.Error(err)
	}
	if len(clientConfig.Auth) != 3 {
		t.Error("There should be 3 auth methods")
	}
}
