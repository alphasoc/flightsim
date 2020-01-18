package simulator

import (
	"math/rand"
	"net"
	"strings"

	"github.com/alphasoc/flightsim/utils"
)

var dgaTLDs = []string{".com", ".net", ".biz", ".top", ".info", ".xyz", ".space"}

// DGA simulator.
type DGA struct {
	DNSResolveSimulator
}

// NewDGA creates domain generation algorithm simulator
func NewDGA() *DGA {
	return &DGA{}
}

func (s *DGA) Init(bind net.IP) error {
	return s.DNSResolveSimulator.Init(bind)
}

func (DGA) Cleanup() {
}

// Hosts returns random generated dga hosts.
func (t *DGA) Hosts(scope string, size int) ([]string, error) {
	var hosts []string

	// choose three random TLDs
	tldIdx := rand.Perm(len(dgaTLDs))[:3]

	// decide on hostname length (7-10)
	labelLen := 7 + rand.Intn(4)

	for i := 0; i < size; i++ {
		label := strings.ToLower(utils.RandString(labelLen))
		tld := dgaTLDs[tldIdx[rand.Intn(len(tldIdx))]]
		hosts = append(hosts, label+tld)
	}

	return hosts, nil
}
