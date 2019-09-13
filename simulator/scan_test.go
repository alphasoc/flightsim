package simulator

import "testing"

func TestRandIP(t *testing.T) {
	// Make sure all random IPs are within the network
	for _, network := range scanIPRanges {
		for n := 0; n < 1000; n++ {
			ip := randIP(network)
			if !network.Contains(ip) {
				t.Errorf("%s is not in %s", ip, network)
			}
		}
	}
}

func TestPortScan_Hosts_count(t *testing.T) {
	// expected number of unique hosts;
	// this assumes we've generated all the possible pairs
	expectedCount := 0
	for _, network := range scanIPRanges {
		ones, zeros := network.Mask.Size()
		hosts := (1 << uint(zeros-ones))
		expectedCount += hosts
	}

	ps := NewPortScan()
	hosts, err := ps.Hosts("", expectedCount*100)
	if err != nil {
		t.Fatal(err)
	}
	count := make(map[string]int)
	for n := range hosts {
		count[hosts[n]]++
	}

	if len(count) != expectedCount {
		t.Errorf("expected %d entries, got %d", expectedCount, len(count))
	}
}
