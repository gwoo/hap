// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"fmt"
	"os/exec"
)

// Git struct
type Git struct {
	Repo string
	Work string
}

// Exists checks whether the git executable exists
func (g Git) Exists() error {
	o, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("%s\n%s", o, err)
	}
	return nil
}

// Commit takes a commit message. It adds and commits all
// files, including untracked, to the repo
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

// Push takes a branch and force pushes it to the git remote
func (g Git) Push(branch string) ([]byte, error) {
	if branch == "" {
		branch = "master"
	}
	cmd := exec.Command("git", "push", "-f", "-q", g.Repo, branch)
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
