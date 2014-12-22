// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/gwoo/hap"
	"github.com/gwoo/hap/cmd/hap/cli"
)

var h = flag.String("h", "", "Individual host to use for commands.\n\t If empty, the default or a random host is used.")
var v = flag.Bool("v", false, "Verbose flag to print output")
var all = flag.Bool("all", false, "Use ALL the hosts.")
var logger VerboseLogger

func main() {
	flag.Usage = Usage
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
		cmdChan := make(chan cli.Command, len(hosts))
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

func start(host *hap.Host, cmd string, cmdChan chan cli.Command, errChan chan error) {
	command, err := run(host, cmd)
	cmdChan <- command
	errChan <- err
}

func run(host *hap.Host, cmd string) (cli.Command, error) {
	remote, err := hap.NewRemote(host)
	defer remote.Close()
	if err != nil {
		return nil, err
	}
	if command := cli.Commands.Get(cmd); command != nil {
		err := command.Run(remote)
		return command, err
	}
	return nil, fmt.Errorf("Command `%s` not found.", cmd)
}

func display(host *hap.Host, cmd string, cmdChan chan cli.Command, errChan chan error, done chan bool) {
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

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\nAvailable Commands:")
	w := new(tabwriter.Writer)
	w.Init(os.Stderr, 0, 8, 0, '\t', 0)
	for _, command := range cli.Commands {
		fmt.Fprintln(w, command.Help())
	}
	w.Flush()
}
