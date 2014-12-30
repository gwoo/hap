# Hap - A simple and effective provisioner.

Hap uses Git to manage build scripts and run them on the remote server.
Hap init will setup a remote host. Hap build will execute the commands specified in the Hapfile. To run arbitrary commands use hap c, and to execute individual scripts use hap exec. After `hap build` a .happened file is saved with the current sha of remote repo. To build again a new commit is required. Make sure every build script is executable before being committed to the local repo.

## Basic Workflow
 - Run `hap create <name>`
 - Modify `Hapfile`
 - Run `hap init` and `hap build`

## Environment Variables
Hap exports `HAP_HOSTNAME`, `HAP_USER`, `HAP_ADDR` for use in your scripts.

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
	  -host="": Individual host to use for commands.
	  -v=false: Verbose flag to print command log.

	Available Commands:
	hap build		Run the builds and commands from the Hapfile.
	hap c <command>		Run an arbitrary command on the remote host.
	hap create <name>	Create a new Hapfile at <name>.
	hap exec <script>	Execute a script on the remote host.
	hap init		Initialize a new remote host.
	hap push		Push current repo to the remote.

## License
The BSD License http://opensource.org/licenses/bsd-license.php.
