package wisdom

import (
	"net"
	"strings"
	"testing"
)

func TestWisdomHosts_hosts(t *testing.T) {
	w := NewWisdomHosts("c2", HostTypeDNS)
	size := 5
	h, err := w.Hosts("", size)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != size {
		t.Errorf("expected %d hosts, got %d", size, len(h))
	}

	for n := range h {
		if strings.Contains(h[n], ":") {
			t.Error("FQDN contains colon (has port?): ", h)
		}
	}
}

func TestWisdomHosts_ipWithFamily(t *testing.T) {
	w := NewWisdomHosts("c2", HostTypeIP)
	hosts, err := w.Hosts("trickbot", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) == 0 {
		t.Fatal("no hosts")
	}
	h, p, err := net.SplitHostPort(hosts[0])
	if err != nil {
		t.Error(err)
	}
	if h == "" || p == "" {
		t.Errorf("invalid ip:port value: %s (%s, %s)", hosts[0], h, p)
	}
}
