// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"encoding/json"
	"fmt"
	"sort"

	"code.google.com/p/gcfg"
)

// The Hapfile
type Hapfile struct {
	Default Default
	Hosts   map[string]*Host  `gcfg:"host"`
	Builds  map[string]*Build `gcfg:"build"`
}

// Get list of hosts
func (h Hapfile) GetHosts(name string, all bool) map[string]*Host {
	if all == false {
		if host := h.Host(name); host != nil {
			return map[string]*Host{name: host}
		}
		return nil
	}
	hosts := h.Hosts
	keys := []string{}
	for key, _ := range hosts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	results := make(map[string]*Host)
	for _, key := range keys {
		results[key] = h.Host(key)
	}
	fmt.Printf("%#v", results)
	return results
}

// Get a host based on the name
// If the name is empty, and if default addr exists, return default
// otherwise return a random host
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

// Return the hapfile config as json
func (h Hapfile) String() string {
	b, err := json.Marshal(h)
	if err != nil {
		return ""
	}
	return string(b)
}

// The default settings
type Default Host

// A remote machine
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

// Use the defaults to fill in missing host specific config
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

// Combine the builds and cmds
func (h *Host) BuildCmds(builds map[string]*Build) {
	h.cmds = []string{}
	for _, build := range h.Build {
		if b, ok := builds[build]; ok {
			h.cmds = append(h.cmds, b.Cmd...)
		}
	}
	h.cmds = append(h.cmds, h.Cmd...)
}

// Get the cmds to build
func (h *Host) Cmds() []string {
	return h.cmds
}

type Build struct {
	Cmd []string
}

// Construct a new hapfile config
func NewHapfile() (Hapfile, error) {
	var hf Hapfile
	err := gcfg.ReadFileInto(&hf, "Hapfile")
	return hf, err
}
