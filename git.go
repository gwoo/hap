package hap

import (
	"fmt"
	"os/exec"
)

type Git struct {
	Repo string
	Work string
}

func (g Git) Exists() error {
	o, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("%s\n%s", o, err)
	}
	return nil
}

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

func (g Git) Push(branch string) ([]byte, error) {
	if branch == "" {
		branch = "master"
	}
	cmd := exec.Command("git", "push", "-f", "-q", g.Repo, branch)
	cmd.Dir = g.Work
	return cmd.CombinedOutput()
}

const postReceiveHook string = `cat > ".git/hooks/post-receive" << "EOF"
#!/bin/bash

test "${PWD%/.git}" != "$PWD" && cd ..
unset GIT_DIR GIT_WORK_TREE
read oldrev newrev ref
branch=${ref#refs/heads/}
git reset --hard
git checkout ${branch}
EOF`
