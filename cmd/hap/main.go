// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
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

var all = flag.Bool("all", false, "Use ALL the hosts.")
var host = flag.String("host", "", "Individual host to use for commands.")
var v = flag.Bool("v", false, "Verbose flag to print command log.")
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
	hosts := hf.GetHosts(*host, *all)
	if len(hosts) < 1 {
		fmt.Printf("No host. Use -all or -host\n")
		return
	}
	done := make(chan bool, len(hosts))
	cmdChan := make(chan cli.Command, len(hosts))
	errChan := make(chan error, len(hosts))
	for name, h := range hosts {
		logger.Printf("[%s] Running `%s` on %s\n", name, cmd, h.Addr)
		go run(h, command, cmdChan, errChan)
	}
	for _, h := range hosts {
		go display(h, cmd, cmdChan, errChan, done)
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
