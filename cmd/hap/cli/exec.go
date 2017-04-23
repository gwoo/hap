// Hap - the simple and effective provisioner
// Copyright (c) 2017 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package cli

import (
	"fmt"
	"strings"

	"github.com/gwoo/hap"
	flag "github.com/ogier/pflag"
)

// Add the exec command
func init() {
	Commands.Add("exec", &ExecCmd{})
}

// ExecCmd is the command
type ExecCmd struct{}

// IsRemote returns whether the command expects a remote or not
func (cmd *ExecCmd) IsRemote() bool {
	return true
}

// Help returns help on hap exec <script>
func (cmd *ExecCmd) Help() string {
	return "hap exec <script>\tExecute a script on the remote host."
}

// Run takes a remote and executes a script from the repo on it
func (cmd *ExecCmd) Run(remote *hap.Remote) (string, error) {
	args := flag.Args()
	if len(args) <= 1 {
		return "", fmt.Errorf("error: expects <script>")
	}
	if result, err := Commands.Get("push").Run(remote); err != nil {
		return result, err
	}
	ex := strings.Join(args[1:], " ")
	cmds := []string{"cd " + remote.Dir}
	cmds = remote.Host.AddEnv(cmds)
	cmds = append(cmds, ex)
	if err := remote.Execute(cmds); err != nil {
		result := fmt.Sprintf("[%s] `%s` failed.", remote.Host.Name, args[1])
		return result, err
	}
	result := fmt.Sprintf("[%s] `%s` completed.", remote.Host.Name, args[1])
	return result, nil
}
