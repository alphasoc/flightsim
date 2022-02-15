package run

import (
	"fmt"
	"time"
)

var Version string

func printHeader() {
	// fmt.Println("Time      Module  Pipeline  Description")
	// fmt.Println("--------------------------------------------------------------------------------")
}

func printMsg(s *Simulation, msg string) {
	if msg == "" {
		return
	}
	// fmt.Printf("%[1]s  %-[3]*.[4]*[2]s %-8[5]s  %[6]s\n",
	// 	time.Now().Format("15:04:05"), s.Name(), 8, len(, s.Pipeline, msg)
	fmt.Printf("%s [%s] %s\n", time.Now().Format("15:04:05"), s.Name(), msg)
}

func printWelcome(ip, dnsIntfIP string) {
	if dnsIntfIP == "" {
		dnsIntfIP = "UNKNOWN, system defaults will be used"
	}
	fmt.Printf(`
AlphaSOC Network Flight Simulator™ %s (https://github.com/alphasoc/flightsim)
The address of the network interface for IP traffic is %s
The address of the network interface for DNS queries is %s
The current time is %s
`, Version, ip, dnsIntfIP, time.Now().Format("02-Jan-06 15:04:05"))
}

func printGoodbye() {
	fmt.Printf("\nAll done! Check your SIEM for alerts using the timestamps and details above.\n")
}
