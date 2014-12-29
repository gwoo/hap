// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
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
	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}
	if err := new(hap.Git).Exists(); err != nil {
		log.Fatal(err)
	}
	logger = VerboseLogger(*v)
	if cmd := flag.Arg(0); cmd != "" {
		command := cli.Commands.Get(cmd)
		if command == nil {
			log.Fatalf("Command `%s` not found.", cmd)
			return
		}
		if !command.IsRemote() {
			local(cmd, command)
			return
		}
		hf, err := hap.NewHapfile()
		if err != nil {
			log.Fatal(err)
		}
		remote(hf, cmd, command)
	}
}

func local(cmd string, command cli.Command) {
	err := command.Run(nil)
	fmt.Print(command.String())
	if err != nil {
		fmt.Printf("[%s] %s\n", cmd, err)
	}
	if log := command.Log(); log != "" {
		fmt.Printf("[%s] %s\n", cmd, log)
	}
}

func remote(hf hap.Hapfile, cmd string, command cli.Command) {
	hosts := hf.GetHosts(*h, *all)
	if len(hosts) < 1 {
		fmt.Printf("No host. Use -all or -host\n")
		return
	}
	done := make(chan bool, len(hosts))
	cmdChan := make(chan cli.Command, len(hosts))
	errChan := make(chan error, len(hosts))
	for name, host := range hosts {
		logger.Printf("[%s] Running `%s` on %s\n", name, cmd, host.Addr)
		go run(hf.Host(name), command, cmdChan, errChan)
	}
	for _, host := range hosts {
		go display(host, cmd, cmdChan, errChan, done)
		<-done
	}
}

func run(host *hap.Host, command cli.Command, cmdChan chan cli.Command, errChan chan error) {
	remote, err := hap.NewRemote(host)
	if err == nil {
		defer remote.Close()
		err = command.Run(remote)
	}
	cmdChan <- command
	errChan <- err
}

func display(host *hap.Host, cmd string, cmdChan chan cli.Command, errChan chan error, done chan bool) {
	logger.Printf("[%s] Results of `%s` on %s\n", host.Name, cmd, host.Addr)
	command := <-cmdChan
	err := <-errChan
	fmt.Print(command.String())
	if err != nil {
		fmt.Printf("[%s] %s\n", host.Name, err)
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
	keys := []string{}
	for key, _ := range cli.Commands {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		fmt.Fprintln(w, cli.Commands.Get(name).Help())
	}
	w.Flush()
}
