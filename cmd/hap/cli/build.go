// Hap - the simple and effective provisioner
// Copyright (c) 2017 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package cli

import (
	"fmt"

	"github.com/gwoo/hap"
	flag "github.com/ogier/pflag"
)

var force = flag.BoolP("force", "", false, "Force build even if it happened before.")
var dry = flag.BoolP("dry", "", false, "Show commands without running them.")

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
	if *dry {
		result := fmt.Sprintf(
			"[%s] --dry run.\n",
			remote.Host.Name,
		)
		cmds := []string{"cd " + remote.Dir,}
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
	if err := remote.Build(*force); err != nil {
		result := fmt.Sprintf("[%s] build failed.", remote.Host.Name)
		return result, err
	}
	result := fmt.Sprintf("[%s] build completed.", remote.Host.Name)
	return result, nil
}
