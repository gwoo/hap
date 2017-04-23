// Hap - the simple and effective provisioner
// Copyright (c) 2017 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package hap

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestGitExists(t *testing.T) {
	if err := new(Git).Exists(); err != nil {
		t.Error(err)
	}
}

func TestGitPush(t *testing.T) {
	work := "/tmp/hap"
	repo := "/tmp/hap.git"
	os.RemoveAll(work)
	os.RemoveAll(repo)
	err := os.MkdirAll(work, os.ModePerm|os.ModeDir)
	if err != nil {
		t.Error(err)
		return
	}
	err = os.MkdirAll(repo, os.ModePerm|os.ModeDir)
	if err != nil {
		t.Error(err)
		return
	}
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = repo
	result, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(string(result))
		t.Error(err)
		return
	}
	cmd = exec.Command("git", "init", ".")
	cmd.Dir = work
	result, err = cmd.CombinedOutput()
	if err != nil {
		t.Log(string(result))
		t.Error(err)
		return
	}
	err = ioutil.WriteFile(work+"/test", []byte("testing"), 0777)
	if err != nil {
		t.Error(err)
		return
	}
	git := Git{Work: work, Repo: repo}
	if result, err := git.Commit("test commit"); err != nil {
		t.Log(string(result))
		t.Error(err)
	}
	if result, err := git.Push("master"); err != nil {
		t.Log(string(result))
		t.Error(err)
	}
}
