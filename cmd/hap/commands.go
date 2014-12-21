// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/gwoo/hap"
)

func init() {
	commands.Add("init", &InitCmd{})
	commands.Add("push", &PushCmd{})
	commands.Add("c", &ArbitraryCmd{})
	commands.Add("exec", &ExecCmd{})
	commands.Add("build", &BuildCmd{})
}

var commands = make(Commands)

type Commands map[string]Command

func (c Commands) Add(name string, cmd Command) {
	c[name] = cmd
}

func (c Commands) Get(name string) Command {
	if strings.Contains(name, ".json") {
		command := commands["build"].(*BuildCmd)
		command.build = name
		return command
	}
	if command, ok := commands[name]; ok {
		return command
	}
	return nil
}

type Command interface {
	Help() string
	Run(*hap.Remote) error
	String() string
	Log() string
}

type InitCmd struct {
	result []byte
	log    string
}

func (cmd *InitCmd) String() string {
	return string(cmd.result)
}

func (cmd *InitCmd) Log() string {
	return cmd.log
}

func (cmd *InitCmd) Help() string {
	return "Initialize a new remote host."
}

func (cmd *InitCmd) Run(remote *hap.Remote) error {
	result, err := remote.Initialize()
	cmd.result = result
	if err != nil {
		cmd.log = fmt.Sprintf("%s failed to initialize on %s.", remote.Dir, remote.Config.Addr)
		return err
	}
	if len(cmd.result) <= 0 {
		cmd.result = []byte(fmt.Sprintf("[%s] Init successful.\n", remote.Host.Name))
	}
	cmd.log = fmt.Sprintf("%s initialized on %s.", remote.Dir, remote.Config.Addr)
	return nil
}

type PushCmd struct {
	result []byte
	log    string
}

func (cmd *PushCmd) String() string {
	return string(cmd.result)
}

func (cmd *PushCmd) Log() string {
	return cmd.log
}

func (cmd *PushCmd) Help() string {
	return "Push current repo to the remote."
}

func (cmd *PushCmd) Run(remote *hap.Remote) error {
	result, err := remote.Push()
	cmd.result = result
	if err != nil {
		cmd.log = fmt.Sprintf("Failed to push to %s.", remote.Config.Addr)
		return err
	}
	if len(cmd.result) <= 0 {
		cmd.result = []byte(fmt.Sprintf("[%s] Push successful.\n", remote.Host.Name))
	}
	cmd.log = fmt.Sprintf("Pushed to %s.", remote.Config.Addr)
	return nil
}

type ArbitraryCmd struct {
	result []byte
	log    string
}

func (cmd *ArbitraryCmd) String() string {
	return string(cmd.result)
}

func (cmd *ArbitraryCmd) Log() string {
	return cmd.log
}

func (cmd *ArbitraryCmd) Help() string {
	return "Run an arbitrary command on the remote."
}

func (cmd *ArbitraryCmd) Run(remote *hap.Remote) error {
	args := flag.Args()
	if len(args) <= 1 {
		return fmt.Errorf("%s", cmd.Help())
	}
	arbitrary := strings.Join(args[1:], " ")
	result, err := remote.Execute([]string{arbitrary})
	cmd.result = result
	cmd.log = fmt.Sprintf("Executed `%s` on %s.", arbitrary, remote.Config.Addr)
	return err
}

type ExecCmd struct {
	result []byte
	log    string
}

func (cmd *ExecCmd) String() string {
	return string(cmd.result)
}

func (cmd *ExecCmd) Log() string {
	return cmd.log
}

func (cmd *ExecCmd) Help() string {
	return "Execute a script on the remote host."
}

func (cmd *ExecCmd) Run(remote *hap.Remote) error {
	args := flag.Args()
	if len(args) <= 1 {
		return fmt.Errorf("%s", cmd.Help())
	}
	push := commands.Get("push")
	if err := push.Run(remote); err != nil {
		cmd.log = push.Log()
		cmd.result = []byte(push.String())
		return err
	}
	ex := strings.Join(args[1:], " ")
	result, err := remote.Execute([]string{"cd " + remote.Dir, "./" + ex})
	cmd.result = result
	cmd.log = fmt.Sprintf("Executed `%s` on %s.", args[1], remote.Config.Addr)
	return err
}

type BuildCmd struct {
	result []byte
	log    string
	build  string
}

func (cmd *BuildCmd) String() string {
	return string(cmd.result)
}

func (cmd *BuildCmd) Log() string {
	return cmd.log
}

func (cmd *BuildCmd) Help() string {
	return fmt.Sprintf("%s\n%s",
		"hap <build.json> : Run the scripts provided by the json formatted build list.",
	)
}

func (cmd *BuildCmd) Run(remote *hap.Remote) error {
	if cmd.build == "" {
		return fmt.Errorf("%s", cmd.Help())
	}
	result, err := remote.Build(cmd.build)
	cmd.result = result
	if err != nil {
		cmd.log = fmt.Sprintf("Build failed %s on %s.", cmd.build, remote.Config.Addr)
		return err
	}
	cmd.log = fmt.Sprintf("Build %s on %s.", cmd.build, remote.Config.Addr)
	return nil
}
