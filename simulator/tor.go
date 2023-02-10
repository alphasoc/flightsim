package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TorSimulator struct {
	TCPConnectSimulator
}

func NewTorSimulator() *TorSimulator {
	return &TorSimulator{}
}

// Init initializes the underlying TCPConnectSimulator.
func (s *TorSimulator) Init(bind BindAddr) error {
	return s.TCPConnectSimulator.Init(bind)
}

func (s *TorSimulator) Cleanup() {}

// DetailsResponse contains the Relays slice we're interested in.
type DetailsResponse struct {
	Version         string  `json:"version"`
	BuildRevision   string  `json:"build_revision"`
	RelaysPublished string  `json:"relays_published"`
	Relays          []Relay `json:"relays"`
}

// Relays gets us what we need via OrAddresses.  There is far more information available
// if future needs expand.
type Relay struct {
	Nickname    string   `json:"nickname"`
	Fingerprint string   `json:"fingerprint"`
	OrAddrs     []string `json:"or_addresses"`
}

// Hosts obtains size number of Tor relays using the onionoo.torproject.org API.
func (s *TorSimulator) Hosts(scope string, size int) ([]string, error) {
	// Setup the query such that we get size number of running relays, ordered by consensus
	// weight from largest to smallest.  For details, refer to:
	// https://metrics.torproject.org/onionoo.html#parameters
	// Note also that the 'details' API is queried in order to get the full (ie. ip:port)
	// relay address.
	queryURL := fmt.Sprintf(
		"https://onionoo.torproject.org/details?limit=%v&running=true&order=-consensus_weight",
		size)
	// Allow 5 seconds for the query.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// response body closed automatically on error.
		return nil, err
	}
	defer resp.Body.Close()
	details := DetailsResponse{}
	err = json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		return nil, err
	}
	relays := details.Relays
	// Setup relayAddrs to be returned.  Per
	// https://metrics.torproject.org/onionoo.html#details_relay_or_addresses, the first
	// addr is the primary onion-routing address used during relay registration.
	var relayAddrs []string
	for _, r := range relays {
		// Paranoia: let's make sure the relay address field has at least 1 address.
		if len(r.OrAddrs) > 0 {
			relayAddrs = append(relayAddrs, r.OrAddrs[0])
		}
	}
	return relayAddrs, nil
}
