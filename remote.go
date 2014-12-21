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

const happened string = "if [[ $(git rev-parse HEAD) = $(cat .happended) ]]; then echo \"[%s] Already completed. Commit again?\"; exit 2; fi"

type Remote struct {
	Git     Git
	Dir     string
	Config  SshConfig
	server  *Server
	session *ssh.Session
	b       bytes.Buffer
	mu      sync.Mutex
}

func NewRemote(server *Server) (*Remote, error) {
	config := SshConfig{
		Addr:     server.Addr,
		Username: server.Username,
		Identity: server.Identity,
		Password: server.Password,
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
	r := &Remote{Git: Git{Repo: repo}, Dir: dir, Config: config, server: server}
	return r, nil
}

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

func (r *Remote) Close() error {
	if r.session != nil {
		err := r.session.Close()
		r.session = nil
		return err
	}
	return nil
}

func (r *Remote) Initialize() ([]byte, error) {
	results := []byte{}
	if err := r.Connect(); err != nil {
		return results, err
	}
	env := fmt.Sprint(
		"GIT_DIR=", r.Dir, ";",
		"HAP_HOSTNAME=", r.server.Name, ";",
		"HAP_ADDR=", r.server.Addr, ";",
		"HAP_USER=", r.server.Username, ";",
	)
	commands := []string{
		fmt.Sprint(env),
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

func (r *Remote) Build(script string) ([]byte, error) {
	results, err := r.Push()
	if err != nil {
		return results, err
	}
	file, err := ioutil.ReadFile(script)
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
		fmt.Sprintf(happened, r.server.Name),
	}
	data = append(cmds, data...)
	data = append(data, "echo `git rev-parse HEAD` > .happended")
	return r.Execute(data)
}

func (r *Remote) Execute(commands []string) ([]byte, error) {
	results := []byte{}
	if err := r.Connect(); err != nil {
		return results, err
	}
	defer r.Close()
	r.session.Stdout = r
	r.session.Stderr = r
	cmd := commands[0]
	if len(commands) > 1 {
		cmd = fmt.Sprintf("sh -c '%s'", strings.Join(commands, "&&"))
	}
	err := r.session.Run(cmd)
	return r.b.Bytes(), err
}

func (r *Remote) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := []byte(fmt.Sprintf("[%s] ", r.server.Name))
	_, err := r.b.Write(append(name, p...))
	return len(p), err
}
