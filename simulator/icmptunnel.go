package simulator

import (
	"context"
	"math/rand"
	"net"
	"os"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const pingCount int = 80
const payloadSize int = 1400

//ICMPtunnel simulator
type ICMPtunnel struct {
}

//NewICMPtunnel Creates new IMCP tunnel simulator
func NewICMPtunnel() *ICMPtunnel {
	return &ICMPtunnel{}
}

func (ICMPtunnel) Init() error {
	return nil
}

func (ICMPtunnel) Cleanup() {
}

//Hosts returns host used for tunneling
func (ICMPtunnel) Hosts(scope string, size int) ([]string, error) {
	return []string{"34.76.148.164"}, nil
}

//Simulate IMCP tunneling for given dst
func (ICMPtunnel) Simulate(ctx context.Context, bind net.IP, dst string) error {
	c, err := icmp.ListenPacket("ip4:icmp", bind.String())
	if err != nil {
		return err
	}
	defer c.Close()

	deadline, _ := ctx.Deadline()
	c.SetDeadline(deadline)

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
		if _, err := c.WriteTo(binmsg, &net.IPAddr{IP: net.ParseIP(dst)}); err != nil {
			return err
		}

		rb := make([]byte, 1500)
		_, _, err = c.ReadFrom(rb)
		if err != nil {
			return err
		}
	}
	return nil
}
