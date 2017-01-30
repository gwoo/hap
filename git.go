// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"fmt"
	"os"
	"os/exec"
)

// Git struct
type Git struct {
	Repo string
	Work string
	Key  string
}

// Exists checks whether the git executable exists
func (g Git) Exists() error {
	o, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("%s\n%s", o, err)
	}
	return nil
}

// Commit adds all files, including untracked to the repo
func (g Git) Commit(message string) ([]byte, error) {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = g.Work
	result, err := cmd.CombinedOutput()
	if err != nil {
		return result, err
	}
	cmd = exec.Command("git", "commit", "-q", "-m", message)
	cmd.Dir = g.Work
	return cmd.CombinedOutput()
}

// Branch creates a new branch
func (g Git) Branch(name string) ([]byte, error) {
	cmd := exec.Command("git", "branch", name)
	cmd.Dir = g.Work
	return cmd.CombinedOutput()
}

// Checkout switches the branch
func (g Git) Checkout(name string) ([]byte, error) {
	cmd := exec.Command("git", "checkout", name)
	cmd.Dir = g.Work
	return cmd.CombinedOutput()
}

// Push forces push to the branch to the remote repo
func (g Git) Push(branch string) ([]byte, error) {
	if branch == "" {
		branch = "master"
	}
	cmd := exec.Command("git", "push", "-f", g.Repo, branch)
	cmd.Dir = g.Work
	env := os.Environ()
	cmd.Env = append(env, "GIT_SSH=.git/hap-ssh")
	return cmd.CombinedOutput()
}

// RevParse returns current revision of HEAD
func (g Git) RevParse() ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.Work
	return cmd.CombinedOutput()
}

// UpdateSubmodules updates and initializes submodules
func (g Git) UpdateSubmodules() ([]byte, error) {
	cmd := exec.Command("git", "submodule", "update", "--init")
	cmd.Dir = g.Work
	return cmd.CombinedOutput()
}

// Add this hook to the remote repo
const postReceiveHook string = `cat > ".git/hooks/post-receive" << "EOF"
#!/bin/bash

test "${PWD%/.git}" != "$PWD" && cd ..
unset GIT_DIR GIT_WORK_TREE
read oldrev newrev ref
branch=${ref#refs/heads/}
git reset -q --hard
git checkout -q ${branch}
EOF`
