// Hap - the simple and effective provisioner
// Copyright (c) 2019 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	gcfg "gopkg.in/gcfg.v1"
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

func TestGcfgInclude(t *testing.T) {
	cfgStr := `; Comment line
[include]
path = path/to/file.cfg`
	var cfg Hapfile
	err := gcfg.ReadStringInto(&cfg, cfgStr)
	if err != nil {
		t.Fatalf("Failed to parse gcfg data: %s", err)
	}
	if cfg.Include.Path[0] != "path/to/file.cfg" {
		t.Fatalf("Unexpected result: %s", cfg.Include.Path[0])
	}
}

func TestDefaultHapfile(t *testing.T) {
	cfgStr := `; Comment line
[default]
username = bob
password = password
identity = ~/.ssh/id_rsa`
	err := ioutil.WriteFile("TestHapfile", []byte(cfgStr), 0666)
	if err != nil {
		t.Error(err)
	}
	hf, err := NewHapfile("TestHapfile")
	if err != nil {
		t.Error(err)
	}
	ws := hf.Default.Username
	gs := "bob"
	if ws != gs {
		t.Error("Want:", ws, "Got:", gs)
	}
	ws = hf.Default.Identity
	gs = "~/.ssh/id_rsa"
	if ws != gs {
		t.Error("Want:", ws, "Got:", gs)
	}
	err = os.Remove("TestHapfile")
	if err != nil {
		t.Error(err)
	}
}

func TestNewHapfileWithInclude(t *testing.T) {
	cfgStr := `
[env]
file = environment

[host "primary"]
addr = "10.0.0.1:22"
build = "init"
build = "test"
env = primary_environment

[include]
path = TestAnotherHapfile

[build "init"]
cmd = "echo init"`

	err := ioutil.WriteFile("TestHapfile", []byte(cfgStr), 0666)
	if err != nil {
		t.Error(err)
	}
	cfgStr = `
[env]
file = another_environment

[host "secondary"]
addr = "10.0.0.2:22"
build = "init"
build = "test"
env = secondary_environment

[build "test"]
cmd = "echo test"`
	err = ioutil.WriteFile("TestAnotherHapfile", []byte(cfgStr), 0666)
	if err != nil {
		t.Error(err)
	}

	hf, err := NewHapfile("TestHapfile")
	if err != nil {
		t.Error(err)
	}
	hosts := hf.GetHosts("*")
	if len(hosts) < 2 {
		t.Error("Expected at least two hosts")
	}

	p := hf.Host("primary")
	w1 := "10.0.0.1:22"
	g1 := p.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 := []string{"init", "test"}
	g2 := p.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}
	w3 := []string{"another_environment", "environment", "primary_environment"}
	g3 := p.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}

	s := hf.Host("secondary")
	w1 = "10.0.0.2:22"
	g1 = s.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 = []string{"init", "test"}
	g2 = s.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}

	w3 = []string{"another_environment", "environment", "secondary_environment"}
	g3 = s.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}

	err = os.Remove("TestHapfile")
	if err != nil {
		t.Error(err)
	}
	err = os.Remove("TestAnotherHapfile")
	if err != nil {
		t.Error(err)
	}
}

func TestNewHapfileWithIncludeBuild(t *testing.T) {
	cfgStr := `
[env]
file = environment

[host "primary"]
addr = "10.0.0.1:22"
build = "init"
build = "test"
env = primary_environment

[include]
path = TestAnotherHapfile`

	err := ioutil.WriteFile("TestHapfile", []byte(cfgStr), 0666)
	if err != nil {
		t.Error(err)
	}
	cfgStr = `
[env]
file = another_environment

[host "secondary"]
addr = "10.0.0.2:22"
build = "init"
build = "test"
env = secondary_environment

[build "test"]
cmd = "echo test"

[build "init"]
cmd = "echo init"`
	err = ioutil.WriteFile("TestAnotherHapfile", []byte(cfgStr), 0666)
	if err != nil {
		t.Error(err)
	}

	hf, err := NewHapfile("TestHapfile")
	if err != nil {
		t.Error(err)
	}
	hosts := hf.GetHosts("*")
	if len(hosts) < 2 {
		t.Error("Expected at least two hosts")
	}

	p := hf.Host("primary")
	w1 := "10.0.0.1:22"
	g1 := p.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 := []string{"init", "test"}
	g2 := p.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}
	w3 := []string{"another_environment", "environment", "primary_environment"}
	g3 := p.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}

	s := hf.Host("secondary")
	w1 = "10.0.0.2:22"
	g1 = s.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 = []string{"init", "test"}
	g2 = s.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}

	w3 = []string{"another_environment", "environment", "secondary_environment"}
	g3 = s.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}

	err = os.Remove("TestHapfile")
	if err != nil {
		t.Error(err)
	}
	err = os.Remove("TestAnotherHapfile")
	if err != nil {
		t.Error(err)
	}
}

