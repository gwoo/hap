// Hap - the simple and effective provisioner
// Copyright (c) 2017 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"gopkg.in/gcfg.v1"
)

// Formatted script that checks if the build happened.
const happened string = "if [ \"$(git rev-parse HEAD)\" = \"$(cat .happended)\" ]; then echo \"Already completed. Commit again?\"; exit 2; fi"

// Remote defines the remote machine to provision
type Remote struct {
	Git       Git
	Dir       string
	Host      *Host
	env       []string
	sshConfig SSHConfig
	session   *ssh.Session
	writer    io.Writer
}

// NewRemote constructs a new remote machine
func NewRemote(host *Host) (*Remote, error) {
	sshConfig := SSHConfig{
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
	dir := host.GetDir()
	repo := fmt.Sprintf("ssh://%s@%s/~/%s", host.Username, host.Addr, dir)
	r := &Remote{
		sshConfig: sshConfig,
		Git:       Git{Repo: repo, Key: sshConfig.Identity},
		Dir:       dir,
		Host:      host,
	}
	return r, nil
}

// Connect starts an ssh session to a remote machine
func (r *Remote) Connect() error {
	if r.session != nil {
		return nil
	}
	var client *ssh.Client
	var err error
	for _ = range r.sshConfig.ClientConfig.Auth {
		client, err = ssh.Dial("tcp", r.sshConfig.Addr, r.sshConfig.ClientConfig)
		if err != nil {
			if len(r.sshConfig.ClientConfig.Auth) == 1 {
				return fmt.Errorf("Failed to connect with username=%s identity=%s password=%s",
					r.sshConfig.Username, r.sshConfig.Identity, r.sshConfig.Password)
			}
			r.sshConfig.ClientConfig.Auth = r.sshConfig.ClientConfig.Auth[1:]
		}
	}
	if client == nil {
		return fmt.Errorf("Failed to connect to %s", r.Host.Addr)
	}
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to create session: %s", err)
	}
	r.session = session
	return nil
}

// Close ends an ssh session with a remote machine
func (r *Remote) Close() error {
	if r.session != nil {
		err := r.session.Close()
		r.session = nil
		return err
	}
	return nil
}

// Initialize sets up a git repo on the remote machine
func (r *Remote) Initialize() error {
	if err := r.Connect(); err != nil {
		return err
	}
	var b = "#!/bin/sh\n\nssh "
	if r.sshConfig.Identity != "" {
		b = b + " -i " + r.sshConfig.Identity
	}
	b = b + " -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -o VerifyHostKeyDNS=no -o CheckHostIP=no $@"
	err := ioutil.WriteFile("/tmp/hap-ssh-"+r.Host.Name, []byte(b), 0777)
	if err != nil {
		return fmt.Errorf("1 %s\n", err)
	}
	commands := []string{
		fmt.Sprintf("GIT_DIR=\"%s\"", r.Dir),
		fmt.Sprint("mkdir -p $GIT_DIR"),
		fmt.Sprint("cd $GIT_DIR"),
		fmt.Sprint("git init -q"),
		fmt.Sprint("git config receive.denyCurrentBranch ignore"),
		fmt.Sprint("touch .git/hooks/post-receive"),
		fmt.Sprint("chmod a+x .git/hooks/post-receive"),
		fmt.Sprint(postReceiveHook),
	}
	return r.Execute(commands)
}

// Push updates the repo on the remote machine
func (r *Remote) Push() error {
	if err := r.Connect(); err != nil {
		return err
	}
	b, err := r.Git.RevParse()
	if err != nil {
		return fmt.Errorf("%s\n%s\n", string(b), err)
	}
	branch := strings.TrimSpace(string(b))
	if branch == "HEAD" {
		branch = fmt.Sprintf("%s:refs/heads/happened", branch)
	}
	if err := r.Initialize(); err != nil {
		return fmt.Errorf("%s\n", err)
	}
	os.Setenv("GIT_SSH", "/tmp/hap-ssh-"+r.Host.Name)

	if output, err := r.Git.Push(branch); err != nil {
		return fmt.Errorf("%s\n%s\n", string(output), err)
	}
	return nil
}

