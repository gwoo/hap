// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
package hap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

// Formatted script that checks if the build happened.
const happened string = "if [[ $(git rev-parse HEAD) = $(cat .happended) ]]; then echo \"Already completed. Commit again?\"; exit 2; fi"

// The remote machine to provision
type Remote struct {
	Git     Git
	Dir     string
	Config  SshConfig
	host    *Host
	session *ssh.Session
	b       bytes.Buffer
	mu      sync.Mutex
}

// Construct a new remote machine
func NewRemote(host *Host) (*Remote, error) {
	config := SshConfig{
		Addr:     host.Addr,
		Username: host.Username,
		Identity: host.Identity,
		Password: host.Password,
	}
	cfg, err := NewClientConfig(config)
	if err != nil {
		return nil, err
	}
	config.ClientConfig = cfg
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	dir := filepath.Base(cwd)
	repo := fmt.Sprintf("ssh://%s@%s/~/%s", config.Username, config.Addr, dir)
	r := &Remote{Git: Git{Repo: repo}, Dir: dir, Config: config, host: host}
	return r, nil
}

// Start ssh session to remote machine.
func (r *Remote) Connect() error {
	if r.session != nil {
		return nil
	}
	client, err := ssh.Dial("tcp", r.Config.Addr, r.Config.ClientConfig)
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
	key, err := NewKeyFile(r.Config.Identity)
	if err != nil {
		return results, err
	}
	cmd := exec.Command("ssh-add", key)
	results, err = cmd.CombinedOutput()
	if err != nil {
		return results, err
	}
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branch, err := cmd.CombinedOutput()
	if err != nil {
		return results, err
	}
	return r.Git.Push(strings.TrimSpace(string(branch)))
}

// Read a json file for a list of commands to execute.
func (r *Remote) Build(list string) ([]byte, error) {
	results, err := r.Push()
	if err != nil {
		return results, err
	}
	file, err := ioutil.ReadFile(list)
	if err != nil {
		return file, err
	}
	var data []string
	err = json.Unmarshal(file, &data)
	if err != nil {
		return file, err
	}
	cmds := []string{
		"cd " + r.Dir,
		"touch .happended",
		happened,
	}
	data = append(cmds, data...)
	data = append(data, "echo `git rev-parse HEAD` > .happended")
	return r.Execute(data)
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
		"export HAP_HOSTNAME=\"", r.host.Name, "\";",
		"export HAP_ADDR=\"", r.host.Addr, "\";",
		"export HAP_USER=\"", r.host.Username, "\";",
	)
}

// Implement io.Writer for printing messages from remote.
func (r *Remote) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := []byte(fmt.Sprintf("[%s] ", r.host.Name))
	_, err := r.b.Write(append(name, p...))
	return len(p), err
}
