package simulator

import (
	"context"
	"testing"
)

func TestTCPConnectSimulator(t *testing.T) {
	var s TCPConnectSimulator
	err := s.Simulate(context.Background(), nil, "google.com:80")
	t.Log(err)
}

func TestDNSResolveSimulator(t *testing.T) {
	var s DNSResolveSimulator
	err := s.Simulate(context.Background(), nil, "dsfnsfadsfds.com")
	t.Log(err)
}
