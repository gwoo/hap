// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"encoding/json"
	"path/filepath"
	"sort"

	"gopkg.in/gcfg.v1"
)

// Hapfile defines the hosts, builds, and default
type Hapfile struct {
	Default Default
	Hosts   map[string]*Host  `gcfg:"host"`
	Builds  map[string]*Build `gcfg:"build"`
}

// GetHosts finds a list of hosts matching name string
func (h Hapfile) GetHosts(name string) map[string]*Host {
	hosts := h.Hosts
	keys := []string{}
	for key, _ := range hosts {
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
	Addr     string
	Username string
	Identity string
	Password string
	Build    []string
	Cmd      []string
	cmds     []string
}

// SetDefaults fills in missing host specific configs with defaults
func (h *Host) SetDefaults(d Default) {
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

// Build holds the cmds
type Build struct {
	Cmd []string
}

// NewHapfile constructs a new hapfile config
func NewHapfile(file string) (Hapfile, error) {
	var hf Hapfile
	err := gcfg.ReadFileInto(&hf, file)
	return hf, err
}
