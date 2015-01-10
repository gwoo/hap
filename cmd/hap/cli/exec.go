// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/gwoo/hap"
)

// Add the exec command
func init() {
	Commands.Add("exec", &ExecCmd{})
}

// Exec command
type ExecCmd struct{}

// Does this command expect a remote
func (cmd *ExecCmd) IsRemote() bool {
	return true
}

// Get help on the exec command
func (cmd *ExecCmd) Help() string {
	return "hap exec <script>\tExecute a script on the remote host."
}

// Execute a script from the repo on the remote host
func (cmd *ExecCmd) Run(remote *hap.Remote) (string, error) {
	args := flag.Args()
	if len(args) <= 1 {
		return "", fmt.Errorf("error: expects <script>")
	}
	if result, err := Commands.Get("push").Run(remote); err != nil {
		return result, err
	}
	ex := strings.Join(args[1:], " ")
	if err := remote.Execute([]string{"cd " + remote.Dir, "./" + ex}); err != nil {
		result := fmt.Sprintf("[%s] `%s` failed.", remote.Host.Name, args[1])
		return result, err
	}
	result := fmt.Sprintf("[%s] `%s` completed.", remote.Host.Name, args[1])
	return result, nil
}
