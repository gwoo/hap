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
	Servers map[string]*Server `gcfg:"server"`
}

// Get a server based on the name
// If the name is empty, and if default addr exists, return default
// otherwise return a random server
func (h Hapfile) Server(name string) *Server {
	if server, ok := h.Servers[name]; ok {
		server.Name = name
		server.SetDefaults(h.Default)
		return server
	}
	if h.Default.Addr != "" {
		server := Server(h.Default)
		server.Name = "default"
		return &server
	}
	for name, server := range h.Servers {
		server.Name = name
		server.SetDefaults(h.Default)
		return server
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
type Default Server

// A remote machine
type Server struct {
	Name     string
	Addr     string
	Username string
	Identity string
	Password string
}

// Use the defaults to fill in missing server specific config
func (s *Server) SetDefaults(d Default) {
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
