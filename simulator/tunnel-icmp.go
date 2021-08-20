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

func (s *ICMPtunnel) Init(bind net.IP) error {
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
	// 104.197.57.232 == sandbox.alphasoc.xyz
	return []string{"104.197.57.232"}, nil
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
		if _, err := s.c.WriteTo(binmsg, &net.IPAddr{IP: net.ParseIP(dst)}); err != nil {
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
