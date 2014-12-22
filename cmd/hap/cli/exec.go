// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
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
type ExecCmd struct {
	result []byte
	log    string
}

// Return the result of the command
func (cmd *ExecCmd) String() string {
	return string(cmd.result)
}

// Return the log generated by the command
func (cmd *ExecCmd) Log() string {
	return cmd.log
}

// Get help on the exec command
func (cmd *ExecCmd) Help() string {
	return "hap exec <script>\tExecute a script on the remote host."
}

// Execute a script from the repo on the remote host
func (cmd *ExecCmd) Run(remote *hap.Remote) error {
	args := flag.Args()
	if len(args) <= 1 {
		return fmt.Errorf("%s", cmd.Help())
	}
	push := Commands.Get("push")
	if err := push.Run(remote); err != nil {
		cmd.log = push.Log()
		cmd.result = []byte(push.String())
		return err
	}
	ex := strings.Join(args[1:], " ")
	result, err := remote.Execute([]string{"cd " + remote.Dir, "./" + ex})
	cmd.result = result
	cmd.log = fmt.Sprintf("Executed `%s` on %s.", args[1], remote.Host.Addr)
	return err
}
