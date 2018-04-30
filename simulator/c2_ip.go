package simulator

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"

	"github.com/pkg/errors"
)

// C2IP simulator.
type C2IP struct{}

// NewC2IP creates c2 ip simulator.
func NewC2IP() *C2IP {
	return &C2IP{}
}

// Simulate c2 ip traffic.
func (*C2IP) Simulate(ctx context.Context, extIP net.IP, host string) error {
	d := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: extIP},
	}

	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// Hosts returns hosts marked as c2 ip threat.
func (t *C2IP) Hosts(size int) ([]string, error) {
	resp, err := http.Get("https://api.open.wisdom.alphasoc.net/v1/c2-ip")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "api.open.wisdom.alphasoc.net read body error")
	}

	response := &struct {
		Hosts []string `json:"hosts"`
	}{}

	if err := json.Unmarshal(b, response); err != nil {
		return nil, errors.Wrapf(err, "api.opne.wisdom.alphasoc.net parse body error")
	}

	var (
		hosts []string
		idx   = rand.Perm(len(response.Hosts))
	)
	for n, i := 0, 0; n < len(response.Hosts) && i < size; n, i = n+1, i+1 {
		hosts = append(hosts, response.Hosts[idx[n]])
	}
	return hosts, nil
}
