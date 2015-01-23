// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package cli

import (
	"fmt"

	"github.com/gwoo/hap"
)

// Add the init command
func init() {
	Commands.Add("init", &InitCmd{})
}

// InitCmd struct for setting up remote repo
type InitCmd struct{}

// IsRemote returns whether the command expects a remote or not
func (cmd *InitCmd) IsRemote() bool {
	return true
}

// Help returns help on the hap init command
func (cmd *InitCmd) Help() string {
	return "hap init\tInitialize a new remote host."
}

// Run takes a remote and runs a command on it
func (cmd *InitCmd) Run(remote *hap.Remote) (string, error) {
	if err := remote.Initialize(); err != nil {
		result := fmt.Sprintf("[%s] init %s failed.", remote.Host.Name, remote.Dir)
		return result, err
	}
	result := fmt.Sprintf("[%s] init %s completed.", remote.Host.Name, remote.Dir)
	return result, nil
}
