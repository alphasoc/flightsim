# Network Flight Simulator

**flightsim** is a lightweight utility used to generate malicious network traffic and help security teams to evaluate security controls and network visibility. The tool performs tests to simulate DNS tunneling, DGA traffic, requests to known active C2 destinations, and other suspicious traffic patterns.

## Installation

Download the latest flightsim binary for your OS from the [GitHub Releases](https://github.com/alphasoc/flightsim/releases) page. Alternatively, the utility can be built using [Golang](https://golang.org/doc/install) in any environment (e.g. Linux, MacOS, Windows), as follows:

```
go get -u github.com/alphasoc/flightsim/...
```

## Running Network Flight Simulator

Upon installation, test flightsim as follows:

```
$ flightsim --help

AlphaSOC Network Flight Simulator™ (https://github.com/alphasoc/flightsim)

flightsim is an application which generates malicious network traffic for security
teams to evaluate security controls (e.g. firewalls) and ensure that monitoring tools
are able to detect malicious traffic.

Usage:
  flightsim <command> [arguments]

Available Commands:
  run         Run all modules, or a particular module
  version     Prints the version number

Cheatsheet:
  flightsim run                Run all the modules
  flightsim run c2             Simulate C2 traffic
  flightsim run c2:trickbot    Simulate C2 traffic for the TrickBot family
```

The utility runs individual modules to generate malicious traffic. To perform all available tests, simply use `flightsim run` which will generate traffic using the first available non-loopback network interface. **Note:** when running many modules, flightsim will gather destination addresses from the AlphaSOC API, so requires egress Internet access.

To list the available modules, use `flightsim run --help`. To execute a particular test, use `flightsim run <module>`, as below.

```
$ flightsim run --help
usage: flightsim run [flags] [modules]

To run all available simulators, call:

    flightsim run

 To run a specific module:

    flightsim run c2

Available modules:

	c2, dga, scan, sink, spambot, tunnel

Available flags:
  -dry
    	print actions without performing any network activity
  -fast
    	reduce sleep intervals between simulation events
  -iface string
    	network interface or local IP address to use
  -size int
    	number of hosts generated for each simulator

$ flightsim run dga

AlphaSOC Network Flight Simulator™  (https://github.com/alphasoc/flightsim)
The IP address of the network interface is 172.20.10.2
The current time is 17-Sep-19 11:59:38

11:59:38 [dga] Generating list of DGA domains
11:59:38 [dga] Resolving slvoody.top
11:59:39 [dga] Resolving zwpajbp.com
11:59:40 [dga] Resolving moijbvx.top
11:59:41 [dga] Resolving yxxatfi.info
11:59:42 [dga] Resolving sbyzqpo.xyz
11:59:43 [dga] Resolving polmhgd.space
11:59:44 [dga] Resolving aqfarux.space
11:59:46 [dga] Resolving zxfkbzr.net
11:59:47 [dga] Resolving bbctlvx.net
11:59:48 [dga] Resolving fwzklyf.biz
11:59:49 [dga] Resolving gwtysmm.com
11:59:50 [dga] Resolving hnrqmuy.biz
11:59:51 [dga] Resolving glaxjlc.net
11:59:52 [dga] Resolving pwdbdgb.biz
11:59:53 [dga] Resolving kutvpxo.top

All done! Check your SIEM for alerts using the timestamps and details above.
```

## Description of Modules

The modules packaged with the utility are listed in the table below.

| Module    | Description                                                                   |
| --------- | ----------------------------------------------------------------------------- |
| `c2`      | Generates a list of C2 destinations and generates DNS and IP traffic to each  |
| `dga`     | Simulates DGA traffic using random labels and top-level domains               |
| `scan`    | Performs a port scan to random RFC 5737 addresses using common ports          |
| `sink`    | Connects to random sinkholed destinations run by security providers           |
| `spambot` | Resolves and connects to random Internet SMTP servers to simulate a spam bot  |
| `tunnel`  | Generates DNS tunneling requests to \*.sandbox.alphasoc.xyz                   |
