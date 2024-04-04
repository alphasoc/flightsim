package simulator

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"syscall"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const pingCount int = 80
const payloadSize int = 1400

//ICMPtunnel simulator
type ICMPtunnel struct {
	c *icmp.PacketConn
}

//NewICMPtunnel Creates new IMCP tunnel simulator
func NewICMPtunnel() *ICMPtunnel {
	return &ICMPtunnel{}
}

func (s *ICMPtunnel) Init(bind BindAddr) error {
	c, err := icmp.ListenPacket("ip4:icmp", bind.String())
	if err != nil {
		// check if it's syscall error 1: "operation not permitted"
		var errno syscall.Errno
		if errors.As(err, &errno); errno == 1 {
			err = fmt.Errorf("%w (make sure you have sufficient network privileges or try to run as root)", err)
		}
		return err
	}
	s.c = c
	return nil
}

func (s *ICMPtunnel) Cleanup() {
	if s.c != nil {
		s.c.Close()
	}
}

//Hosts returns host used for tunneling
func (ICMPtunnel) Hosts(scope string, size int) ([]string, error) {
	return []string{"icmp.sandbox-services.alphasoc.xyz"}, nil
}

// hostnameToIPv4 does a hostname lookup of host and tries to return the first valid IPv4
// address.  An error is returned on failure.
func hostnameToIPv4(host string) (net.IP, error) {
	ipArr, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	// Find the first valid IPv4 address and use it.
	var ipOfDst net.IP
	for _, ip := range ipArr {
		if ipOfDst = ip.To4(); ipOfDst != nil {
			break
		}
	}
	if ipOfDst == nil {
		return nil, fmt.Errorf("unable to resolve '%s' to a valid IPv4 address", host)
	}
	return ipOfDst, nil
}

//Simulate IMCP tunneling for given dst
func (s *ICMPtunnel) Simulate(ctx context.Context, dst string) error {
	deadline, _ := ctx.Deadline()
	s.c.SetDeadline(deadline)

	for i := 0; i < pingCount; i++ {
		r := make([]byte, payloadSize)
		rand.Read(r)
		data := append([]byte("alphasoc-flightsim:"), r...)

		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: i,
				Data: data,
			},
		}
		binmsg, err := msg.Marshal(nil)
		if err != nil {
			return err
		}
		ipAddr, err := hostnameToIPv4(dst)
		if err != nil {
			return err
		}
		if _, err := s.c.WriteTo(binmsg, &net.IPAddr{IP: net.ParseIP(ipAddr.String())}); err != nil {
			return err
		}

		rb := make([]byte, 1500)
		_, _, err = s.c.ReadFrom(rb)
		if err != nil {
			return err
		}
	}
	return nil
}
