// Hap - the simple and effective provisioner
// Copyright (c) 2019 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package cli

import (
	"fmt"

	"github.com/gwoo/hap"
)

// Add the build command
func init() {
	Commands.Add("deploy", &DeployCmd{})
}

// DeployCmd is the build command
type DeployCmd struct{}

// IsRemote returns whether this command expects a remote
func (cmd *DeployCmd) IsRemote() bool {
	return true
}

// Help returns help for the build command
func (cmd *DeployCmd) Help() string {
	return "hap deploy <name>\tRun the named deploy defined in the Hapfile."
}

// Run the build command on the remote host
func (cmd *DeployCmd) Run(remote *hap.Remote) (string, error) {
	if *dry {
		result := fmt.Sprintf(
			"[%s] --dry run.\n",
			remote.Host.Name,
		)
		cmds := []string{"cd " + remote.Dir}
		cmds = remote.Host.AddEnv(cmds)
		cmds = append(cmds, remote.Host.Cmds()...)
		for _, cmd := range cmds {
			result = result + fmt.Sprintf("[%s] %s\n", remote.Host.Name, cmd)
		}
		result = result + fmt.Sprintf("[%s] --dry run completed.\n", remote.Host.Name)
		return result, nil
	}
	if result, err := Commands.Get("push").Run(remote); err != nil {
		return result, err
	}
	if err := remote.Build(true); err != nil {
		result := fmt.Sprintf("[%s] deploy failed.", remote.Host.Name)
		return result, err
	}
	result := fmt.Sprintf("[%s] deploy completed.", remote.Host.Name)
	return result, nil
}
