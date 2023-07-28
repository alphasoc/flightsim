package simulator

import (
	"context"
	"math/rand"
	"net"
	"time"
)

// generateRandomData genereates n random bytes.
// TODO: this method should be moved to utils along with all the other use cases of generating random data
func generateRandomData(n int) []byte {
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)
	buffer := make([]byte, n)
	_, _ = r.Read(buffer)
	return buffer
}

// CleartextProtocolSimulator simulates cleartext protocol traffic
type CleartextProtocolSimulator struct {
	bind           BindAddr
	TargetHostName string
	TargetIP       string
	Data           []byte
}

// NewCleartextProtocolSimulator creates new instance of CleartextProtocolSimulator
func NewCleartextProtocolSimulator() *CleartextProtocolSimulator {
	const TargetHostName = "cleartext.sandbox-services.alphasoc.xyz"
	return &CleartextProtocolSimulator{TargetHostName: TargetHostName}
}

func (cps *CleartextProtocolSimulator) Init(bind BindAddr) error {
	cps.bind = bind

	ips, err := net.LookupIP(cps.TargetHostName)

	if err != nil {
		return err
	}

	// take the first IP address returned by LookupIP
	cps.TargetIP = ips[0].String()

	// random bytes are generated in Init because it's not necessary
	// to generate them everytime Simulate method is run
	data := generateRandomData(1000)

	cps.Data = data

	return nil
}

func (CleartextProtocolSimulator) Cleanup() {

}

// Simulate cleartext protocol traffic
func (cps *CleartextProtocolSimulator) Simulate(ctx context.Context, dst string) error {
	d := &net.Dialer{LocalAddr: &net.TCPAddr{IP: cps.bind.Addr}}
	conn, err := d.DialContext(ctx, "tcp", dst)

	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err = conn.Write(cps.Data); err != nil {
		return err
	}

	if _, err = conn.Read(nil); err != nil {
		return err
	}

	return nil
}

// Hosts returns a domain name of AlphaSOC sandbox
func (cps *CleartextProtocolSimulator) Hosts(scope string, size int) ([]string, error) {
	var hosts []string

	ports := []string{"21", "23", "110", "143", "873"}

	for _, port := range ports {
		hosts = append(hosts, net.JoinHostPort(cps.TargetIP, port))
	}

	return hosts, nil
}
