package simulator

import (
	"math/rand"
	"strconv"
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

func (s *DGA) Init(bind BindAddr) error {
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
		tmpLabel := strings.ToLower(utils.RandString(labelLen))
		replaceAt := rand.Intn(labelLen)
		label := tmpLabel[:replaceAt] + strconv.Itoa(rand.Intn(10)) + tmpLabel[replaceAt+1:]
		tld := dgaTLDs[tldIdx[rand.Intn(len(tldIdx))]]
		hosts = append(hosts, label+tld)
	}

	return hosts, nil
}
