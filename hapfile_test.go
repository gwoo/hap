// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"testing"
)

func TestGetHostsString(t *testing.T) {
	hf := &Hapfile{Hosts: map[string]*Host{"local": &Host{Addr: "127.0.0.1"}}}
	results := hf.GetHosts("local")
	if len(results) > 0 && results["local"].Addr == "127.0.0.1" {
		return
	}
	t.Error("host could not be found")
}

func TestGetHostsRegex(t *testing.T) {
	hf := &Hapfile{Hosts: map[string]*Host{"local": &Host{Addr: "127.0.0.1"}}}
	results := hf.GetHosts("lo*")
	if len(results) > 0 && results["local"].Addr == "127.0.0.1" {
		return
	}
	t.Error("host could not be found")
}

func TestGetHostsAll(t *testing.T) {
	hf := &Hapfile{Hosts: map[string]*Host{"local": &Host{Addr: "127.0.0.1"}}}
	results := hf.GetHosts("*")
	if len(results) > 0 && results["local"].Addr == "127.0.0.1" {
		return
	}
	t.Error("host could not be found")
}
