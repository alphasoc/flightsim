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
	bind BindAddr
	data []byte
}

// NewCleartextProtocolSimulator creates new instance of CleartextProtocolSimulator
func NewCleartextProtocolSimulator() *CleartextProtocolSimulator {
	return &CleartextProtocolSimulator{}
}

func (cps *CleartextProtocolSimulator) Init(bind BindAddr) error {
	cps.bind = bind

	// random bytes are generated in Init because it's not necessary
	// to generate them everytime Simulate method is run
	data := generateRandomData(1000)

	cps.data = data

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

	if _, err = conn.Write(cps.data); err != nil {
		return err
	}

	if _, err = conn.Read(nil); err != nil {
		return err
	}

	return nil
}

// Hosts returns IP:port pairs used to connect to AlphaSOC sandbox
func (cps *CleartextProtocolSimulator) Hosts(scope string, size int) ([]string, error) {
	var hosts []string

	ports := []string{"21", "23", "110", "143", "873"}

	ips, err := net.LookupIP("cleartext.sandbox-services.alphasoc.xyz")

	if err != nil {
		return nil, err
	}

	// take the first IP address returned by LookupIP
	targetIP := ips[0].String()

	for _, port := range ports {
		hosts = append(hosts, net.JoinHostPort(targetIP, port))
	}

	return hosts, nil
}
