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

var h = flag.String("h", "", "Individual host to use for commands. If empty, the default or a random host is used.")
var v = flag.Bool("v", false, "Verbose flag to print output")
var all = flag.Bool("all", false, "Use ALL the hosts.")
var logger VerboseLogger

func main() {
	flag.Parse()
	if err := new(hap.Git).Exists(); err != nil {
		log.Fatal(err)
	}
	hf, err := hap.NewHapfile()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}
	logger = VerboseLogger(*v)
	if cmd := flag.Arg(0); cmd != "" {
		hosts := hf.Hosts
		if *all == false {
			host := hf.Host(*h)
			hosts = map[string]*hap.Host{host.Name: host}
		}
		done := make(chan bool, len(hosts))
		cmdChan := make(chan Command, len(hosts))
		errChan := make(chan error, len(hosts))
		for name, host := range hosts {
			host.Name = name
			host.SetDefaults(hf.Default)
			logger.Printf("[%s] Running `%s` on %s\n", host.Name, cmd, host.Addr)
			go start(host, cmd, cmdChan, errChan)
		}
		for _, host := range hosts {
			go display(host, cmd, cmdChan, errChan, done)
			<-done
		}
	}
}

func start(host *hap.Host, cmd string, cmdChan chan Command, errChan chan error) {
	command, err := run(host, cmd)
	cmdChan <- command
	errChan <- err
}

func run(host *hap.Host, cmd string) (Command, error) {
	remote, err := hap.NewRemote(host)
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

func display(host *hap.Host, cmd string, cmdChan chan Command, errChan chan error, done chan bool) {
	logger.Printf("[%s] Results of `%s` on %s\n", host.Name, cmd, host.Addr)
	command := <-cmdChan
	err := <-errChan
	fmt.Print(command.String())
	if err != nil {
		logger.Printf("[%s] %s\n", host.Name, err)
	}
	logger.Printf("[%s] %s\n", host.Name, command.Log())
	done <- true
}
