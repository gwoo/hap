// Hap - the simple and effective provisioner
// Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"text/tabwriter"

	"github.com/gwoo/hap"
	"github.com/gwoo/hap/cmd/hap/cli"
	flag "github.com/ogier/pflag"
)

var host = flag.StringP("host", "h", "", "Host to use for commands. Use glob patterns to match multiple hosts. Use --host=* for all hosts.")
var verbose = flag.BoolP("verbose", "v", false, "Verbose flag to print command log.")
var hapfile = flag.StringP("file", "f", "Hapfile", "Location of a Hapfile.")

var logger VerboseLogger

// Version is just the version of hap
var Version string

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
	logger = VerboseLogger(*verbose)
	if cmd := flag.Arg(0); cmd != "" {
		command := cli.Commands.Get(cmd)
		if command == nil {
			log.Fatalf("Command `%s` not found.", cmd)
			return
		}
		if !command.IsRemote() {
			run(nil, command)
			return
		}
		hf, err := hap.NewHapfile(*hapfile)
		if err != nil {
			log.Fatal(err)
		}
		if *host == "" {
			fmt.Println("Missing host flag: Please specify -h or --host=")
			return
		}
		hosts := hf.GetHosts(*host)
		if len(hosts) < 0 {
			log.Fatalf("No hosts found for %s", *host)
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
