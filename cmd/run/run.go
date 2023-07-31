package run

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/alphasoc/flightsim/simulator"
	"github.com/alphasoc/flightsim/utils"
	"github.com/alphasoc/flightsim/wisdom"
)

var (
	fast   bool
	dryRun bool
)

var allModuleNames []string = func() []string {
	var (
		names []string
		seen  = make(map[string]bool)
	)

	for _, m := range allModules {
		if !seen[m.Name] && !m.Experimental {
			names = append(names, m.Name)
			seen[m.Name] = true
		}
	}

	sort.Strings(names)
	return names
}()

var usage = `usage: flightsim run [flags] [modules]

To run all available modules, call:

    flightsim run

 To run a specific module:

    flightsim run c2

Available modules:

	%s

Available flags:
`

func RunCmd(args []string) error {
	cmdLine := flag.NewFlagSet("run", flag.ExitOnError)
	cmdLine.BoolVar(&fast, "fast", false, "reduce sleep intervals between simulation events")
	cmdLine.BoolVar(&dryRun, "dry", false, "print actions without performing any network activity")
	ifaceName := cmdLine.String("iface", "", "network interface or local IP address to use")
	size := cmdLine.Int("size", 0, "number of hosts generated for each simulator")

	cmdLine.Usage = func() {
		fmt.Fprintf(cmdLine.Output(), usage, strings.Join(allModuleNames, ", "))
		cmdLine.PrintDefaults()
	}
	cmdLine.Parse(args)

	modules := cmdLine.Args()
	if len(modules) == 0 {
		modules = allModuleNames
	}

	if *size < 0 {
		*size = 0
	}
	// Grab a "usable" IP address for ifaceName.
	bindIP, err := utils.UsableIP(*ifaceName)
	if err != nil {
		return fmt.Errorf("Unable to determine usable IP address for '%v': %v", *ifaceName, err)
	}
	bind := simulator.BindAddr{Addr: bindIP}
	if *ifaceName != "" {
		bind.UserSet = true
	}

	sims, err := selectSimulations(modules)
	if err != nil {
		return err
	}

	// if fast {
	// 	for i := range sims {
	// 		sims[i].Timeout = 100 * time.Millisecond
	// 	}
	// }
	return run(sims, bind, *size)
}

