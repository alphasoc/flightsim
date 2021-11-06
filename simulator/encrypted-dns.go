package simulator

import (
	"context"
	"fmt"
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
	// TODO: along with issues/39, bind if iface specififed.
	s.bind = bind
	return nil
}

func (EncryptedDNS) Cleanup() {
}

// randomProvider returns a random Protocol p Provider.
func randomProvider(ctx context.Context, p encdns.Protocol) encdns.Queryable {
	switch p {
	case encdns.DoH:
		return dohproviders.NewRandom(ctx)
	case encdns.DoT:
		return dotproviders.NewRandom(ctx)
	case encdns.DNSCrypt:
		return dnscryptproviders.NewRandom(ctx)
	default:
		return nil
	}
}

// Simulate lookups for txt records for give host.
func (s *EncryptedDNS) Simulate(ctx context.Context, host string) error {
	host = utils.FQDN(host)
	// Select random Protocol (DoH/DoT/etc) if not specified on the commandline.
	if s.Proto == encdns.Random {
		s.Proto = encdns.RandomProtocol()
	}
	// Select a random Provider to be used in this simulation.
	p := randomProvider(ctx, s.Proto)
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
		// TODO: Need timeout/dial error check from issues/39
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
			fmt.Println(dnsCryptResp)
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
		s.Proto = encdns.Random
	}
	return []string{"sandbox.alphasoc.xyz"}, nil
}
