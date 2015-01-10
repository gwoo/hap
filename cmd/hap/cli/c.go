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

// Load all the available commands
func init() {
	Commands.Add("c", &ArbitraryCmd{})
}

type ArbitraryCmd struct {
	result []byte
	log    string
}

// Does this command expect a remote
func (cmd *ArbitraryCmd) IsRemote() bool {
	return true
}

// Get help on c (arbitrary) command
func (cmd *ArbitraryCmd) Help() string {
	return "hap c <command>\tRun an arbitrary command on the remote host."
}

// Run an arbitrary command on the remote host
func (cmd *ArbitraryCmd) Run(remote *hap.Remote) (string, error) {
	args := flag.Args()
	if len(args) <= 1 {
		return "", fmt.Errorf("error: expects <command>")
	}
	arbitrary := strings.Join(args[1:], " ")
	if err := remote.Execute([]string{arbitrary}); err != nil {
		result := fmt.Sprintf("[%s] `%s` failed.", remote.Host.Name, arbitrary)
		return result, err
	}
	result := fmt.Sprintf("[%s] `%s` completed.", remote.Host.Name, arbitrary)
	return result, nil
}
