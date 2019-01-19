// Hap - the simple and effective provisioner
// Copyright (c) 2019 GWoo (https://github.com/gwoo)
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
	hf, err := hap.NewHapfile(*hapfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	var hosts = make(map[string]*hap.Host, 0)

	if _, ok := command.(*cli.DeployCmd); ok {
		deploy := flag.Arg(1)
		if *host == "" {
			*host = "*"
		}
		hosts, err = hf.GetDeployHosts(deploy, *host)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else if *host != "" {
		hosts = hf.GetHosts(*host)
	} else {
		fmt.Println("Missing host please specify -h or --host=")
		os.Exit(2)
	}
	if len(hosts) == 0 {
		fmt.Println("No host found")
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
