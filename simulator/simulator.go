package simulator

import "net"

// Simulator is an interface for generating hosts and simulating
// traffic for different kind of threaths.
type Simulator interface {
	Simulate(extIP net.IP, host string) error
	Hosts() ([]string, error)
}
