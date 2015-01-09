// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"code.google.com/p/gcfg"
	"golang.org/x/crypto/ssh"
)

// Formatted script that checks if the build happened.
const happened string = "if [[ $(git rev-parse HEAD) = $(cat .happended) ]]; then echo \"Already completed. Commit again?\"; exit 2; fi"

// The remote machine to provision
type Remote struct {
	Git       Git
	Dir       string
	Host      *Host
	sshConfig SshConfig
	session   *ssh.Session
	b         bytes.Buffer
	mu        sync.Mutex
}

// Construct a new remote machine
func NewRemote(host *Host) (*Remote, error) {
	sshConfig := SshConfig{
		Addr:     host.Addr,
		Username: host.Username,
		Identity: host.Identity,
		Password: host.Password,
	}
	clientConfig, err := NewClientConfig(sshConfig)
	if err != nil {
		return nil, err
	}
	sshConfig.ClientConfig = clientConfig
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	dir := filepath.Base(cwd)
	repo := fmt.Sprintf("ssh://%s@%s/~/%s", host.Username, host.Addr, dir)
	r := &Remote{Git: Git{Repo: repo}, Dir: dir, sshConfig: sshConfig, Host: host}
	return r, nil
}

// Start ssh session to remote machine.
func (r *Remote) Connect() error {
	if r.session != nil {
		return nil
	}
	client, err := ssh.Dial("tcp", r.sshConfig.Addr, r.sshConfig.ClientConfig)
	if err != nil {
		return err
	}
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	r.session = session
	return nil
}

// End session with remote machine
func (r *Remote) Close() error {
	if r.session != nil {
		err := r.session.Close()
		r.session = nil
		return err
	}
	return nil
}

// Setup a git repo on the remote machine
func (r *Remote) Initialize() ([]byte, error) {
	results := []byte{}
	if err := r.Connect(); err != nil {
		return results, err
	}
	commands := []string{
		fmt.Sprintf("GIT_DIR=\"%s\"", r.Dir),
		fmt.Sprint("mkdir -p $GIT_DIR"),
		fmt.Sprint("cd $GIT_DIR"),
		fmt.Sprint("git init"),
		fmt.Sprint("git config receive.denyCurrentBranch ignore"),
		fmt.Sprint("touch .git/hooks/post-receive"),
		fmt.Sprint("chmod a+x .git/hooks/post-receive"),
		fmt.Sprint(postReceiveHook),
	}
	return r.Execute(commands)
}

// Update repo on the remote machine
func (r *Remote) Push() ([]byte, error) {
	results := []byte{}
	if err := r.Connect(); err != nil {
		return results, err
	}
	key, err := NewKeyFile(r.sshConfig.Identity)
	if err != nil {
		return results, err
	}
	cmd := exec.Command("ssh-add", key)
	results, err = cmd.CombinedOutput()
	if err != nil {
		return results, err
	}
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = r.Git.Work
	b, err := cmd.CombinedOutput()
	if err != nil {
		return results, err
	}
	branch := strings.TrimSpace(string(b))
	if branch == "HEAD" {
		branch = fmt.Sprintf("%s:refs/heads/happened", branch)
	}
	return r.Git.Push(branch)
}

// Initialize and Push submodules into proper location on remote
func (r *Remote) PushSubmodules() ([]byte, error) {
	results := []byte{}
	var modules struct {
		Submodules map[string]*struct {
			Path string
			Url  string
		} `gcfg:"submodule"`
	}
	err := gcfg.ReadFileInto(&modules, ".gitmodules")
	if err != nil {
		return results, err
	}
	errors := []string{}
	for _, module := range modules.Submodules {
		sr := &Remote{
			Dir:       filepath.Join(r.Dir, module.Path),
			sshConfig: r.sshConfig,
			Host:      r.Host,
			Git: Git{
				Repo: fmt.Sprint(r.Git.Repo, "/", module.Path),
				Work: module.Path,
			},
		}
		_, err := sr.Initialize()
		if err != nil {
			errors = append(errors, fmt.Sprintf("[%s] %s", module.Path, err.Error()))
		}
		r, err := sr.Push()
		results = append(results, r...)
		if err != nil {
			errors = append(errors, fmt.Sprintf("[%s] %s", module.Path, err.Error()))
		}
	}
	if len(errors) > 0 {
		return results, fmt.Errorf("%s", strings.Join(errors, "\n"))
	}
	return results, nil
}

// Execute the builds and cmds
// First execute builds specified in Hapfile
// Then execute any cmds specified in Hapfile
func (r *Remote) Build() ([]byte, error) {
	results, err := r.Push()
	if err != nil {
		return results, err
	}
	cmds := []string{
		"cd " + r.Dir,
		"touch .happended",
		happened,
	}
	cmds = append(cmds, r.Host.Cmds()...)
	cmds = append(cmds, "echo `git rev-parse HEAD` > .happended")
	return r.Execute(cmds)
}

// Shell out to the multiple commands or run one
func (r *Remote) Execute(commands []string) ([]byte, error) {
	results := []byte{}
	if err := r.Connect(); err != nil {
		return results, err
	}
	defer r.Close()
	r.session.Stdout = r
	r.session.Stderr = r
	cmd := fmt.Sprintf("%s%s", r.Env(), commands[0])
	if len(commands) > 1 {
		cmd = fmt.Sprintf("sh -c '%s%s'", r.Env(), strings.Join(commands, "&&"))
	}
	err := r.session.Run(cmd)
	return r.b.Bytes(), err
}

// Return preset environment variables to pass to execute
func (r *Remote) Env() string {
	return fmt.Sprint(
		"export HAP_HOSTNAME=\"", r.Host.Name, "\";",
		"export HAP_ADDR=\"", r.Host.Addr, "\";",
		"export HAP_USER=\"", r.Host.Username, "\";",
	)
}

// Implement io.Writer for printing messages from remote.
func (r *Remote) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := []byte(fmt.Sprintf("[%s] ", r.Host.Name))
	_, err := r.b.Write(append(name, p...))
	return len(p), err
}
