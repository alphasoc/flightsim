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

	c2, dga, miner, scan, sink, spambot, tunnel-dns, tunnel-icmp

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
The current time is 23-Jan-20 11:33:21

11:33:21 [dga] Generating a list of DGA domains
11:33:21 [dga] Resolving nurqatp.space
11:33:22 [dga] Resolving uahscqe.top
11:33:23 [dga] Resolving asimazf.biz
11:33:24 [dga] Resolving phxeohj.biz
11:33:25 [dga] Resolving crgwsoe.biz
11:33:26 [dga] Resolving sazafls.biz
11:33:27 [dga] Resolving gljyxdv.space
11:33:28 [dga] Resolving eiontgl.top
11:33:29 [dga] Resolving pqjseqc.top
11:33:30 [dga] Resolving mamsnmu.biz
11:33:31 [dga] Resolving ntettqn.top
11:33:32 [dga] Resolving niyvbvg.top
11:33:33 [dga] Resolving bxgqonb.biz
11:33:34 [dga] Resolving encggla.top
11:33:35 [dga] Resolving qphfoxn.biz
11:33:35 [dga] Done (15/15)

All done! Check your SIEM for alerts using the timestamps and details above.
```

## Description of Modules

The modules packaged with the utility are listed in the table below.

| Module        | Description                                                                   |
| ------------- | ----------------------------------------------------------------------------- |
| `c2`          | Generates both DNS and IP traffic to a random list of known C2 destinations   |
| `dga`         | Simulates DGA traffic using random labels and top-level domains               |
| `miner`       | Generates Stratum mining protocol traffic to known cryptomining pools         |
| `scan`        | Performs a port scan of random RFC 5737 addresses using common TCP ports      |
| `sink`        | Connects to known sinkholed destinations run by security researchers          |
| `spambot`     | Resolves and connects to random Internet SMTP servers to simulate a spam bot  |
| `tunnel-dns`  | Generates DNS tunneling requests to \*.sandbox.alphasoc.xyz                   |
| `tunnel-icmp` | Generates ICMP tunneling traffic to an Internet service operated by AlphaSOC  |
