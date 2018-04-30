package simulator

import (
	"context"
	"net"
)

// Simulator is an interface for generating hosts and simulating
// traffic for different kind of threaths.
type Simulator interface {
	Simulate(ctx context.Context, extIP net.IP, host string) error
	Hosts(size int) ([]string, error)
}
