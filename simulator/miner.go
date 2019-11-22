package simulator

import (
	"bufio"
	"context"
	"net"
	"time"
)

const miningSubscribeBody string = `{"jsonrpc": "2.0", "id": 1, "method": "mining.subscribe", "params": []}` + "\n"
const handshakeTimeout time.Duration = 1 * time.Second

//StratumMiner simulator
type StratumMiner struct {
}

//NewStratumMiner creates new StratumMiner simulator
func NewStratumMiner() *StratumMiner {
	return &StratumMiner{}
}

//Connect to mining pool using stratum protocol. Waits for server response
func stratumHandshakeWithContext(ctx context.Context, conn net.Conn) error {
	_, err := conn.Write([]byte(miningSubscribeBody))
	if err != nil {
		return err
	}
	e := make(chan error)
	go func() {
		_, err = bufio.NewReader(conn).ReadString('\n')
		e <- err
	}()
	select {
	case err := <-e:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

//Simulate connection to mining pool using Stratum protocol
func (m StratumMiner) Simulate(ctx context.Context, bind net.IP, dst string) error {
	d := &net.Dialer{}
	if bind != nil {
		d.LocalAddr = &net.TCPAddr{IP: bind}
	}
	conn, err := d.DialContext(ctx, "tcp", dst)
	if conn != nil {
		ctx, cancel := context.WithTimeout(ctx, handshakeTimeout)
		err = stratumHandshakeWithContext(ctx, conn)
		cancel()
		conn.Close()
	}

	if isSoftError(err, "connect: connection refused") {
		return nil
	}
	return err
}
