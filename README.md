# Hap - A simple and effective provisioner.

Hap uses Git to manage build scripts and run them on the remote server.
Hap init will setup a remote host. Hap build will execute the commands specified in the Hapfile. To run arbitrary commands use hap c, and to execute individual scripts use hap exec.

## Basic Workflow
 - Create a local repository to hold builds scripts
 - Add `Hapfile` to local repository
 - Run `hap init` and `hap build`

## Hapfile
The Hapfile uses [git-config](http://git-scm.com/docs/git-config#_syntax) syntax. Thre are 3 sections, `default`, `host`, `build`.
The `default` section holds host config that will be applied to all hosts
The `host` section holds a named host config. A host config includes `addr`, `username`, `password`, `identity`, `build`, and `cmd`. Only `addr` is required. The `identity` should point to an ssh private key. Multiple `build` and `cmd` can be used.
The `build` section holds cmds that could be applied to hosts

## Example Hapfile
This example does not specify a `username`, `password` or `identity` so the details as specified in ~/.ssh/config are used to connect to hosts.
A default build is specified, so update.sh and build.sh are executed for each host.
Host one specifies two commands, notify.sh and cleanup.sh, to be run after the default build commands.

	[default]
	build = "default" ; applied to all hosts

	[host "one"]
	addr = "10.0.20.10:22"
	cmd = "./notify.sh"
	cmd = "./cleanup.sh"

	[host "two"]
	addr = "10.0.20.11:22"

	[build "default"]
	cmd = ./update.sh
	cmd = ./default.sh


## Usage
	Usage of hap:
	  -all=false: Use ALL the hosts.
	  -h="": Individual host to use for commands.
		 If empty, the default or a random host is used.
	  -v=false: Verbose flag to print output

	Available Commands:
	hap build			Run the builds and commands from the Hapfile.
	hap c <command>		Run an arbitrary command on the remote host.
	hap exec <script>	Execute a script on the remote host.
	hap init			Initialize a new remote host.
	hap push			Push current repo to the remote.

## License
The BSD License http://opensource.org/licenses/bsd-license.php.
