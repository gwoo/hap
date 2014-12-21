// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
package hap

import (
	"encoding/json"

	"code.google.com/p/gcfg"
)

type Config struct {
	Default Default
	Servers map[string]*Server `gcfg:"server"`
}

func (h Config) Server(name string) *Server {
	if server, ok := h.Servers[name]; ok {
		server.SetDefaults(h.Default)
		return server
	}
	for _, server := range h.Servers {
		server.SetDefaults(h.Default)
		return server
	}
	return nil
}

func (h Config) String() string {
	b, err := json.Marshal(h)
	if err != nil {
		return ""
	}
	return string(b)
}

type Default struct {
	Username string
	Identity string
	Password string
}

type Server struct {
	Addr     string
	Username string
	Identity string
	Password string
}

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
func NewConfig() (Config, error) {
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, "Hapfile")
	return cfg, err
}
