package simulator

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
)

// Tor simulator.
type Tor struct {
	TCPConnectSimulator
}

// NewTor creates tor client simulator.
func NewTor() *Tor {
	return &Tor{}
}

// Hosts returns tor exit nodes.
func (s *Tor) Hosts(size int) ([]string, error) {

	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ip := "1.1.1.1"
	if resp.StatusCode == http.StatusOK {
		body, err2 := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err2
		}
		ip = string(body)
	}

	url := fmt.Sprintf("https://check.torproject.org/cgi-bin/TorBulkExitList.py?ip=%s", ip)
	resp, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("check.torproject.org returned %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ips := strings.Split(string(body), "\n")
	// first 3 lines contain comment
	ips = ips[3:]

	var (
		hosts []string
		idx   = rand.Perm(len(ips))
	)
	for n, i := 0, 0; n < len(ips) && i < size; n, i = n+1, i+1 {
		hosts = append(hosts, ips[idx[n]]+":80")
	}

	return hosts, nil
}
