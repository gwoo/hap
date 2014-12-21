// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gwoo/hap"
)

var s = flag.String("s", "", "Individual server to use for commands. If empty first server will be used.")
var v = flag.Bool("v", false, "Verbose flag to print output")
var all = flag.Bool("all", false, "Use ALL the servers.")
var logger VerboseLogger

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
	logger = VerboseLogger(*v)
	if cmd := flag.Arg(0); cmd != "" {
		servers := cfg.Servers
		if *all == false {
			server := cfg.Server(*s)
			servers = map[string]*hap.Server{server.Name: server}
		}
		done := make(chan bool, len(servers))
		cmdChan := make(chan Command, len(servers))
		errChan := make(chan error, len(servers))
		for name, server := range servers {
			server.Name = name
			server.SetDefaults(cfg.Default)
			logger.Printf("[%s] Running `%s` on %s\n", server.Name, cmd, server.Addr)
			go start(server, cmd, cmdChan, errChan)
		}
		for _, server := range servers {
			go display(server, cmd, cmdChan, errChan, done)
			<-done
		}
	}
}

func start(server *hap.Server, cmd string, cmdChan chan Command, errChan chan error) {
	command, err := run(server, cmd)
	cmdChan <- command
	errChan <- err
}

func run(server *hap.Server, cmd string) (Command, error) {
	remote, err := hap.NewRemote(server)
	defer remote.Close()
	if err != nil {
		return nil, err
	}
	if command := commands.Get(cmd); command != nil {
		err := command.Run(remote)
		return command, err
	}
	return nil, fmt.Errorf("Command `%s` not found.", cmd)
}

func display(server *hap.Server, cmd string, cmdChan chan Command, errChan chan error, done chan bool) {
	logger.Printf("[%s] Results of `%s` on %s\n", server.Name, cmd, server.Addr)
	command := <-cmdChan
	err := <-errChan
	fmt.Print(command.String())
	if err != nil {
		logger.Printf("[%s] %s\n", server.Name, err)
	}
	logger.Printf("[%s] %s\n", server.Name, command.Log())
	done <- true
}
