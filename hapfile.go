// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
package hap

import (
	"encoding/json"

	"code.google.com/p/gcfg"
)

// The Hapfile
type Hapfile struct {
	Default Default
	Hosts   map[string]*Host `gcfg:"host"`
}

// Get a host based on the name
// If the name is empty, and if default addr exists, return default
// otherwise return a random host
func (h Hapfile) Host(name string) *Host {
	if host, ok := h.Hosts[name]; ok {
		host.Name = name
		host.SetDefaults(h.Default)
		return host
	}
	if h.Default.Addr != "" {
		host := Host(h.Default)
		host.Name = "default"
		return &host
	}
	for name, host := range h.Hosts {
		host.Name = name
		host.SetDefaults(h.Default)
		return host
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
}

// Use the defaults to fill in missing host specific config
func (s *Host) SetDefaults(d Default) {
	if s.Username == "" {
		s.Username = d.Username
	}
	if s.Identity == "" {
		s.Identity = d.Identity
	}
	if s.Password == "" {
		s.Password = d.Password
	}
}

// Construct a new hapfile config
func NewHapfile() (Hapfile, error) {
	var hf Hapfile
	err := gcfg.ReadFileInto(&hf, "Hapfile")
	return hf, err
}
