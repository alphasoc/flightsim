package run

import (
	"context"
	"flag"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/alphasoc/flightsim/simulator"
	"github.com/alphasoc/flightsim/utils"
	"github.com/alphasoc/flightsim/wisdom"
)

var (
	fast      bool
	size      int
	ifaceName string
)

var allModuleNames []string = func() []string {
	var (
		names []string
		seen  = make(map[string]bool)
	)

	for _, m := range allModules {
		if !seen[m.Name] {
			names = append(names, m.Name)
			seen[m.Name] = true
		}
	}

	sort.Strings(names)
	return names
}()

var usage = `usage: flightsim run [flags] [modules]

To run all available simulators, call:

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
	cmdLine.StringVar(&ifaceName, "iface", "", "network interface to use")
	cmdLine.IntVar(&size, "size", 10, "number of hosts generated for each simulator")

	cmdLine.Usage = func() {
		fmt.Fprintf(cmdLine.Output(), usage, strings.Join(allModuleNames, ", "))
		cmdLine.PrintDefaults()
	}
	cmdLine.Parse(args)

	modules := cmdLine.Args()
	if len(modules) == 0 {
		modules = allModuleNames
	}

	if size <= 0 {
		return fmt.Errorf("size must be positive")
	}

	extIP, err := utils.ExternalIP(ifaceName)
	if err != nil {
		return err
	}

	sims, err := selectSimulations(modules)
	if err != nil {
		return err
	}

	if fast {
		for i := range sims {
			sims[i].Timeout = 100 * time.Millisecond
		}
	}

	return run(sims, extIP)
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
			if m.Name == name {
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
	Name       string
	Pipeline   Pipeline
	HeaderMsg  string
	HidePort   bool
	HostMsg    string
	Timeout    time.Duration
	FailMsg    string
	SuccessMsg string
}

func (m *Module) FormatHost(host string) string {
	if m.HidePort || m.Pipeline == PipelineDNS {
		h, _, _ := net.SplitHostPort(host)
		if h != "" {
			host = h
		}
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
		Module:    simulator.CreateModule(wisdom.NewWisdomHosts("c2", "dns"), new(simulator.DNSResolveSimulator)),
		Name:      "c2",
		Pipeline:  PipelineDNS,
		HeaderMsg: "Preparing random sample of C2 domains",
		Timeout:   1 * time.Second,
	},
	Module{
		Module:    simulator.CreateModule(wisdom.NewWisdomHosts("c2", "ip"), new(simulator.TCPConnectSimulator)),
		Name:      "c2",
		Pipeline:  PipelineIP,
		HeaderMsg: "Preparing random sample of C2 IP:port pairs",
		Timeout:   1 * time.Second,
	},
	Module{
		Module:    simulator.NewDGA(),
		Name:      "dga",
		Pipeline:  PipelineDNS,
		HeaderMsg: "Generating list of DGA domains",
		Timeout:   1 * time.Second,
	},
	Module{
		Module:     simulator.NewHijack(),
		Name:       "hijack",
		Pipeline:   PipelineDNS,
		HeaderMsg:  "",
		HostMsg:    "Resolving %s via ns1.sandbox.alphasoc.xyz",
		Timeout:    1 * time.Second,
		FailMsg:    "Test failed (queries to arbitrary DNS servers are blocked)",
		SuccessMsg: "Success! DNS hijacking is possible in this environment",
	},
	Module{
		Module:    simulator.NewPortScan(),
		Name:      "scan",
		Pipeline:  PipelineIP,
		HeaderMsg: "Preparing random sample of RFC 5737 destinations",
		HostMsg:   "Port scanning %s",
		HidePort:  true,
		Timeout:   30 * time.Millisecond,
	},
	Module{
		Module:    simulator.CreateModule(wisdom.NewWisdomHosts("sinkholed", "dns"), new(simulator.DNSResolveSimulator)),
		Name:      "sink",
		Pipeline:  PipelineDNS,
		HeaderMsg: "Preparing random sample of sinkhole IP:port pairs",
		Timeout:   1 * time.Second,
	},
	Module{
		Module:    simulator.CreateModule(wisdom.NewWisdomHosts("sinkholed", "ip"), new(simulator.TCPConnectSimulator)),
		Name:      "sink",
		Pipeline:  PipelineIP,
		HeaderMsg: "Preparing random sample of sinkhole IP:port pairs",
		Timeout:   1 * time.Second,
	},
	Module{
		Module:    simulator.NewSpambot(),
		Name:      "spambot",
		Pipeline:  PipelineIP,
		HeaderMsg: "Preparing random sample of Internet mail servers",
		Timeout:   1 * time.Second,
	},
	Module{
		Module:    simulator.NewTunnel(),
		Name:      "tunnel",
		Pipeline:  PipelineDNS,
		HeaderMsg: "Preparing DNS tunnel hostnames",
		Timeout:   1 * time.Second,
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

func run(sims []*Simulation, extIP net.IP) error {
	printWelcome(extIP.String())
	printHeader()
	for _, sim := range sims {
		// printMsg(sim, "Starting")
		printMsg(sim, sim.HeaderMsg)

		hosts, err := sim.Module.Hosts(sim.Scope, size)
		if err != nil {
			printMsg(sim, "failed: "+err.Error())
			continue
		}

		var prevMsg string
		for _, host := range hosts {
			msg := sim.FormatHost(host)
			if prevMsg != msg {
				printMsg(sim, msg)
			}
			prevMsg = msg

			ctx, cancel := context.WithTimeout(context.Background(), sim.Timeout)
			if err := sim.Module.Simulate(ctx, extIP, host); err != nil {
				printMsg(sim, sim.FailMsg)
			} else {
				printMsg(sim, sim.SuccessMsg)
			}

			if !fast {
				<-ctx.Done()
			}
			cancel()
		}
		// printMsg(sim, "Finished")
	}

	printGoodbye()
	return nil
}
