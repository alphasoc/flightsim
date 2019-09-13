package simulator

import (
	"math/rand"
	"strings"

	"github.com/alphasoc/flightsim/utils"
)

var tlds = []string{".com", ".net", ".biz", ".top", ".info", ".xyz", ".space"}

// DGA simulator.
type DGA struct {
	DNSResolveSimulator
}

// NewDGA creates domain generation algorithm simulator
func NewDGA() *DGA {
	return &DGA{}
}

// Hosts returns random generated dga hosts.
func (t *DGA) Hosts(scope string, size int) ([]string, error) {
	var hosts []string

	for i := 0; i < size; i++ {
		label := strings.ToLower(utils.RandString(7))
		hosts = append(hosts, label+tlds[rand.Intn(len(tlds))])
	}

	return hosts, nil
}
