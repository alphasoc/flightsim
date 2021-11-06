package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alphasoc/flightsim/cmd/get"
	"github.com/alphasoc/flightsim/cmd/run"
)

var Version = ""

var usage = `AlphaSOC Network Flight Simulator™ (https://github.com/alphasoc/flightsim)

flightsim is an application which generates malicious network traffic for security
teams to evaluate security controls (e.g. firewalls) and ensure that monitoring tools
are able to detect malicious traffic.

Usage:
    flightsim <command> [arguments]

Available commands:
    get         Get a list of elements (ie. families) of a certain category (ie. c2)
    run         Run all modules, or a particular module
    version     Prints the version number

Cheatsheet:
    flightsim run                   Run all the modules
    flightsim run c2                Simulate C2 traffic
    flightsim run c2:trickbot       Simulate C2 traffic for the TrickBot family
    flightsim run ssh-transfer:1GB  Simulate a 1GB SSH/SFTP file transfer

    flightsim get families:c2       Get a list of all c2 families
`

func main() {
	cmdRoot := flag.NewFlagSet("flightsim", flag.ExitOnError)
	cmdRoot.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
	}

	cmdRoot.Parse(os.Args[1:])

	if len(cmdRoot.Args()) == 0 {
		cmdRoot.Usage()
		os.Exit(1)
	}

	cmd, args := cmdRoot.Arg(0), cmdRoot.Args()[1:]

	var err error

	switch cmd {
	case "get":
		get.Version = Version
		err = get.RunCmd(args)
	case "run":
		run.Version = Version
		err = run.RunCmd(args)
	case "version":
		fmt.Printf("flightsim version %s\n", Version)
		return
	default:
		err = fmt.Errorf("invalid command: %s", cmd)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