func TestCustomWorkingDir(t *testing.T) {
	cfgStr := `; Comment line
[default]
dir = hap-working-directory

[host "one"]
addr = 10.0.0.1:22`
	err := ioutil.WriteFile("TestHapfile", []byte(cfgStr), 0666)
	if err != nil {
		t.Error(err)
	}
	hf, err := NewHapfile("TestHapfile")
	if err != nil {
		t.Error(err)
	}
	host := hf.Host("one")
	ws := host.GetDir()
	gs := "hap-working-directory"
	if ws != gs {
		t.Error("Want:", ws, "Got:", gs)
	}
	err = os.Remove("TestHapfile")
	if err != nil {
		t.Error(err)
	}
}

func TestNewHapfileWithDeploy(t *testing.T) {
	cfgStr := `
[host "one"]
addr = "10.0.0.1:22"

[host "two"]
addr = "10.0.0.2:22"

[env]
file = environment

[deploy "together"]
host = one
host = two
build = init
build = test
cmd = echo
env = another_environment

[build "test"]
cmd = "echo test"

[build "init"]
cmd = "echo init"`
	err := ioutil.WriteFile("TestHapfile", []byte(cfgStr), 0666)
	if err != nil {
		t.Error(err)
	}

	hf, err := NewHapfile("TestHapfile")
	if err != nil {
		t.Error(err)
	}
	hosts, err := hf.GetDeployHosts("together", "*")
	if err != nil {
		t.Error(err)
	}
	if len(hosts) < 2 {
		t.Error("Expected at least two hosts")
	}

	p := hf.DeployHost("together", "one")
	w1 := "10.0.0.1:22"
	g1 := p.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 := []string{"init", "test"}
	g2 := p.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}
	w21 := []string{"echo"}
	g21 := p.Cmd
	if !reflect.DeepEqual(w21, g21) {
		t.Error("Want:", w21, "Got:", g21)
	}
	w3 := []string{"environment", "another_environment"}
	g3 := p.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}
	w4 := []string{"echo init", "echo test", "echo"}
	g4 := p.Cmds()
	if !reflect.DeepEqual(w4, g4) {
		t.Error("Want:", w21, "Got:", g21)
	}
	s := hf.DeployHost("together", "two")
	w1 = "10.0.0.2:22"
	g1 = s.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 = []string{"init", "test"}
	g2 = s.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}
	w21 = []string{"echo"}
	g21 = p.Cmd
	if !reflect.DeepEqual(w21, g21) {
		t.Error("Want:", w21, "Got:", g21)
	}
	w3 = []string{"environment", "another_environment"}
	g3 = s.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}

	err = os.Remove("TestHapfile")
	if err != nil {
		t.Error(err)
	}
}

