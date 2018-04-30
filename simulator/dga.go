package simulator

import (
	"context"
	"math/rand"
	"net"
	"strings"

	"github.com/alphasoc/flightsim/utils"
)

var tlds = []string{".com", ".net", ".biz", ".top", ".info", ".xyz", ".space"}

// DGA simulator.
type DGA struct{}

// NewDGA creates domain generation algorithm simulator
func NewDGA() *DGA {
	return &DGA{}
}

// Simulate dga traffic.
func (*DGA) Simulate(ctx context.Context, extIP net.IP, host string) error {
	d := &net.Dialer{
		LocalAddr: &net.UDPAddr{IP: extIP},
	}
	r := &net.Resolver{
		Dial: d.DialContext,
	}

	_, err := r.LookupHost(ctx, host)
	return err
}

// Hosts returns random generated dga hosts.
func (t *DGA) Hosts(size int) ([]string, error) {
	var hosts []string

	idx := rand.Perm(len(tlds))
	for i := 0; i < size; i++ {
		label := strings.ToLower(utils.RandString(7))
		hosts = append(hosts, label+tlds[idx[0]])
		hosts = append(hosts, label+tlds[idx[1]])
		hosts = append(hosts, label+tlds[idx[2]])
	}

	return hosts, nil
}
