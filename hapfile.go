// Hap - the simple and effective provisioner
// Copyright (c) 2019 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"

	gcfg "gopkg.in/gcfg.v1"
)

// Hapfile defines the hosts, builds, and default
type Hapfile struct {
	Default Default
	Deploys map[string]*Deploy `gcfg:"deploy"`
	deploys map[string]map[string]*Host
	Hosts   map[string]*Host  `gcfg:"host"`
	Builds  map[string]*Build `gcfg:"build"`
	Include Include           `gcfg:"include"`
	Env     Env               `gcfg:"env"`
}

// NewHapfile constructs a new hapfile config
func NewHapfile(file string) (Hapfile, error) {
	hf, err := include(file)
	if err != nil {
		return hf, err
	}
	hf.deploys = map[string]map[string]*Host{}
	for _, file := range hf.Include.Path {
		nhf, err := include(file)
		if err != nil {
			return hf, err
		}
		for n, c := range nhf.Deploys {
			if _, ok := hf.Deploys[n]; !ok {
				hf.Deploys[n] = c
			}
		}
		for n, h := range nhf.Hosts {
			if _, ok := hf.Hosts[n]; !ok {
				hf.Hosts[n] = h
			}
		}
		for n, b := range nhf.Builds {
			if _, ok := hf.Builds[n]; !ok {
				hf.Builds[n] = b
			}
		}
		for d, deploy := range nhf.Deploys {
			hf.deploys[d] = map[string]*Host{}
			for _, n := range deploy.Host {
				if _, ok := hf.Hosts[n]; ok {
					hf.deploys[d][n] = hf.newDeployHost(hf.Hosts[n], deploy)
				}
			}
		}
		for _, file := range nhf.Env.File {
			hf.Env.File = append(hf.Env.File, file)
		}
	}
	for _, host := range hf.Hosts {
		for _, file := range hf.Env.File {
			host.Env = append(host.Env, file)
		}
		for i, j := 0, len(host.Env)-1; i < j; i, j = i+1, j-1 {
			host.Env[i], host.Env[j] = host.Env[j], host.Env[i]
		}
	}
	for d, deploy := range hf.Deploys {
		hf.deploys[d] = map[string]*Host{}
		for _, n := range deploy.Host {
			if _, ok := hf.Hosts[n]; ok {
				hf.deploys[d][n] = hf.newDeployHost(hf.Hosts[n], deploy)
			}
		}
	}
	return hf, err
}

func include(file string) (Hapfile, error) {
	var hf Hapfile
	err := gcfg.ReadFileInto(&hf, file)
	if hf.Deploys == nil {
		hf.Deploys = make(map[string]*Deploy, 0)
	}
	if hf.Hosts == nil {
		hf.Hosts = make(map[string]*Host, 0)
	}
	if hf.Builds == nil {
		hf.Builds = make(map[string]*Build, 0)
	}
	return hf, err
}

func (hf Hapfile) newDeployHost(host *Host, d *Deploy) *Host {
	h := &Host{}
	h.Name = host.Name
	h.Addr = host.Addr
	h.Dir = host.Dir
	h.Username = host.Username
	h.Identity = host.Identity
	h.Password = host.Password
	h.Env = append(host.Env, d.Env...)
	h.Build = d.Build
	h.Cmd = d.Cmd
	return h
}

// GetDeployHosts finds a list of hosts matching name string
func (hf Hapfile) GetDeployHosts(deploy, host string) (map[string]*Host, error) {
	hosts, ok := hf.deploys[deploy]
	if !ok {
		return map[string]*Host{}, fmt.Errorf("deploy '%s' not found", deploy)
	}
	keys := []string{}
	for key := range hosts {
		if matched, _ := filepath.Match(host, key); matched == true {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	results := make(map[string]*Host)
	for _, key := range keys {
		results[key] = hf.DeployHost(deploy, key)
	}
	return results, nil
}

// DeployHost takes a name and returns the host
// If the name is empty and default addr exists return default.
// If no default is set it returns a random host.
func (hf Hapfile) DeployHost(deploy, host string) *Host {
	if h, ok := hf.deploys[deploy][host]; ok {
		h.Name = host
		h.SetDefaults(hf.Default)
		h.BuildCmds(hf.Builds)
		return h
	}
	if hf.Default.Addr != "" {
		h := Host(hf.Default)
		h.Name = "default"
		h.BuildCmds(hf.Builds)
		return &h
	}
	return nil
}

// GetHosts finds a list of hosts matching name string
func (hf Hapfile) GetHosts(name string) map[string]*Host {
	hosts := hf.Hosts
	keys := []string{}
	for key := range hosts {
		if matched, _ := filepath.Match(name, key); matched == true {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	results := make(map[string]*Host)
	for _, key := range keys {
		results[key] = hf.Host(key)
	}
	return results
}

// Host takes a name and returns the host
// If the name is empty and default addr exists return default.
// If no default is set it returns a random host.
func (hf Hapfile) Host(name string) *Host {
	if host, ok := hf.Hosts[name]; ok {
		host.Name = name
		host.SetDefaults(hf.Default)
		host.BuildCmds(hf.Builds)
		return host
	}
	if hf.Default.Addr != "" {
		host := Host(hf.Default)
		host.Name = "default"
		host.BuildCmds(hf.Builds)
		return &host
	}
	return nil
}

// String returns the hapfile config as json
func (hf Hapfile) String() string {
	b, err := json.Marshal(hf)
	if err != nil {
		return ""
	}
	return string(b)
}

// Deploy describes a group of remote machine
type Deploy struct {
	Host  []string
	Build []string
	Cmd   []string
	Env   []string
}

// Default holds the default settings
type Default Host

// Host describes a remote machine
type Host struct {
	Name     string
	Dir      string
	Addr     string
	Username string
	Identity string
	Password string
	Env      []string
	Build    []string
	Cmd      []string
	cmds     []string
}

// SetDefaults fills in missing host specific configs with defaults
func (h *Host) SetDefaults(d Default) {
	if h.Addr == "" {
		h.Addr = d.Addr
	}
	if h.Dir == "" {
		h.Dir = d.Dir
	}
	if h.Username == "" {
		h.Username = d.Username
	}
	if h.Identity == "" {
		h.Identity = d.Identity
	}
	if h.Password == "" {
		h.Password = d.Password
	}
	if len(h.Build) < 1 {
		h.Build = d.Build
	}
	if len(h.Cmd) < 1 {
		h.Cmd = d.Cmd
	}
}

// GetDir returns the current working directory
func (h *Host) GetDir() string {
	if h.Dir != "" {
		return h.Dir
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	return path.Base(cwd)
}

// BuildCmds combines the builds and cmds
func (h *Host) BuildCmds(builds map[string]*Build) {
	h.cmds = []string{}
	for _, build := range h.Build {
		if b, ok := builds[build]; ok {
			h.cmds = append(h.cmds, b.Cmd...)
		}
	}
	h.cmds = append(h.cmds, h.Cmd...)
}

// Cmds returns the cmds to build
func (h *Host) Cmds() []string {
	return h.cmds
}

// AddEnv includes env files in cmds
func (h *Host) AddEnv(cmds []string) []string {
	for _, file := range h.Env {
		cmds = append(cmds, fmt.Sprint(". ./", file))
	}
	return cmds
}

// Build holds the cmds
type Build struct {
	Cmd []string
}

// Include holds the files to include
type Include struct {
	Path []string
}

// Env holds the files to source when running commands
type Env struct {
	File []string
}