func selectSimulations(names []string) ([]*Simulation, error) {
	var res []*Simulation

	for _, name := range names {
		scope := ""
		if i := strings.IndexByte(name, ':'); i >= 0 {
			scope = name[i+1:]
			name = name[:i]
		}

		var found bool
		for _, m := range allModules {
			if name == m.Name || strings.HasPrefix(m.Name, name+"-") {
				res = append(res, &Simulation{
					Module: m,
					Scope:  scope,
					Size:   0, // TODO
				})
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown module: %s", name)
		}
	}

	return res, nil
}

type Pipeline string

const (
	PipelineDNS Pipeline = "dns"
	PipelineIP           = "ip"
)

type Module struct {
	simulator.Module
	Name         string
	Pipeline     Pipeline
	Experimental bool
	NumOfHosts   int
	HeaderMsg    string
	HostMsg      string
	Timeout      time.Duration
	// FailMsg    string
	SuccessMsg string
	// False by default.  If true, don't wait until Timeout between simulation
	// runs of this module.
	Fast bool
}

func (m *Module) FormatHost(host string) string {
	if m.Pipeline == PipelineDNS {
		h, _, _ := net.SplitHostPort(host)
		if h != "" {
			host = h
		}
	}
	// Check if the simulator module implements the HostMsgFormatter interface.
	if hostMsgFormatter, ok := m.Module.(simulator.HostMsgFormatter); ok {
		return hostMsgFormatter.HostMsg(host)
	}
	f := m.HostMsg
	if f == "" {
		switch m.Pipeline {
		case PipelineDNS:
			f = "Resolving %s"
		case PipelineIP:
			f = "Connecting to %s"
		}
	}

	return fmt.Sprintf(f, host)
}

var allModules = []Module{
	Module{
		Module:     simulator.CreateModule(wisdom.NewWisdomHosts("c2", wisdom.HostTypeDNS), new(simulator.DNSResolveSimulator)),
		Name:       "c2",
		Pipeline:   PipelineDNS,
		NumOfHosts: 5,
		HeaderMsg:  "Preparing a random sample of C2 domains",
		Timeout:    1 * time.Second,
	},
	Module{
		Module:     simulator.CreateModule(wisdom.NewWisdomHosts("c2", wisdom.HostTypeIP), new(simulator.TCPConnectSimulator)),
		Name:       "c2",
		Pipeline:   PipelineIP,
		NumOfHosts: 5,
		HeaderMsg:  "Preparing a random sample of C2 IP:port pairs",
		Timeout:    1 * time.Second,
	},
	Module{
		Module:     simulator.NewDGA(),
		Name:       "dga",
		Pipeline:   PipelineDNS,
		NumOfHosts: 15,
		HeaderMsg:  "Generating a list of DGA domains",
		Timeout:    1 * time.Second,
	},
	// Module{
	// 	Module:     simulator.NewHijack(),
	// 	Name:       "hijack",
	// 	Pipeline:   PipelineDNS,
	// 	NumOfHosts: 1,
	// 	HeaderMsg:  "",
	// 	HostMsg:    "Resolving %s via dns.sandbox-services.alphasoc.xyz",
	// 	Timeout:    1 * time.Second,
	// 	// FailMsg:    "Test failed (queries to arbitrary DNS servers are blocked)",
	// 	SuccessMsg: "Success! DNS hijacking is possible in this environment",
	// },
	Module{
		Module:     simulator.NewPortScan(),
		Name:       "scan",
		Pipeline:   PipelineIP,
		NumOfHosts: 10,
		HeaderMsg:  "Preparing a random sample of RFC 5737 destinations",
		HostMsg:    "Port scanning %s",
		Timeout:    3 * time.Second,
	},
	Module{
		Module:     simulator.CreateModule(wisdom.NewWisdomHosts("sinkholed", wisdom.HostTypeDNS), new(simulator.DNSResolveSimulator)),
		Name:       "sink",
		Pipeline:   PipelineDNS,
		NumOfHosts: 5,
		HeaderMsg:  "Preparing a random sample of sinkholed domains",
		Timeout:    1 * time.Second,
	},
	Module{
		Module:     simulator.CreateModule(wisdom.NewWisdomHosts("sinkholed", wisdom.HostTypeIP), new(simulator.TCPConnectSimulator)),
		Name:       "sink",
		Pipeline:   PipelineIP,
		NumOfHosts: 5,
		HeaderMsg:  "Preparing a random sample of sinkholed IP:port pairs",
		Timeout:    1 * time.Second,
	},
	Module{
		Module:     simulator.NewSpambot(),
		Name:       "spambot",
		Pipeline:   PipelineIP,
		NumOfHosts: 10,
		HeaderMsg:  "Preparing a random sample of Internet mail servers",
		Timeout:    1 * time.Second,
	},
	Module{
		Module:   simulator.NewTunnel(),
		Name:     "tunnel-dns",
		Pipeline: PipelineDNS,
		HostMsg:  "Simulating DNS tunneling via *.%s",
		Timeout:  10 * time.Second,
	},
	Module{
		Module:     simulator.CreateModule(wisdom.NewWisdomHosts("cryptomining", wisdom.HostTypeIP), simulator.NewStratumMiner()),
		Name:       "miner",
		Pipeline:   PipelineIP,
		NumOfHosts: 5,
		HeaderMsg:  "Preparing a random sample of cryptomining IP:port pairs",
		Timeout:    1 * time.Second,
	},
	Module{
		Module:       simulator.NewTorSimulator(),
		Name:         "tor",
		Pipeline:     PipelineDNS,
		Experimental: true,
		NumOfHosts:   5,
		HeaderMsg:    "Preparing Tor connection",
		HostMsg:      "Connecting to %s",
		SuccessMsg:   "Tor use is permitted in this environment",
		// FailMsg:    "Couldn't contact Tor network",
		Timeout: 10 * time.Second,
	},
	Module{
		Module:     simulator.NewICMPtunnel(),
		Name:       "tunnel-icmp",
		Pipeline:   PipelineDNS,
		NumOfHosts: 1,
		HostMsg:    "Simulating ICMP tunneling via %s",
		Timeout:    20 * time.Second,
	},
	Module{
		Module:     simulator.CreateModule(wisdom.NewWisdomHosts("imposter", wisdom.HostTypeDNS), new(simulator.DNSResolveSimulator)),
		Name:       "imposter",
		Pipeline:   PipelineDNS,
		NumOfHosts: 5,
		HeaderMsg:  "Resolving random imposter domains",
		Timeout:    1 * time.Second,
	},
	Module{
		Module:     simulator.NewOAST(),
		Name:       "oast",
		Pipeline:   PipelineDNS,
		NumOfHosts: 1,
		HeaderMsg:  "Preparing to simulate OAST traffic",
		Timeout:    3 * time.Second,
	},
	Module{
		Module:    simulator.NewSSHTransfer(),
		Name:      "ssh-transfer",
		Pipeline:  PipelineIP,
		HeaderMsg: "Preparing to send randomly generated data to a standard SSH port",
		Timeout:   5 * time.Minute,
		Fast:      true,
	},
	Module{
		Module:    simulator.NewSSHExfil(),
		Name:      "ssh-exfil",
		Pipeline:  PipelineIP,
		HeaderMsg: "Preparing to send randomly generated data to a non-standard SSH port",
		Timeout:   5 * time.Minute,
		Fast:      true,
	},
	Module{
		Module:     simulator.CreateModule(wisdom.NewWisdomHosts("irc", wisdom.HostTypeDNS), simulator.NewIRCClient()),
		Name:       "irc",
		Pipeline:   PipelineDNS,
		NumOfHosts: 5,
		HeaderMsg:  "Preparing a random sample of IRC server domains",
		Timeout:    5 * time.Second,
		Fast:       true,
		HostMsg:    "Simulating IRC traffic to %s",
	},
	Module{
		Module:     simulator.CreateModule(wisdom.NewWisdomHosts("irc", wisdom.HostTypeIP), simulator.NewIRCClient()),
		Name:       "irc",
		Pipeline:   PipelineIP,
		NumOfHosts: 5,
		HeaderMsg:  "Preparing a random sample of IRC server IP:port pairs",
		Timeout:    5 * time.Second,
		Fast:       true,
		HostMsg:    "Simulating IRC traffic to %s",
	},
	Module{
		Module:    simulator.NewTelegramBot(),
		Name:      "telegram-bot",
		Pipeline:  PipelineDNS,
		HeaderMsg: "Preparing to simulate Telegram bot traffic",
		Timeout:   3 * time.Second,
		HostMsg:   "Simulating Telegram Bot API traffic to %s",
	},
	Module{
		Module:    simulator.NewCleartextProtocolSimulator(),
		Name:      "cleartext",
		Pipeline:  PipelineIP,
		HeaderMsg: "Preparing to simulate cleartext protocol traffic",
		Timeout:   3 * time.Second,
		HostMsg:   "Sending random data to %s",
	},
}

type Simulation struct {
	Module
	Scope string
	Size  int
}

func (s *Simulation) Name() string {
	name := s.Module.Name
	if s.Scope != "" {
		name += ":" + s.Scope
	}
	return name
}

const (
	msgPrefixErrorInit    = "FATAL: Couldn't start the module: "
	msgPrefixErrorRecover = "FATAL: Module terminated: "
)

// getDefaultDNSIntf runs a DNS probe using default system resolver and returns the IP of
// the interface used, or an empty string.  Thanks @tg.
func getDefaultDNSIntf() string {
	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var defaultDNSServer string
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			// we can capture the address here
			defaultDNSServer = address
			// still use the default dialer
			var d net.Dialer
			return d.DialContext(ctx, network, address)
		},
	}
	_, err := r.LookupHost(ctx, "alphasoc.com")
	if err != nil {
		return ""
	}
	// In cases where r.Dial() is not invoked, defaultDNSServer will be "", so don't bother
	// continuing with detection (ie. on Windows).
	if defaultDNSServer == "" {
		return ""
	}
	conn, err := net.DialTimeout("udp", defaultDNSServer, timeout)
	if err != nil {
		return ""
	}
	dnsIntfIP, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return ""
	}
	return dnsIntfIP
}

