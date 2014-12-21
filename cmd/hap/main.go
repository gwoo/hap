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
var v = flag.Bool("v", false, "Verbose flag to print output")

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
	if cmd := flag.Arg(0); cmd != "" {
		run(remote, cmd)
		return
	}
}

func run(remote *hap.Remote, cmd string) {
	if strings.Contains(cmd, ".json") {
		os.Args = append(os.Args, cmd)
		flag.Parse()
		cmd = "build"
	}
	if command := commands.Get(cmd); command != nil {
		err := command.Run(remote)
		fmt.Print(command)
		if err != nil {
			fmt.Println(err)
		}
		if command.String() == "" || *v == true {
			log.Println(command.Log())
		}
		return
	}
	log.Fatalf("Command `%s` not found.", cmd)

}
