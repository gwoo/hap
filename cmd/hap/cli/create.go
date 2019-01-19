// Hap - the simple and effective provisioner
// Copyright (c) 2019 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"

	"github.com/gwoo/hap"
	flag "github.com/ogier/pflag"
)

// Add the create command
func init() {
	Commands.Add("create", &CreateCmd{})
}

// CreateCmd inits the command for setting up remote repo
type CreateCmd struct{}

// IsRemote returns whether the command expects a remote or not
func (cmd *CreateCmd) IsRemote() bool {
	return false
}

// Help returns the help on hap create <name>
func (cmd *CreateCmd) Help() string {
	return "hap create <name>\tCreate a new Hapfile at <name>."
}

// Run the command against the remote
func (cmd *CreateCmd) Run(remote *hap.Remote) (string, error) {
	var err error
	args := flag.Args()
	if len(args) <= 1 {
		return "", fmt.Errorf("error: expects <name>")
	}
	work := flag.Arg(1)
	if err = os.MkdirAll(work, os.ModePerm|os.ModeDir); err != nil {
		result := fmt.Sprintf("create %s failed.", work)
		return result, err
	}
	c := exec.Command("git", "init", ".")
	c.Dir = work
	if result, err := c.CombinedOutput(); err != nil {
		result := fmt.Sprint(string(result))
		return result, err
	}
	file, err := os.OpenFile(work+"/Hapfile", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		result := fmt.Sprintf("create %s failed.", work+"/Hapfile")
		return result, err
	}
	file.WriteString("[default]\n")
	if u, err := user.Current(); err == nil {
		file.WriteString(fmt.Sprintf("username = \"%s\"\n", u.Username))
	}
	if _, err = hap.NewKeyFile("~/.ssh/id_rsa"); err == nil {
		file.WriteString(fmt.Sprintf("identity = \"%s\"\n", "~/.ssh/id_rsa"))
	}
	if err != nil {
		result := fmt.Sprintf("create %s failed.", work)
		return result, err
	}
	result := fmt.Sprintf("create %s completed.", work)
	return result, nil
}
