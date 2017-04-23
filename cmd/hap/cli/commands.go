// Hap - the simple and effective provisioner
// Copyright (c) 2017 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package cli

import (
	"github.com/gwoo/hap"
)

// List of registered commands
var Commands = make(commands)

type commands map[string]Command

// Add takes a name and a command and adds it to the list
func (c commands) Add(name string, cmd Command) {
	c[name] = cmd
}

// Get takes a name and returns the associated command
func (c commands) Get(name string) Command {
	if command, ok := Commands[name]; ok {
		return command
	}
	return nil
}

// Command interface
type Command interface {
	IsRemote() bool
	Help() string
	Run(*hap.Remote) (string, error)
}
