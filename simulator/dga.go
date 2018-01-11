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
func (*DGA) Simulate(extIP net.IP, host string) error {
	d := &net.Dialer{
		LocalAddr: &net.UDPAddr{IP: extIP},
	}
	r := &net.Resolver{
		Dial: d.DialContext,
	}

	_, err := r.LookupHost(context.Background(), host)
	return err
}

// Hosts returns random generated dga hosts.
func (t *DGA) Hosts() ([]string, error) {
	const nLookup = 5
	var hosts []string

	idx := rand.Perm(len(tlds))
	for i := 0; i < nLookup; i++ {
		label := strings.ToLower(utils.RandString(7))
		hosts = append(hosts, label+tlds[idx[0]])
		hosts = append(hosts, label+tlds[idx[1]])
		hosts = append(hosts, label+tlds[idx[2]])
	}

	return hosts, nil
}
