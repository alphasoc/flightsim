package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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
    run         Run all modules, or a particular module
    version     Prints the version number

Cheatsheet:
    flightsim run                Run all the modules
    flightsim run c2             Simulate C2 traffic
    flightsim run c2:trickbot    Simulate C2 traffic for the TrickBot family
`

func main() {
	url, ok := isLatestVersion()
	if !ok {
		fmt.Printf("New release found, check: %v\n\n", url)
	}

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

func isLatestVersion() (string, bool) {
	version, url, err := getLatestReleaseVersion()
	if err != nil || Version == version {
		return "", true
	}
	return url, false
}

func getLatestReleaseVersion() (string, string, error) {
	var version string
	var url string
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/alphasoc/flightsim/releases/latest", nil)
	if err != nil {
		return "", "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(body, &objmap)
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal(*objmap["tag_name"], &version)
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal(*objmap["html_url"], &url)
	if err != nil {
		return "", "", err
	}

	return version, url, nil
}
