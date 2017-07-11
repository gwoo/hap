// Hap - the simple and effective provisioner
// Copyright (c) 2017 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"text/tabwriter"

	"github.com/gwoo/hap"
	"github.com/gwoo/hap/cmd/hap/cli"
	flag "github.com/ogier/pflag"
)

var cluster = flag.StringP("cluster", "c", "", "Cluster of hosts to use for commands.")
var host = flag.StringP("host", "h", "", "Host to use for commands. Use glob patterns to match multiple hosts. Use --host=* for all hosts.")
var hapfile = flag.StringP("file", "f", "Hapfile", "Location of a Hapfile.")
var help = flag.BoolP("help", "", false, "Show help")
var verbose = flag.BoolP("verbose", "v", false, "[deprecated] Verbose mode is always on")

var logger VerboseLogger

// Version is just the version of hap
var Version string

func main() {
	flag.Usage = Usage
	flag.Parse()

	if err := new(hap.Git).Exists(); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	var command cli.Command
	if cmd := flag.Arg(0); cmd != "" {
		command = cli.Commands.Get(cmd)
		if command == nil {
			fmt.Printf("Command `%s` not found.\n", cmd)
		}
	}
	if len(os.Args) <= 1 || *help || command == nil {
		flag.Usage()
		return
	}
	if !command.IsRemote() {
		run(nil, command)
		return
	}
	if *host == "" && *cluster == "" {
		fmt.Println("Missing host or cluster flag: Please specify -h or --host=, -c or --cluster=")
		os.Exit(2)
	}
	hf, err := hap.NewHapfile(*hapfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	var hosts = make(map[string]*hap.Host, 0)
	if c, ok := hf.Clusters[*cluster]; ok {
		for _, n := range c.Host {
			hosts[n] = hf.Host(n)
		}
	} else if *host != "" {
		hosts = hf.GetHosts(*host)
	}
	if len(hosts) == 0 {
		fmt.Printf("No hosts found for `%s`\n", *host)
		return
	}
	var wg sync.WaitGroup
	for _, h := range hosts {
		wg.Add(1)
		go func(h *hap.Host) {
			defer wg.Done()
			run(h, command)
		}(h)
	}
	wg.Wait()
}

func run(host *hap.Host, command cli.Command) {
	var remote *hap.Remote
	var err error
	if host != nil {
		remote, err = hap.NewRemote(host)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer remote.Close()
	}
	result, err := command.Run(remote)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

// Usage prints out the hap CLI usage
func Usage() {
	fmt.Printf("Version: %s\n", Version)
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\nAvailable Commands:")
	w := new(tabwriter.Writer)
	w.Init(os.Stderr, 0, 8, 0, '\t', 0)
	keys := []string{}
	for key := range cli.Commands {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		fmt.Fprintln(w, cli.Commands.Get(name).Help())
	}
	w.Flush()
}
