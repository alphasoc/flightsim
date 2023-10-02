# Network Flight Simulator

**flightsim** is a lightweight utility used to generate malicious network traffic and help security teams to evaluate security controls and network visibility. The tool performs tests to simulate DNS tunneling, DGA traffic, requests to known active C2 destinations, and other suspicious traffic patterns.

## Installation

Download the latest flightsim binary for your OS from the [GitHub Releases](https://github.com/alphasoc/flightsim/releases) page. Alternatively, the utility can be built using [Golang](https://golang.org/doc/install) in any environment (e.g. Linux, MacOS, Windows), as follows:

```
go install github.com/alphasoc/flightsim@latest
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
```

The utility runs individual modules to generate malicious traffic. To perform all available tests, simply use `flightsim run` which will generate traffic using the first available non-loopback network interface. **Note:** when running many modules, flightsim will gather destination addresses from the AlphaSOC API, so requires egress Internet access.

To list the available modules, use `flightsim run --help`. To execute a particular test, use `flightsim run <module>`, as below.

```
$ flightsim run --help
usage: flightsim run [flags] [modules]

To run all available modules, call:

    flightsim run

 To run a specific module:

    flightsim run c2

Available modules:

        c2, dga, imposter, miner, scan, sink, spambot, ssh-exfil, ssh-transfer, tunnel-dns, tunnel-icmp

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
The address of the network interface for IP traffic is 192.168.220.38
The address of the network interface for DNS queries is 192.168.220.38
The current time is 26-Oct-21 17:28:51

17:28:51 [dga] Generating a list of DGA domains
17:28:51 [dga] Resolving 6kauziij.com
17:28:52 [dga] Resolving paxeo0jk.biz
17:28:53 [dga] Resolving iuuub8al.biz
17:28:54 [dga] Resolving bxsei3nj.com
17:28:55 [dga] Resolving zbwltf1h.space
17:28:56 [dga] Resolving yoze7avi.com
17:28:57 [dga] Resolving ijax8aqw.space
17:28:58 [dga] Resolving wwrjyj4l.space
17:28:59 [dga] Resolving uioc5hky.com
17:29:00 [dga] Resolving lcwdji5t.biz
17:29:01 [dga] Resolving zluwcb4h.biz
17:29:02 [dga] Resolving 8jodcvhj.space
17:29:03 [dga] Resolving ju5haxur.com
17:29:04 [dga] Resolving ivthu2dl.biz
17:29:05 [dga] Resolving ha0bsxft.com
17:29:05 [dga] Done (15/15)

All done! Check your SIEM for alerts using the timestamps and details above.
```

The utility also has a `get` command which can be used to query information that can later be used with the simulation modules. At present, a list of C2 families can be obtained to be used with the C2 module. To see how to use the `get` command, run `flightsim get -h` as below.

```
$ flightsim get -h

AlphaSOC Network Flight Simulator™  (https://github.com/alphasoc/flightsim)
The current time is 26-Oct-21 17:42:23

usage: flightsim get [flags] element:category

Available elements:

        families

Available categories:

        c2

Available flags:
```

To get a list of C2 families, run:

```
$ flightsim get families:c2

AlphaSOC Network Flight Simulator™  (https://github.com/alphasoc/flightsim)
The current time is 16-Nov-21 11:16:51

11:16:51 [families:c2] Fetching c2 families
11:16:55 [families:c2] Adwind, Agent Tesla, Amadey, AsyncRAT, AZORult, BASHLITE, BazarBackdoor, BlackNET RAT, Cobalt Strike, Collector Stealer, CryptBot, DarkComet, DiamondFox, Dridex, Emotet, Gozi, IcedID, Kimsuky, KPOT Stealer, LokiBot, Mirai, NanoCore RAT, njRAT, Oski Stealer, Pony, Predator the Thief, Quakbot, RedLine, RedLine Stealer, Remcos RAT, Smoke Loader, Taurus, TrickBot, XtremeRAT, Zloader
11:16:55 [families:c2] Fetched 35 c2 families

All done!
```

## Description of Modules

The modules packaged with the utility are listed in the table below.

| Module        | Description                                                                      |
| ------------- | -------------------------------------------------------------------------------- |
| `c2`          | Generates both DNS and IP traffic to a random list of known C2 destinations      |
| `cleartext`   | Generates random cleartext traffic to an Internet service operated by AlphaSOC   |
| `dga`         | Simulates DGA traffic using random labels and top-level domains                  |
| `imposter`    | Generates DNS traffic to a list of imposter domains                              |
| `irc`         | Connects to a random list of public IRC servers                                  |
| `miner`       | Generates Stratum mining protocol traffic to known cryptomining pools            |
| `oast`        | Simulates out-of-band application security testing (OAST) traffic                |
| `scan`        | Performs a port scan of random RFC 5737 addresses using common TCP ports         |
| `sink`        | Connects to known sinkholed destinations run by security researchers             |
| `spambot`     | Resolves and connects to random Internet SMTP servers to simulate a spam bot     |
| `ssh-exfil`   | Simulates an SSH file transfer to a service running on a non-standard SSH port   |
| `ssh-transfer`| Simulates an SSH file transfer to a service running on an SSH port               |
| `telegram-bot`| Generates Telegram Bot API traffic using a random or provided token              |
| `tunnel-dns`  | Generates DNS tunneling requests to \*.sandbox.alphasoc.xyz                      |
| `tunnel-icmp` | Generates ICMP tunneling traffic to an Internet service operated by AlphaSOC     |

