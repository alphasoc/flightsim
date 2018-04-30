package simulator

import (
	"context"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// C2DNS simulator.
type C2DNS struct{}

// NewC2DNS creates c2 dns simulator.
func NewC2DNS() *C2DNS {
	return &C2DNS{}
}

// Simulate c2 dns traffic.
func (*C2DNS) Simulate(ctx context.Context, extIP net.IP, host string) error {
	d := &net.Dialer{
		LocalAddr: &net.UDPAddr{IP: extIP},
	}
	r := &net.Resolver{
		Dial: d.DialContext,
	}
	_, err := r.LookupHost(ctx, host)
	return err
}

// Hosts returns hosts marked c2 dns threat.
func (t *C2DNS) Hosts(size int) ([]string, error) {
	resp, err := http.Get("https://cybercrime-tracker.net/all.php")
	if err != nil {
		return nil, errors.Wrapf(err, "cyber crime tracker get http")
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "cyber crime tracker read http body")
	}

	var (
		hosts []string
		c2s   = strings.Split(string(b), "\n")
	)

	if len(c2s) == 0 {
		return hosts, nil
	}

	for len(hosts) < size {
		c2URL := c2s[rand.Intn(len(c2s))]
		u, err := url.Parse("http://" + c2URL)
		if err != nil {
			continue
		}
		// do not include ips
		if net.ParseIP(u.Hostname()) != nil {
			continue
		}
		hosts = append(hosts, u.Hostname())
	}

	return hosts, nil
}