func TestNewHapfileWithDeployAndHost(t *testing.T) {
	cfgStr := `
[host "one"]
addr = "10.0.0.1:22"
build = test

[host "two"]
addr = "10.0.0.2:22"

[env]
file = environment

[deploy "together"]
host = one
host = two
build = init
build = test
cmd = echo
env = another_environment

[build "test"]
cmd = "echo test"

[build "init"]
cmd = "echo init"`
	err := ioutil.WriteFile("TestHapfile", []byte(cfgStr), 0666)
	if err != nil {
		t.Error(err)
	}

	hf, err := NewHapfile("TestHapfile")
	if err != nil {
		t.Error(err)
	}
	hosts := hf.GetHosts("*")
	if len(hosts) < 2 {
		t.Error("Expected at least two hosts")
	}
	dhosts, err := hf.GetDeployHosts("together", "*")
	if err != nil {
		t.Error(err)
	}
	if len(dhosts) < 2 {
		t.Error("Expected at least two hosts")
	}

	p := hf.DeployHost("together", "one")
	w1 := "10.0.0.1:22"
	g1 := p.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 := []string{"init", "test"}
	g2 := p.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}
	w21 := []string{"echo"}
	g21 := p.Cmd
	if !reflect.DeepEqual(w21, g21) {
		t.Error("Want:", w21, "Got:", g21)
	}
	w3 := []string{"environment", "another_environment"}
	g3 := p.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}
	w4 := []string{"echo init", "echo test", "echo"}
	g4 := p.Cmds()
	if !reflect.DeepEqual(w4, g4) {
		t.Error("Want:", w21, "Got:", g21)
	}
	s := hf.DeployHost("together", "two")
	w1 = "10.0.0.2:22"
	g1 = s.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 = []string{"init", "test"}
	g2 = s.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}
	w21 = []string{"echo"}
	g21 = p.Cmd
	if !reflect.DeepEqual(w21, g21) {
		t.Error("Want:", w21, "Got:", g21)
	}
	w3 = []string{"environment", "another_environment"}
	g3 = s.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}

	// This host should have fewer commands
	p = hf.Host("one")
	w1 = "10.0.0.1:22"
	g1 = p.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 = []string{"test"}
	g2 = p.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}
	var w22 []string
	g22 := p.Cmd
	if !reflect.DeepEqual(w22, g22) {
		t.Error("Want:", w22, "Got:", g22)
	}
	w3 = []string{"environment"}
	g3 = p.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}
	w4 = []string{"echo test"}
	g4 = p.Cmds()
	if !reflect.DeepEqual(w4, g4) {
		t.Error("Want:", w21, "Got:", g21)
	}

	s = hf.Host("two")
	w1 = "10.0.0.2:22"
	g1 = s.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	g22 = s.Build
	if !reflect.DeepEqual(w22, g22) {
		t.Error("Want:", w22, "Got:", g22)
	}
	g22 = p.Cmd
	if !reflect.DeepEqual(w22, g22) {
		t.Error("Want:", w22, "Got:", g22)
	}
	w3 = []string{"environment"}
	g3 = s.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}

	err = os.Remove("TestHapfile")
	if err != nil {
		t.Error(err)
	}
}

func TestNewHapfileWithIncludeHosts(t *testing.T) {
	cfgStr := `
[env]
file = another_environment

[host "one"]
addr = "10.0.0.1:22"

[host "two"]
addr = "10.0.0.2:22"

`

	err := ioutil.WriteFile("TestHostsfile", []byte(cfgStr), 0666)
	if err != nil {
		t.Error(err)
	}
	cfgStr = `
[env]
file = environment

[include]
path = TestHostsfile

[deploy "together"]
host = one
host = two
build = init
build = test
cmd = echo

[build "test"]
cmd = "echo test"

[build "init"]
cmd = "echo init"`
	err = ioutil.WriteFile("TestHapfile", []byte(cfgStr), 0666)
	if err != nil {
		t.Error(err)
	}

	hf, err := NewHapfile("TestHapfile")
	if err != nil {
		t.Error(err)
	}
	hosts := hf.GetHosts("*")
	if len(hosts) < 2 {
		t.Error("Expected at least two hosts")
	}

	p := hf.DeployHost("together", "one")
	w1 := "10.0.0.1:22"
	g1 := p.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 := []string{"init", "test"}
	g2 := p.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}
	w21 := []string{"echo"}
	g21 := p.Cmd
	if !reflect.DeepEqual(w21, g21) {
		t.Error("Want:", w21, "Got:", g21)
	}
	w3 := []string{"another_environment", "environment"}
	g3 := p.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}
	w4 := []string{"echo init", "echo test", "echo"}
	g4 := p.Cmds()
	if !reflect.DeepEqual(w4, g4) {
		t.Error("Want:", w21, "Got:", g21)
	}
	s := hf.DeployHost("together", "two")
	w1 = "10.0.0.2:22"
	g1 = s.Addr
	if w1 != g1 {
		t.Error("Want:", w1, "Got:", g1)
	}
	w2 = []string{"init", "test"}
	g2 = s.Build
	if !reflect.DeepEqual(w2, g2) {
		t.Error("Want:", w2, "Got:", g2)
	}
	w21 = []string{"echo"}
	g21 = p.Cmd
	if !reflect.DeepEqual(w21, g21) {
		t.Error("Want:", w21, "Got:", g21)
	}
	w3 = []string{"another_environment", "environment"}
	g3 = s.Env
	if !reflect.DeepEqual(w3, g3) {
		t.Error("Want:", w3, "Got:", g3)
	}

	err = os.Remove("TestHostsfile")
	if err != nil {
		t.Error(err)
	}
	err = os.Remove("TestHapfile")
	if err != nil {
		t.Error(err)
	}
}
