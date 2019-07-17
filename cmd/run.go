package cmd

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alphasoc/flightsim/simulator"
	"github.com/alphasoc/flightsim/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	fast           bool
	size           int
	ifaceName      string
	simulatorNames = []string{"c2-dns", "c2-ip", "dga", "hijack", "scan", "sink", "spambot", "tunnel"}
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("run [%s]", strings.Join(simulatorNames, "|")),
		Short: "Run all simulators (default) or a particular test",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if !utils.StringsContains(simulatorNames, arg) {
					return fmt.Errorf("simulator %s not recognized", arg)
				}
			}

			if len(args) > 0 {
				simulatorNames = args
			}

			if size <= 0 {
				return fmt.Errorf("n must be positive")
			}

			extIP, err := utils.ExternalIP(ifaceName)
			if err != nil {
				return err
			}

			simulators := selectSimulators(simulatorNames)

			if fast {
				for i := range simulators {
					simulators[i].timeout = 100 * time.Millisecond
				}
			}

			run(simulators, extIP)
			return nil
		},
	}

	cmd.Flags().BoolVar(&fast, "fast", false, "run simulator fast without sleep intervals")
	cmd.Flags().IntVarP(&size, "", "n", 10, "number of hosts generated for each simulator")
	cmd.Flags().StringVarP(&ifaceName, "interface", "i", "", "network interface to use")
	return cmd
}

func selectSimulators(names []string) []simulatorInfo {
	var simulators []simulatorInfo
	for _, s := range allsimualtors {
		if utils.StringsContains(names, s.name) {
			simulators = append(simulators, s)
		}
	}
	return simulators
}

type simulatorInfo struct {
	name        string
	infoHeaders []string
	infoRun     string
	s           simulator.Simulator
	timeout     time.Duration
	displayPort bool

	onError         string
	onSuccess       string
	breakOnNilError bool
}

var allsimualtors = []simulatorInfo{
	{
		"c2-dns",
		[]string{"Preparing random sample of current C2 domains"},
		"Resolving %s",
		simulator.NewC2DNS(),
		1 * time.Second,
		false,
		"",
		"",
		false,
	},
	{
		"c2-ip",
		[]string{"Preparing random sample of current C2 IP:port pairs"},
		"Connecting to %s",
		simulator.NewC2IP(),
		1 * time.Second,
		true,
		"",
		"",
		false,
	},
	{
		"dga",
		[]string{"Generating list of DGA domains"},
		"Resolving %s",
		simulator.NewDGA(),
		1 * time.Second,
		false,
		"",
		"",
		false,
	},
	{
		"hijack",
		nil,
		"Resolving %s via ns1.sandbox.alphasoc.xyz",
		simulator.NewHijack(),
		1 * time.Second,
		false,
		"Test failed (queries to arbitrary DNS servers are blocked)",
		"Success! DNS hijacking is possible in this environment",
		false,
	},
	{
		"scan",
		[]string{
			"Preparing random sample of RFC 5737 destinations",
			// "Preparing random sample of common TCP destination ports",
		},
		"Port scanning %s",
		simulator.NewPortScan(),
		30 * time.Millisecond,
		false,
		"",
		"",
		false,
	},
	{
		"sink",
		[]string{"Preparing random sample of current sinkhole IP:port pairs"},
		"Connecting to %s",
		simulator.NewSinkhole(),
		1 * time.Second,
		true,
		"",
		"",
		false,
	},
	{
		"spambot",
		[]string{
			"Preparing random sample of Internet mail servers",
		},
		"Connecting to %s",
		simulator.NewSpambot(),
		1 * time.Second,
		true,
		"",
		"",
		false,
	},
	/*
		{
			"tor",
			[]string{"Establishing Tor circuit"},
			"Connecting to %s exit note",
			simulator.NewTor(),
			1 * time.Second,
			true,
			"Test failed (unable to establish Tor circuit)",
			"Success! Tor use is permitted in this environment",
			true,
		},
	*/
	{
		"tunnel",
		[]string{"Preparing DNS tunnel hostnames"},
		"Resolving %s",
		simulator.NewTunnel(),
		1 * time.Second,
		false,
		"",
		"",
		false,
	},
}

func run(simulators []simulatorInfo, extIP net.IP) error {
	printWelcome(extIP.String())
	printHeader()
	for _, s := range simulators {
		printMsg(s.name, "Starting")
		printMsg(s.name, s.infoHeaders...)

		hosts, err := s.s.Hosts(size)
		if err != nil {
			printMsg(s.name, color.RedString("failed: ")+err.Error())
			continue
		}

		var prevHostname string
		for _, host := range hosts {
			hostname, _, err := net.SplitHostPort(host)
			if err != nil {
				hostname = host
			}

			// only print hostname when it has changed
			if prevHostname != hostname {
				if s.displayPort {
					printMsg(s.name, fmt.Sprintf(s.infoRun, host))
				} else {
					printMsg(s.name, fmt.Sprintf(s.infoRun, hostname))
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
			if err := s.s.Simulate(ctx, extIP, host); err != nil {
				if s.onError != "" {
					printMsg(s.name, s.onError)
				}
			} else {
				if s.onSuccess != "" {
					printMsg(s.name, s.onSuccess)
				}
				if s.breakOnNilError {
					cancel()
					break
				}
			}

			if !fast {
				<-ctx.Done()
			}
			cancel()
			prevHostname = hostname
		}
		printMsg(s.name, "Finished")
	}
	printGoodbye()
	return nil
}

func printHeader() {
	fmt.Println("Time      Module   Description")
	fmt.Println("--------------------------------------------------------------------------------")
}

func printMsg(module string, msg ...string) {
	for i := range msg {
		fmt.Printf("%s  %-7s  %s\n", time.Now().Format("15:04:05"), module, msg[i])
	}
}

func printWelcome(ip string) {
	fmt.Printf(`
AlphaSOC Network Flight Simulator™ %s (https://github.com/alphasoc/flightsim)
The IP address of the network interface is %s
The current time is %s

`, Version, ip, time.Now().Format("02-Jan-06 15:04:05"))
}

func printGoodbye() {
	fmt.Printf("\nAll done! Check your SIEM for alerts using the timestamps and details above.\n")
}
