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
  flightsim [command]

Available Commands:
  help        Help about any command
  run         Run all simulators (default) or a particular test
  version     Print version and exit

Flags:
  -h, --help   help for flightsim

Use "flightsim [command] --help" for more information about a command
```

The utility runs individual modules to generate malicious traffic. To perform all available tests, simply use `flightsim run` which will generate traffic using the first available non-loopback network interface. **NB:** when running the C2 modules, flightsim will gather current C2 addresses from the [Cybercrime Tracker](http://cybercrime-tracker.net) and AlphaSOC API, so requires egress Internet access.

To list the available modules, use `flightsim run --help`. To execute a particular test, use `flightsim run <module>`, as below.

```
$ flightsim run --help
Run all simulators (default) or a particular test

Usage:
  flightsim run [c2-dns|c2-ip|dga|hijack|scan|sink|spambot|tor|tunnel] [flags]

Flags:
  -n,                      number of hosts generated for each simulator (default 10)
      --fast               run simulator fast without sleep intervals
  -h, --help               help for run
  -i, --interface string   network interface to use

$ flightsim run dga

AlphaSOC Network Flight Simulator™ (https://github.com/alphasoc/flightsim)
The IP address of the network interface is 172.31.84.103
The current time is 10-Jan-18 09:30:28

Time      Module   Description
--------------------------------------------------------------------------------
09:30:28  dga      Starting
09:30:28  dga      Generating list of DGA domains
09:30:30  dga      Resolving rdumomx.xyz
09:30:31  dga      Resolving rdumomx.biz
09:30:31  dga      Resolving rdumomx.top
09:30:32  dga      Resolving qtovmrn.xyz
09:30:32  dga      Resolving qtovmrn.biz
09:30:33  dga      Resolving qtovmrn.top
09:30:33  dga      Resolving pbuzkkk.xyz
09:30:34  dga      Resolving pbuzkkk.biz
09:30:34  dga      Resolving pbuzkkk.top
09:30:35  dga      Resolving wfoheoz.xyz
09:30:35  dga      Resolving wfoheoz.biz
09:30:36  dga      Resolving wfoheoz.top
09:30:36  dga      Resolving lhecftf.xyz
09:30:37  dga      Resolving lhecftf.biz
09:30:37  dga      Resolving lhecftf.top
09:30:38  dga      Finished

All done! Check your SIEM for alerts using the timestamps and details above.
```

## Description of Modules

The modules packaged with the utility are listed in the table below.

| Module    | Description                                                                   |
| --------- | ----------------------------------------------------------------------------- |
| `c2-dns`  | Generates a list of current C2 destinations and performs DNS requests to each |
| `c2-ip`   | Connects to 10 random current C2 IP:port pairs to simulate egress sessions    |
| `dga`     | Simulates DGA traffic using random labels and top-level domains               |
| `hijack`  | Tests for DNS hijacking support via ns1.sandbox.alphasoc.xyz                  |
| `scan`    | Performs a port scan of 10 random RFC 1918 addresses using common ports       |
| `sink`    | Connects to 10 random sinkholed destinations run by security providers        |
| `spambot` | Resolves and connects to random Internet SMTP servers to simulate a spam bot  |
| `tunnel`  | Generates DNS tunneling requests to \*.sandbox.alphasoc.xyz                   |
