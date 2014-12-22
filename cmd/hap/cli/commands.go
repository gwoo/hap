// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
package cli

import (
	"github.com/gwoo/hap"
)

// List of registered commands
var Commands = make(commands)

type commands map[string]Command

// Add command to list
func (c commands) Add(name string, cmd Command) {
	c[name] = cmd
}

// Get command from registered list
func (c commands) Get(name string) Command {
	if command, ok := Commands[name]; ok {
		return command
	}
	return nil
}

// Command interface
type Command interface {
	Help() string
	Run(*hap.Remote) error
	String() string
	Log() string
}
