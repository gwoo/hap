// Hap - the simple and effective provisioner
// Copyright (c) 2019 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package cli

import (
	"fmt"

	"github.com/gwoo/hap"
)

// Add the Push command
func init() {
	Commands.Add("push", &PushCmd{})
}

// PushCmd is the push command
type PushCmd struct{}

// IsRemote returns whether the command expects a remote
func (cmd *PushCmd) IsRemote() bool {
	return true
}

// Help returns help on the hap push command
func (cmd *PushCmd) Help() string {
	return "hap push\tPush current repo to the remote."
}

// Run takes a remote and pushes to it
func (cmd *PushCmd) Run(remote *hap.Remote) (string, error) {
	fmt.Printf("[%s] connecting to %s\n", remote.Host.Name, remote.Host.Addr)
	if err := remote.Push(); err != nil {
		result := fmt.Sprintf("[%s] push failed.", remote.Host.Name)
		return result, err
	}
	if err := remote.PushSubmodules(); err != nil {
		result := fmt.Sprintf("[%s] push failed.", remote.Host.Name)
		return result, err
	}
	result := fmt.Sprintf("[%s] push completed.", remote.Host.Name)
	return result, nil
}
