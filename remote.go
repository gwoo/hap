// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	writer    io.Writer
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
	r := &Remote{
		sshConfig: sshConfig,
		Git:       Git{Repo: repo},
		Dir:       dir,
		Host:      host,
	}
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
func (r *Remote) Initialize() error {
	if err := r.Connect(); err != nil {
		return err
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

// Update repo on the remote machine
func (r *Remote) Push() error {
	if err := r.Connect(); err != nil {
		return err
	}
	key, err := NewKeyFile(r.sshConfig.Identity)
	if err != nil {
		return err
	}
	cmd := exec.Command("ssh-add", key)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = r.Git.Work
	b, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	branch := strings.TrimSpace(string(b))
	if branch == "HEAD" {
		branch = fmt.Sprintf("%s:refs/heads/happened", branch)
	}
	if output, err := r.Git.Push(branch); err != nil {
		return fmt.Errorf("%s\n%s", string(output), err)
	}
	return nil
}

// Initialize and Push submodules into proper location on remote
func (r *Remote) PushSubmodules() error {
	var modules struct {
		Submodules map[string]*struct {
			Path string
			Url  string
		} `gcfg:"submodule"`
	}
	if err := gcfg.ReadFileInto(&modules, ".gitmodules"); err != nil {
		return err
	}
	errors := []string{}
	for _, module := range modules.Submodules {
		sr := &Remote{
			sshConfig: r.sshConfig,
			Dir:       filepath.Join(r.Dir, module.Path),
			Host:      r.Host,
			Git: Git{
				Repo: fmt.Sprint(r.Git.Repo, "/", module.Path),
				Work: module.Path,
			},
		}
		if err := sr.Initialize(); err != nil {
			errors = append(errors, fmt.Sprintf("[%s] %s", module.Path, err))
		}
		if err := sr.Push(); err != nil {
			errors = append(errors, fmt.Sprintf("[%s] %s", module.Path, err))
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "\n"))
	}
	return nil
}

// Execute the builds and cmds
// First execute builds specified in Hapfile
// Then execute any cmds specified in Hapfile
func (r *Remote) Build() error {
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
func (r *Remote) Execute(commands []string) error {
	if err := r.Connect(); err != nil {
		return err
	}
	defer r.Close()
	r.session.Stdout = NewRemoteWriter(r.Host.Name, os.Stdout)
	r.session.Stderr = NewRemoteWriter(r.Host.Name, os.Stderr)
	cmd := fmt.Sprintf("%s%s", r.Env(), commands[0])
	if len(commands) > 1 {
		cmd = fmt.Sprintf("sh -c '%s%s'", r.Env(), strings.Join(commands, "&&"))
	}
	if err := r.session.Run(cmd); err != nil {
		return fmt.Errorf("[%s] %s", r.Host.Name, err)
	}
	return nil
}

// Return preset environment variables to pass to execute
func (r *Remote) Env() string {
	return fmt.Sprint(
		"export HAP_HOSTNAME=\"", r.Host.Name, "\";",
		"export HAP_ADDR=\"", r.Host.Addr, "\";",
		"export HAP_USER=\"", r.Host.Username, "\";",
	)
}

// Writer with [host] prepended to output
func NewRemoteWriter(host string, w io.Writer) io.Writer {
	return &RemoteWriter{host: host, w: w}
}

// Writer with host and io.Writer
type RemoteWriter struct {
	host string
	w    io.Writer
}

// Implement io.Writer interface
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
