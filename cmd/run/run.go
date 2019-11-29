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
	fast   bool
	dryRun bool
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

	extIP, err := utils.ExternalIP(*ifaceName)
	if err != nil {
		return err
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

	return run(sims, extIP, *size)
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
	NumOfHosts int
	HeaderMsg  string
	HostMsg    string
	Timeout    time.Duration
	// FailMsg    string
	SuccessMsg string
}

func (m *Module) FormatHost(host string) string {
	if m.Pipeline == PipelineDNS {
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
	// 	HostMsg:    "Resolving %s via ns1.sandbox.alphasoc.xyz",
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
		Module:     simulator.NewTunnel(),
		Name:       "tunnel",
		Pipeline:   PipelineDNS,
		NumOfHosts: 25,
		// HeaderMsg:  "Preparing DNS tunnel hostnames",
		HostMsg: "Simulating DNS tunneling via *.%s",
		Timeout: 10 * time.Second,
	},
	Module{
		Module:     simulator.NewTorSimulator(),
		Name:       "tor",
		Pipeline:   PipelineDNS,
		NumOfHosts: 5,
		HeaderMsg:  "Preparing Tor connection",
		HostMsg:    "Connecting to %s",
		SuccessMsg: "Success! Tor use is permitted in this environment",
		Timeout:    10 * time.Second,
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

func run(sims []*Simulation, extIP net.IP, size int) error {
	printWelcome(extIP.String())
	printHeader()
	for simN, sim := range sims {
		var isSuccessfull bool = false
		err := sim.Init()
		if err != nil {
			printMsg(sim, "ERROR: "+fmt.Sprint(err))
		} else {
			printMsg(sim, sim.HeaderMsg)

			numOfHosts := size
			if numOfHosts == 0 {
				numOfHosts = sim.Module.NumOfHosts
			}

			hosts, err := sim.Module.Hosts(sim.Scope, numOfHosts)
			if err != nil {
				printMsg(sim, "failed: "+err.Error())
				continue
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						printMsg(sim, "ERROR: "+fmt.Sprint(r))
					}
				}()

				for hostN, host := range hosts {
					printMsg(sim, sim.FormatHost(host))

					if !dryRun {
						ctx, cancel := context.WithTimeout(context.Background(), sim.Timeout)
						if err := sim.Module.Simulate(ctx, extIP, host); err != nil {
							// TODO: some module can return custom messages (e.g. hijack)
							// and "ERROR" prefix shouldn't be printed then
							printMsg(sim, "ERROR: "+err.Error())
						} else {
							isSuccessfull = true
						}

						// wait until context expires (unless fast mode or very last iteration)
						if !fast && ((simN < len(sims)-1) || (hostN < len(hosts)-1)) {
							<-ctx.Done()
						}

						cancel()
					}
				}
			}()
			if isSuccessfull {
				printMsg(sim, sim.SuccessMsg)
			}
		}
		sim.Cleanup()
	}

	printGoodbye()
	return nil
}
