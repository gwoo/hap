// Hap - the simple and effective provisioner
// Copyright (c) 2017 GWoo (https://github.com/gwoo)
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

	"gopkg.in/gcfg.v1"
)

// Hapfile defines the hosts, builds, and default
type Hapfile struct {
	Default Default
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
	for _, file := range hf.Include.Path {
		nhf, err := include(file)
		if err != nil {
			return hf, err
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
	return hf, err
}

func include(file string) (Hapfile, error) {
	var hf Hapfile
	err := gcfg.ReadFileInto(&hf, file)
	if hf.Hosts == nil {
		hf.Hosts = make(map[string]*Host, 0)
	}
	if hf.Builds == nil {
		hf.Builds = make(map[string]*Build, 0)
	}
	return hf, err
}

// GetHosts finds a list of hosts matching name string
func (h Hapfile) GetHosts(name string) map[string]*Host {
	hosts := h.Hosts
	keys := []string{}
	for key := range hosts {
		if matched, _ := filepath.Match(name, key); matched == true {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	results := make(map[string]*Host)
	for _, key := range keys {
		results[key] = h.Host(key)
	}
	return results
}

// Host takes a name and returns the host
// If the name is empty and default addr exists return default.
// If no default is set it returns a random host.
func (h Hapfile) Host(name string) *Host {
	if host, ok := h.Hosts[name]; ok {
		host.Name = name
		host.SetDefaults(h.Default)
		host.BuildCmds(h.Builds)
		return host
	}
	if h.Default.Addr != "" {
		host := Host(h.Default)
		host.Name = "default"
		host.BuildCmds(h.Builds)
		return &host
	}
	return nil
}

// String returns the hapfile config as json
func (h Hapfile) String() string {
	b, err := json.Marshal(h)
	if err != nil {
		return ""
	}
	return string(b)
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
