// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gwoo/hap"
)

var s = flag.String("s", "", "Individual server to use for commands. If empty first server will be used.")

//var all = flag.Bool("all", false, "Use ALL the servers.")

func main() {
	flag.Parse()
	if err := new(hap.Git).Exists(); err != nil {
		log.Fatal(err)
	}
	cfg, err := hap.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}
	server := cfg.Server(*s)
	config := hap.SshConfig{
		Addr:     server.Addr,
		Username: server.Username,
		Identity: server.Identity,
		Password: server.Password,
	}
	remote, err := hap.NewRemote(config)
	defer remote.Close()
	if err != nil {
		log.Fatal(err)
	}
	if cmd := os.Args[1]; cmd != "" {
		run(remote, cmd)
		return
	}
}

func run(remote *hap.Remote, cmd string) {
	scripts := ""
	if strings.Contains(cmd, ".json") {
		scripts = cmd
		cmd = "build"
	}
	switch cmd {
	case "init":
		result, err := remote.Initialize()
		if err != nil {
			fmt.Print(string(result))
			log.Fatal(err)
		}
		log.Printf("%s initialized on %s.", remote.Dir, remote.Config.Addr)
	case "push":
		result, err := remote.Push()
		if err != nil {
			fmt.Print(string(result))
			log.Fatal(err)
		}
		log.Printf("%s pushed to %s.", remote.Dir, remote.Config.Addr)
	case "-c":
		if len(os.Args) <= 2 {
			log.Fatal("Missing arguments.")
		}
		cmd := strings.Join(os.Args[2:], " ")
		result, err := remote.Execute([]string{cmd})
		if string(result) == "" {
			log.Printf("Executed %s to %s.", cmd, remote.Config.Addr)
		}
		fmt.Print(string(result))
		if err != nil {
			log.Fatal(err)
		}
	case "exec":
		if len(os.Args) <= 2 {
			log.Fatal("Missing arguments.")
		}
		result, err := remote.Push()
		if err != nil {
			fmt.Print(string(result))
			log.Fatal(err)
		}
		cmd := strings.Join(os.Args[2:], " ")
		result, err = remote.Execute([]string{"cd " + remote.Dir, "bash " + cmd})
		if string(result) == "" {
			log.Printf("Executed %s to %s.", cmd, remote.Config.Addr)
		}
		fmt.Print(string(result))
		if err != nil {
			log.Fatal(err)
		}
	case "build":
		result, err := remote.Build(scripts)
		fmt.Print(string(result))
		if err != nil {
			log.Fatal(err)
		}
	}
}