// PushSubmodules runs Push() to put submodules
// into the proper location on the remote machine
func (r *Remote) PushSubmodules() error {
	var modules struct {
		Submodules map[string]*submodule `gcfg:"submodule"`
	}
	if err := gcfg.ReadFileInto(&modules, ".gitmodules"); err != nil {
		return nil
	}
	if output, err := r.Git.UpdateSubmodules(); err != nil {
		return fmt.Errorf("%s\n%s\n", string(output), err)
	}
	var wg sync.WaitGroup
	errors := []string{}
	for _, module := range modules.Submodules {
		wg.Add(1)
		go func(module *submodule) {
			defer wg.Done()
			sr := &Remote{
				sshConfig: r.sshConfig,
				Dir:       filepath.Join(r.Dir, module.Path),
				Host:      r.Host,
				Git: Git{
					Repo: fmt.Sprint(r.Git.Repo, "/", module.Path),
					Work: module.Path,
				},
			}
			if err := sr.Push(); err != nil {
				errors = append(errors, fmt.Sprintf("[%s] %s", module.Path, err))
			}
		}(module)
	}
	wg.Wait()
	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "\n"))
	}
	return nil
}

// Build executes the builds and cmds
// First execute builds specified in Hapfile
// Then execute any cmds specified in Hapfile
func (r *Remote) Build(force bool) error {
	cmds := []string{"cd " + r.Dir, "touch .happended"}
	if !force {
		cmds = append(cmds, happened)
	}
	cmds = r.Host.AddEnv(cmds)
	cmds = append(cmds, r.Host.Cmds()...)
	cmds = append(cmds, "echo `git rev-parse HEAD` > .happended")
	return r.Execute(cmds)
}

// Execute will shell out to run one or more commands
func (r *Remote) Execute(commands []string) error {
	if err := r.Connect(); err != nil {
		return err
	}
	defer r.Close()
	r.session.Stdout = NewRemoteWriter(r.Host.Name, os.Stdout)
	r.session.Stderr = NewRemoteWriter(r.Host.Name, os.Stderr)
	cmd := fmt.Sprint(r.Env(), commands[0])
	if len(commands) > 1 {
		cmd = fmt.Sprintf("sh -c '%s%s'", r.Env(), strings.Join(commands, "&&"))
	}
	if err := r.session.Run(cmd); err != nil {
		return fmt.Errorf("[%s] %s", r.Host.Name, err)
	}
	return nil
}

// Env returns the preset environment variables to pass to execute
func (r *Remote) Env() string {
	host, port, _ := net.SplitHostPort(r.Host.Addr)
	return fmt.Sprint(
		"export HAP_HOSTNAME=\"", r.Host.Name, "\";",
		"export HAP_ADDR=\"", r.Host.Addr, "\";",
		"export HAP_USER=\"", r.Host.Username, "\";",
		"export HAP_IP=\"", host, "\";",
		"export HAP_PORT=\"", port, "\";",
	)
}

// NewRemoteWriter returns a Writer with [host] prepended to the output
func NewRemoteWriter(host string, w io.Writer) io.Writer {
	return &RemoteWriter{host: host, w: w}
}

// RemoteWriter is a Writer with host and io.Writer
type RemoteWriter struct {
	host string
	w    io.Writer
}

// Write implements the io.Writer interface
func (hw *RemoteWriter) Write(p []byte) (int, error) {
	var err error
	l := len(p)
	scanner := bufio.NewScanner(bytes.NewReader(p))
	for scanner.Scan() {
		_, err = fmt.Fprintf(hw.w, "[%s] %s\n", hw.host, scanner.Bytes())
	}
	if err != nil {
		return l, err
	}
	if err := scanner.Err(); err != nil {
		return l, err
	}
	return l, nil
}

type submodule struct {
	Path string
	URL  string
}
