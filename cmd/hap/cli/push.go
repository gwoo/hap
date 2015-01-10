// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
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

// Push command
type PushCmd struct{}

// Does this command expect a remote
func (cmd *PushCmd) IsRemote() bool {
	return true
}

// Get help on the push command
func (cmd *PushCmd) Help() string {
	return "hap push\tPush current repo to the remote."
}

// Push to the remote
func (cmd *PushCmd) Run(remote *hap.Remote) (string, error) {
	if err := remote.PushSubmodules(); err != nil {
		result := fmt.Sprintf("[%s] push failed.", remote.Host.Name)
		return result, err
	}
	if err := remote.Push(); err != nil {
		result := fmt.Sprintf("[%s] push failed.", remote.Host.Name)
		return result, err
	}
	result := fmt.Sprintf("[%s] push completed.", remote.Host.Name)
	return result, nil
}
