package cmd

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alphasoc/flightsim/simulator"
	"github.com/alphasoc/flightsim/utils"
	"github.com/alphasoc/flightsim/version"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	fast           bool
	ifaceName      string
	simulatorNames = []string{"c2-dns", "c2-ip", "dga", "scan", "spambot", "tunnel"}
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
}

var allsimualtors = []simulatorInfo{
	{
		"c2-dns",
		[]string{"Preparing random sample of current C2 domains"},
		"Resolving %s",
		simulator.NewC2DNS(),
		1 * time.Second,
		false,
	},
	{
		"c2-ip",
		[]string{"Preparing random sample of current C2 IP:port pairs"},
		"Connecting to %s",
		simulator.NewC2IP(),
		1 * time.Second,
		true,
	},
	{
		"dga",
		[]string{"Generating list of DGA domains"},
		"Resolving %s",
		simulator.NewDGA(),
		1 * time.Second,
		false,
	},
	{
		"scan",
		[]string{
			"Preparing random sample of RFC 1918 destinations",
			"Preparing random sample of common TCP destination ports",
		},
		"Port scanning %s",
		simulator.NewPortScan(),
		100 * time.Millisecond,
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
	},
	{
		"tunnel",
		[]string{"Preparing DNS tunnel hostnames"},
		"Resolving %s",
		simulator.NewTunnel(),
		1 * time.Second,
		false,
	},
}

func run(simulators []simulatorInfo, extIP net.IP) error {
	printWelcome(extIP.String())
	printHeader()
	for _, s := range simulators {
		printMsg(s.name, "Starting")
		printMsg(s.name, s.infoHeaders...)

		hosts, err := s.s.Hosts()
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
			s.s.Simulate(ctx, extIP, host)
			if !fast {
				<-ctx.Done()
			}
			cancel()
			prevHostname = hostname
		}
		printMsg(s.name, "Finished")
	}
	printGoodbay()
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

`, version.Version, ip, time.Now().Format("02-Jan-06 15:04:05"))
}

func printGoodbay() {
	fmt.Printf("\nAll done! Check your SIEM for alerts using the timestamps and details above.\n")
}
