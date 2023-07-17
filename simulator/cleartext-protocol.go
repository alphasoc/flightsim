package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	HostName = "cleartext.sandbox-services.alphasoc.xyz"
)

// generateRandomData genereates n random bytes
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
	ctx  context.Context
}

// NewCleartextProtocolSimulator creates new instance of CleartextProtocolSimulator
func NewCleartextProtocolSimulator() *CleartextProtocolSimulator {
	return &CleartextProtocolSimulator{}
}

// SendDataToPort sends given data to specified port number.
// It works as a goroutine and returns result via channel
func (cps *CleartextProtocolSimulator) SendDataToPort(port string, data []byte, wg *sync.WaitGroup, results chan error) {
	defer wg.Done()

	address := net.JoinHostPort(HostName, port)
	d := &net.Dialer{LocalAddr: &net.TCPAddr{IP: cps.bind.Addr}}
	conn, err := d.DialContext(cps.ctx, "tcp", address)

	if err != nil {
		results <- err
		return
	}
	defer conn.Close()

	if _, err = conn.Write(data); err != nil {
		results <- err
		return
	}

	if _, err = conn.Read(nil); err != nil {
		results <- err
		return
	}

	results <- nil
}

func (cps *CleartextProtocolSimulator) Init(bind BindAddr) error {
	cps.bind = bind
	return nil
}

func (CleartextProtocolSimulator) Cleanup() {

}

// Simulate cleartext protocol traffic
func (cps *CleartextProtocolSimulator) Simulate(ctx context.Context, dst string) error {
	cps.ctx = ctx

	ports := []string{"21", "23", "110", "143", "873"}

	data := generateRandomData(1000)

	var wg sync.WaitGroup
	results := make(chan error, len(ports))

	for i := range ports {
		fmt.Printf("%s [cleartext] Sending data to port %s\n", time.Now().Format("15:04:05"), ports[i])
		wg.Add(1)
		go cps.SendDataToPort(ports[i], data, &wg, results)
	}
	wg.Wait()
	close(results)

	for result := range results {
		if result != nil {
			return result
		}
	}

	return nil
}

// Hosts returns a domain name of AlphaSOC sandbox
func (CleartextProtocolSimulator) Hosts(scope string, size int) ([]string, error) {
	return []string{HostName}, nil
}