func run(sims []*Simulation, bind simulator.BindAddr, size int) error {
	// If user override on iface, both IP and DNS traffic will flow through bind.Addr.
	// NOTE: not passing the DNS server to printWelcome(), as it may be confusing in cases
	// where there are multiple nameservers configured (ie. resolver errors will carry
	// the address of the last queried nameserver).
	defaultDNSIntfIP := getDefaultDNSIntf()
	if bind.UserSet {
		printWelcome(bind.String(), bind.String())
	} else {
		// NOTE: defaultDNSIntfIP _may_ be "".
		printWelcome(bind.String(), defaultDNSIntfIP)
	}
	printHeader()

	for simN, sim := range sims {
		fmt.Print("\n")

		okHosts := 0
		err := sim.Init(bind)
		if err != nil {
			printMsg(sim, msgPrefixErrorInit+fmt.Sprint(err))
		} else {
			printMsg(sim, sim.HeaderMsg)

			numOfHosts := size
			if numOfHosts == 0 {
				numOfHosts = sim.Module.NumOfHosts
			}

			hosts, err := sim.Module.Hosts(sim.Scope, numOfHosts)
			if err != nil {
				printMsg(sim, msgPrefixErrorInit+err.Error())
				continue
			}

			// Pick random hosts if we have more than we need
			if numOfHosts > 0 && len(hosts) > numOfHosts {
				newHosts := make([]string, numOfHosts)
				for n, k := range rand.Perm(len(hosts))[:numOfHosts] {
					newHosts[n] = hosts[k]
				}
				hosts = newHosts
			}

			// Wrap module execution in a function, so we can recover from panics
			func() {
				defer func() {
					if r := recover(); r != nil {
						printMsg(sim, msgPrefixErrorRecover+fmt.Sprint(r))
					}
				}()

				for hostN, host := range hosts {
					printMsg(sim, sim.FormatHost(host))

					if !dryRun {
						ctx, cancel := context.WithTimeout(context.Background(), sim.Timeout)
						if err := sim.Module.Simulate(ctx, host); err != nil {
							// TODO: some module can return custom messages (e.g. hijack)
							// and "ERROR" prefix shouldn't be printed then
							printMsg(sim, fmt.Sprintf("ERROR: %s: %s", host, err.Error()))
						} else {
							okHosts++
						}

						// Wait until context expires, unless fast global mode,
						// fast module (default false) or very last iteration.
						if !(fast || sim.Fast) && ((simN < len(sims)-1) || (hostN < len(hosts)-1)) {
							<-ctx.Done()
						}

						cancel()
					}
				}
			}()

			msg := fmt.Sprintf("Done (%d/%d)", okHosts, len(hosts))
			if okHosts > 0 && sim.SuccessMsg != "" {
				msg = fmt.Sprintf("%s: %s", msg, sim.SuccessMsg)
			}

			printMsg(sim, msg)
		}
		sim.Cleanup()
	}

	printGoodbye()
	return nil
}
