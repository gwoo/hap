// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package cli

import (
	"fmt"

	"github.com/gwoo/hap"
)

// Add the build command
func init() {
	Commands.Add("build", &BuildCmd{})
}

// BuildCmd is the build command
type BuildCmd struct{}

// IsRemote returns whether this command expects a remote
func (cmd *BuildCmd) IsRemote() bool {
	return true
}

// Help returns help for the build command
func (cmd *BuildCmd) Help() string {
	return "hap build\tRun the builds and commands from the Hapfile."
}

// Run the build command on the remote host
func (cmd *BuildCmd) Run(remote *hap.Remote) (string, error) {
	if result, err := Commands.Get("push").Run(remote); err != nil {
		return result, err
	}
	if err := remote.Build(); err != nil {
		result := fmt.Sprintf("[%s] build failed.", remote.Host.Name)
		return result, err
	}
	result := fmt.Sprintf("[%s] build completed.", remote.Host.Name)
	return result, nil
}
