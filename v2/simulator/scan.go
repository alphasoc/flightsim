package simulator

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"time"
)

var (
	scanPorts = []int{21, 22, 23, 25, 80, 88, 111, 135, 139, 143, 389, 443, 445, 1433, 1521, 3306, 3389, 5432, 5900, 6000, 8443}

	// IP ranges are for TEST-NET-[123] networks, as per RFC 5737.
	// These IPs should be safe to scan as they're not assigned.
	scanIPRanges = []*net.IPNet{
		{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.CIDRMask(24, 32),
		},
		{
			IP:   net.IPv4(198, 51, 100, 0),
			Mask: net.CIDRMask(24, 32),
		},
		{
			IP:   net.IPv4(203, 0, 113, 0),
			Mask: net.CIDRMask(24, 32),
		},
	}
)

func randIP(network *net.IPNet) net.IP {
	randIP := make(net.IP, len(network.Mask))
	rand.Read(randIP)

	// reverse mask and map randIP, so it only contains bits
	// which can be randomized
	mask := make(net.IPMask, len(network.Mask))
	for n := range mask {
		mask[n] = ^network.Mask[n]
	}
	randIP = randIP.Mask(mask)

	netIP := network.IP.To16()[16-len(randIP):]
	for n := range randIP {
		randIP[n] = randIP[n] | netIP[n]
	}

	return randIP
}

// PortScan simulator.
type PortScan struct {
	tcp TCPConnectSimulator
}

// NewPortScan creates port scan simulator.
func NewPortScan() *PortScan {
	return &PortScan{}
}

func (s *PortScan) Init(bind BindAddr) error {
	return s.tcp.Init(bind)
}

func (PortScan) Cleanup() {
}

// Hosts returns host:port generated from RFC 5737 addresses.
func (s *PortScan) Hosts(scope string, size int) ([]string, error) {
	var hosts []string

	// TODO: make caller responsible for deduplication
	dedup := make(map[string]bool)

	numOfNets := size/20 + 1
	if numOfNets > len(scanIPRanges) {
		numOfNets = len(scanIPRanges)
	}
	netIdx := rand.Perm(len(scanIPRanges))[:numOfNets]

	for k := 0; k < 2*size && len(hosts) < size; k++ {
		// random IP from one of the defined IP ranges
		// TODO: skip IPs ending with zero?
		ip := randIP(scanIPRanges[netIdx[len(hosts)%len(netIdx)]]).String()

		if dedup[ip] {
			continue
		}
		dedup[ip] = true

		hosts = append(hosts, ip)
	}

	sort.Slice(hosts, func(i, j int) bool {
		return bytes.Compare(net.ParseIP(hosts[i]), net.ParseIP(hosts[j])) < 0
	})
	return hosts, nil
}

func (s *PortScan) Simulate(ctx context.Context, dst string) error {
	callTimeout := 200 * time.Millisecond
	// If deadline set, divide the global timeout across every call (port)
	if d, ok := ctx.Deadline(); ok {
		callTimeout = d.Sub(time.Now()) / time.Duration(len(scanPorts))
	}
	// TODO: allow for multiple connection in parallel and hence a longer deadline

	for _, port := range scanPorts {
		ctx, cancelFn := context.WithTimeout(ctx, callTimeout)
		defer cancelFn()
		err := s.tcp.Simulate(ctx, fmt.Sprintf("%s:%d", dst, port))
		if err != nil {
			return err
		}
		// wait until context done
		<-ctx.Done()
	}

	return nil
}
