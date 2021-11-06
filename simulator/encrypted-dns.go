package simulator

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alphasoc/flightsim/simulator/encdns"
	"github.com/alphasoc/flightsim/simulator/encdns/dnscrypt"
	dnscryptproviders "github.com/alphasoc/flightsim/simulator/encdns/dnscrypt/providers"
	"github.com/alphasoc/flightsim/simulator/encdns/doh"
	dohproviders "github.com/alphasoc/flightsim/simulator/encdns/doh/providers"
	"github.com/alphasoc/flightsim/simulator/encdns/dot"
	dotproviders "github.com/alphasoc/flightsim/simulator/encdns/dot/providers"
	"github.com/alphasoc/flightsim/utils"
)

// Tunnel simulator.
type EncryptedDNS struct {
	bind  BindAddr
	Proto encdns.Protocol
}

// NewTunnel creates dns tunnel simulator.
func NewEncryptedDNS() *EncryptedDNS {
	return &EncryptedDNS{}
}

func (s *EncryptedDNS) Init(bind BindAddr) error {
	s.bind = bind
	return nil
}

func (EncryptedDNS) Cleanup() {
}

// HostMsg implements the HostMsgFormatter interface, returning a custom host message
// string to be output by the run command.
func (s *EncryptedDNS) HostMsg(host string) string {
	var protoStr string
	switch s.Proto {
	case encdns.DoH:
		protoStr = "DNS-over-HTTPS"
	case encdns.DoT:
		protoStr = "DNS-over-TLS"
	case encdns.DNSCrypt:
		protoStr = "DNSCrypt"
	}
	return fmt.Sprintf("Simulating Encrypted DNS (%s) via *.%s", protoStr, host)
}

// randomProvider returns a random Protocol p Provider.
func (s *EncryptedDNS) randomProvider(ctx context.Context) encdns.Queryable {
	// If the user has set a bind interface via the -iface flag, have providers use it.
	var bindIP net.IP
	if s.bind.UserSet {
		bindIP = s.bind.Addr
	}
	switch s.Proto {
	case encdns.DoH:
		return dohproviders.NewRandom(ctx, bindIP)
	case encdns.DoT:
		return dotproviders.NewRandom(ctx, bindIP)
	case encdns.DNSCrypt:
		return dnscryptproviders.NewRandom(ctx, bindIP)
	default:
		return nil
	}
}

// Simulate lookups for txt records for give host.
func (s *EncryptedDNS) Simulate(ctx context.Context, host string) error {
	host = utils.FQDN(host)
	// Select a random Provider to be used in this simulation.
	p := s.randomProvider(ctx)
	if p == nil {
		return fmt.Errorf("invalid DNS protocol: unable to select provider")
	}

	for {
		// keep going until the passed context expires
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		label := strings.ToLower(utils.RandString(30))

		ctx, cancelFn := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancelFn()
		resp, err := p.QueryTXT(ctx, fmt.Sprintf("%s.%s", label, host))

		// Ignore timeout.  In case of DoH, when err != nil, resp.Body has already been
		// closed.
		if err != nil {
			if isSoftError(err) {
				continue
			}
			return err
		}
		// Light verification of resp.
		switch s.Proto {
		case encdns.DoH:
			dohResp, err := resp.DOHResponse()
			if err != nil {
				return fmt.Errorf("failed extracting DoH response: %v", err)
			}
			if !doh.IsValidResponse(dohResp) {
				dohResp.Body.Close()
				return fmt.Errorf("bad response: %v", dohResp.Status)
			}
			// All good.  We don't care anymore about the actual response.  Just close it.
			dohResp.Body.Close()
		case encdns.DoT:
			dotResp, err := resp.DOTResponse()
			if err != nil {
				return fmt.Errorf("failed extracting DoT response: %v", err)
			}
			if !dot.IsValidResponse(dotResp) {
				return fmt.Errorf("bad response: %v", dotResp)
			}
			// All good.  We don't care anymore about the actual response
			// (ie. no such host, etc).
			// TODO: If that's not the case, we can add more comprehensive response parsing.
		case encdns.DNSCrypt:
			dnsCryptResp, err := resp.DNSCryptResponse()
			if err != nil {
				return fmt.Errorf("failed extracting DNSCrypt response: %v", err)
			}
			if !dnscrypt.IsValidResponse(dnsCryptResp) {
				return fmt.Errorf("bad response: %v", dnsCryptResp)
			}
		}
		<-ctx.Done()
	}
}

// Hosts returns random generated hosts to alphasoc sandbox.
func (s *EncryptedDNS) Hosts(scope string, size int) ([]string, error) {
	if scope != "" {
		// Protocol parsing (DoH/DoT/etc) from commandline.
		proto, err := encdns.ParseScope(scope)
		if err != nil {
			return []string{}, err
		}
		s.Proto = proto
	} else {
		// Select random Protocol (DoH/DoT/etc) if not specified on the commandline.
		// NOTE: doing this from Hosts() to display in HostMsg().
		s.Proto = encdns.RandomProtocol()
	}
	return []string{"sandbox.alphasoc.xyz"}, nil
}
