package simulator

import (
	"context"
	"math/rand"
	"net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

//IMCP tunneling simulator
type ICMPtunnel struct {
}

//Creates new IMCP tunnel simulator
func NewICMPtunnel() *ICMPtunnel {
	return &ICMPtunnel{}
}

func (ICMPtunnel) Init() error {

	return nil
}

func (ICMPtunnel) Cleanup() {
}

//Returns host used for tunneling
func (ICMPtunnel) Hosts(scope string, size int) ([]string, error) {
	//return []string{"sandbox.alphasoc.xyz"}, nil
	return []string{"1.1.1.1"}, nil
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

	data := make([]byte, 1472)
	rand.Read(data)

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
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
	return err
}
