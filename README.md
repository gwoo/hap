# Hap - A simple and effective provisioner.

Hap helps manage build scripts with git and run them concurrently on multiple remote hosts using composable blocks.

First, `hap create` to setup a new local repo. Then add hosts to the generated Hapfile. Once hosts are in place, run `hap build` to execute the build blocks and commands specified in the Hapfile for each host. After `hap build` a .happened file is saved with the current sha of remote repo. To run `hap build` again a new commit is required or use the `--force` param.

To run arbitrary commands use `hap c`, and to execute individual scripts with `hap exec`.
`hap c` will not push the latest or run from the current directory or use the environment files, those must be added as part of the command.
`hap exec` will push the latest, use the current directory and the environment.

If you only have one host, just use the `default` section. Then the `-h,--host` flag while running `hap` is not necessary.

Make sure every build script is executable before committing to the local repo.

## Installation
#### via Go

	go get -u github.com/gwoo/hap/cmd/hap

#### Binaries

darwin/amd64

	curl -L -C - -O https://github.com/gwoo/hap/releases/download/v1.6/hap-darwin-amd64; chmod a+x hap-darwin-amd64

linux/amd64

	curl -L -C - -O https://github.com/gwoo/hap/releases/download/v1.6/hap-linux-amd64; chmod a+x hap-linux-amd64


## Basic Workflow
 - Run `hap create <name>`
 - Modify `Hapfile`
 - Run `hap -h <host> build`

## Environment Variables
Hap exports `HAP_HOSTNAME`, `HAP_USER`, `HAP_ADDR` for use in scripts. You can add your own by using the `env` section or the `env` statement in the `host` section.

## Hapfile
The Hapfile uses [git-config](http://git-scm.com/docs/git-config#_syntax) syntax. There are 5 sections, `default`, `host`, `build`, `include`, and `env`. The `default` section holds host config that will be applied to all hosts. The `host` section holds a named host config. A host config includes `addr`, `username`, `password`, `identity`, `build`, and `cmd`, `env`. Only `addr` is required. The `identity` should point to a local ssh private key that has access to the host via the authorized_keys. The `build` section holds mulitple cmds that could be applied to a host. Multiple `build`, `cmd`, or `env` are permitted for each host. In addition, an `include` section accepts multiple `path` statements and an `env` section accepts multiple `file` statements.

### sections
 - `host`: Holds the configuration for a machine
  - `addr`: the host:port of the remote machine
  - `username`: the name of the user to login and run commands
  - `password`: password for ssh password based authentication
  - `identity`: path to ssh private key for key based authentication
  - `build`: one or more groups of commands to run
  - `cmd`: one or more commands to run on a specific host
  - `env`: one or more environment files to apply to this host (can override env sections)
 - `build`: sets of commands to run
  - `cmd`: one or more commands to run
 - `default` : Holds the standard configurations that can be applied to all hosts
  - <same as host>
 - `include`: Allows other files to be included in the current configuration
  - `path`: a path to the Hapfile the hap
 - `env`: make variables available to the all commands
  - `file`: path to a file that can be sourced


## Example Hapfile
A default build is specified, so init.sh and update.sh are executed for each host.
Host `one` specifies two commands, notify.sh and cleanup.sh, to be run after the default build commands. For host `one`, the `HAP_HOSTNAME` will be `one`, the `HAP_USER` will be `root`, and the `HAP_ADDR` will be `10.0.20.10:22`. Host `two` specifies no commands, so only the default build will be applied. For host `two`, the `HAP_HOSTNAME` will be `two`, the `HAP_USER` will be `admin`, and the `HAP_ADDR` will be `10.0.20.11:22`.

	[default]
	username = "root"
	identity = "~/.ssh/id_rsa"
	build = "initialize" ; applied to all hosts

	[host "one"]
	addr = "10.0.20.10:22"
	cmd = "./notify.sh"
	cmd = "./cleanup.sh"

	[host "two"]
	username = "admin"
	identity = "~/.ssh/admin_rsa"
	addr = "10.0.20.11:22"

	[build "initialize"]
	cmd = ./init.sh
	cmd = ./update.sh
	cmd = echo "initialized"

## Usage
	Usage of ./bin/hap:
	  -f, --file="Hapfile": Location of a Hapfile.
	      --force=false: Force build even if it happened before.
	  -h, --host="": Individual host to use for commands.
	  -v, --verbose=false: Verbose flag to print command log.

	Available Commands:
	hap build		Run the builds and commands from the Hapfile.
	hap c <command>	Run an arbitrary command on the remote host.
	hap create <name>	Create a new Hapfile at <name>.
	hap exec <script>	Execute a script on the remote host.
	hap push		Push current repo to the remote.

## Advanced Usage
Sometimes you want to `build` more than one host. If the hosts follow a similar pattern
you can reference all the hosts with a `*`. For example, `app-01` and `app-02` are configured.
Then you can build both with `hap -h app-* build` or `hap -h a* build`.

Sometimes you have a lot of hosts that you want to manage in clusters. If you create multiple files
you can use the `--file` flag to specify the location of the config. The file can be named anything.
For example, `hap -f Appfile -h app* push`, will push all the app hosts in the Appfile.

## License
The BSD License http://opensource.org/licenses/bsd-license.php.
