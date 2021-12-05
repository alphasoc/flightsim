package simulator

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/alphasoc/flightsim/simulator/encdns"
	"github.com/alphasoc/flightsim/simulator/encdns/dnscrypt"
	"github.com/alphasoc/flightsim/simulator/encdns/doh"
	"github.com/alphasoc/flightsim/simulator/encdns/dot"
	"github.com/alphasoc/flightsim/utils"
	"github.com/alphasoc/flightsim/wisdom"
)

type Protocol string

const DoH = Protocol("doh")
const DoT = Protocol("dot")
const DCr = Protocol("dnscrypt")

var knownProtos = []Protocol{DoH, DoT, DCr}
var knownProtoStrs = []string{string(DoH), string(DoT), string(DCr)}
var protoLongNames = map[Protocol]string{
	DoH: "DNS-over-HTTPS",
	DoT: "DNS-over-TLS",
	DCr: "DNSCrypt",
}

// Encrypted DNS simulator.
type EncryptedDNS struct {
	bind      BindAddr
	protos    []Protocol
	resolvers []encdns.Resolver
}

// NewEncryptedDNS creates an encrypted DNS simulator.
func NewEncryptedDNS() *EncryptedDNS {
	return &EncryptedDNS{}
}

// Init sets the bind address for the simulator.
func (s *EncryptedDNS) Init(bind BindAddr) error {
	s.bind = bind
	return nil
}

// Cleanup does nothing at the moment.
func (EncryptedDNS) Cleanup() {
}

// HostMsg implements the HostMsgFormatter interface, returning a custom host message
// string to be output by the run command.
func (s *EncryptedDNS) HostMsg(host string) string {
	// User has selected a specific protocol on the commandline, else all protocols
	// will be used.
	if len(s.protos) == 1 {
		return fmt.Sprintf(
			"Simulating Encrypted DNS (%v) via *.%s",
			protoLongNames[s.protos[0]],
			host)
	}
	return fmt.Sprintf("Simulating Encrypted DNS via *.%s", host)
}

const numResolversPerProto = 2

// initResolvers initializes resolvers to be used for DNS queries.  The resolvers are
// stored in the s.resolvers slice.  An error is returned.
func (s *EncryptedDNS) initResolvers(ctx context.Context) error {
	var bindIP net.IP
	if s.bind.UserSet {
		bindIP = s.bind.Addr
	}
	for _, p := range s.protos {
		switch p {
		case DoH:
			servers, err := wisdom.EncryptedDNSServers(string(DoH), numResolversPerProto)
			if err != nil {
				return err
			}
			for _, srv := range servers {
				addr := net.JoinHostPort(srv.Domain, strconv.Itoa(srv.Port))
				if srv.Extras.DOHWireProto != "" {
					s.resolvers = append(s.resolvers, doh.NewWireResolver(
						ctx,
						addr,
						srv.Extras.DOHQueryURL,
						srv.Extras.DOHQueryParams,
						srv.Extras.DOHWireProto,
						bindIP))
				} else {
					s.resolvers = append(s.resolvers, doh.NewJSONResolver(
						ctx,
						addr,
						srv.Extras.DOHQueryURL,
						srv.Extras.DOHQueryParams,
						bindIP))
				}
			}
		case DoT:
			servers, err := wisdom.EncryptedDNSServers(string(DoT), numResolversPerProto)
			if err != nil {
				return err
			}
			for _, srv := range servers {
				addr := net.JoinHostPort(srv.Domain, strconv.Itoa(srv.Port))
				s.resolvers = append(s.resolvers, dot.NewResolver(ctx, addr, bindIP))
			}
		case DCr:
			servers, err := wisdom.EncryptedDNSServers(string(DCr), numResolversPerProto)
			if err != nil {
				return err
			}
			for _, srv := range servers {
				s.resolvers = append(
					s.resolvers,
					dnscrypt.NewResolver(ctx, srv.Protocol, srv.Extras.DNSCryptSDNS, bindIP))
			}
		}
	}
	return nil
}

// Simulate lookups for txt records for give host.
func (s *EncryptedDNS) Simulate(ctx context.Context, host string) error {
	host = utils.FQDN(host)
	// At this point we know what protocols we want to use in the simulation.  For each
	// protocol, obtain a set of DNS servers and from them initialize appropriate
	// resolvers.
	err := s.initResolvers(ctx)
	if err != nil {
		return err
	}
	for {
		// Keep going until the passed context expires.
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		// Round robin over the resolvers
		label := strings.ToLower(utils.RandString(30))
		toResolve := fmt.Sprintf("%s.%s", label, host)
		for _, r := range s.resolvers {
			ctx, cancelFn := context.WithTimeout(ctx, 200*time.Millisecond)
			defer cancelFn()
			// Don't actuall care about the responses.
			_, err := r.LookupTXT(ctx, toResolve)
			// Ignore timeout.
			if err != nil {
				if !isSoftError(err) {
					return err
				}
			}
		}
	}
}

func scopeToProto(s string) (Protocol, error) {
	for _, p := range knownProtos {
		if s == string(p) {
			return p, nil
		}
	}
	return Protocol(""), fmt.Errorf("unknown protocol: '%v'", s)
}

// Hosts returns random generated hosts to alphasoc sandbox.
func (s *EncryptedDNS) Hosts(scope string, size int) ([]string, error) {
	if scope != "" {
		p, err := scopeToProto(scope)
		if err != nil {
			return nil, fmt.Errorf(
				"%v: protocol must be one of: %v",
				err,
				strings.Join(knownProtoStrs, ", "))
		}
		s.protos = append(s.protos, p)
	} else {
		s.protos = knownProtos
	}
	return []string{"sandbox.alphasoc.xyz"}, nil
}
