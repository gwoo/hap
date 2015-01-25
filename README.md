# Hap - A simple and effective provisioner.

Hap helps manage build scripts with git and run them concurrently on multiple remote hosts using composable blocks.

First, `hap create` to setup a new local repo. Then add hosts to the generated Hapfile.Once hosts are in place, `hap init` will setup the remote hosts. Finally, `hap build` will execute the build blocks and commands specified in the Hapfile for each host. After `hap build` a .happened file is saved with the current sha of remote repo. To run `hap build` again a new commit is required.

Tun arbitrary commands use `hap c`, and to execute individual scripts with `hap exec`.

If you only have one host, just use the `default` section. Then the `-all` or `-host` flag while running `hap` is not necessary.

Make sure every build script is executable before committing to the local repo.

## Installation
#### via Go

	go get github.com/gwoo/hap/cmd/hap

#### Binaries

darwin/amd64

	curl -L -C - -O https://github.com/gwoo/hap/releases/download/v1.4/hap-darwin-amd64; chmod a+x hap-darwin-amd64

linux/amd64

	curl -L -C - -O https://github.com/gwoo/hap/releases/download/v1.4/hap-linux-amd64; chmod a+x hap-linux-amd64


## Basic Workflow
 - Run `hap create <name>`
 - Modify `Hapfile`
 - Run `hap init` and `hap build`

## Environment Variables
Hap exports `HAP_HOSTNAME`, `HAP_USER`, `HAP_ADDR` for use in scripts.

## Hapfile
The Hapfile uses [git-config](http://git-scm.com/docs/git-config#_syntax) syntax. There are 3 sections, `default`, `host`, and `build`.
The `default` section holds host config that will be applied to all hosts.
The `host` section holds a named host config. A host config includes `addr`, `username`, `password`, `identity`, `build`, and `cmd`. Only `addr` is required. The `identity` should point to a local ssh private key that has access to the host via the authorized_keys. The `build` section holds mulitple cmds that could be applied to a host. Multiple `build` and `cmd` are permitted for each host.

## Example Hapfile
A default build is specified, so init.sh and update.sh are executed for each host.
Host one specifies two commands, notify.sh and cleanup.sh, to be run after the default build commands.

	[default]
	username = "root"
	identity = "~/.ssh/id_rsa"
	build = "default" ; applied to all hosts

	[host "one"]
	addr = "10.0.20.10:22"
	cmd = "./notify.sh"
	cmd = "./cleanup.sh"

	[host "two"]
	addr = "10.0.20.11:22"

	[build "default"]
	cmd = ./init.sh
	cmd = ./update.sh


## Usage
	Usage of hap:
	  -a, --all=false: Use ALL the hosts.
	  -h, --host="": Individual host to use for commands.
	  -v, --verbose=false: Verbose flag to print command log.

	Available Commands:
	hap build			Run the builds and commands from the Hapfile.
	hap c <command>		Run an arbitrary command on the remote host.
	hap create <name>	Create a new Hapfile at <name>.
	hap exec <script>	Execute a script on the remote host.
	hap init			Initialize a new remote host.
	hap push			Push current repo to the remote.

## License
The BSD License http://opensource.org/licenses/bsd-license.php.
