package simulator

import (
	"bufio"
	"context"
	"net"
)

const miningSubscribeBody string = `{"jsonrpc": "2.0", "id": 1, "method": "mining.subscribe", "params": []}` + "\n"

//StratumMiner simulator
type StratumMiner struct {
	bind BindAddr
}

//NewStratumMiner creates new StratumMiner simulator
func NewStratumMiner() *StratumMiner {
	return &StratumMiner{}
}

func (s *StratumMiner) Init(bind BindAddr) error {
	s.bind = bind
	return nil
}

func (StratumMiner) Cleanup() {
}

//Simulate connection to mining pool using Stratum protocol
func (s *StratumMiner) Simulate(ctx context.Context, dst string) error {
	d := &net.Dialer{LocalAddr: &net.TCPAddr{IP: s.bind.Addr}}
	conn, err := d.DialContext(ctx, "tcp", dst)
	if conn != nil {
		deadline, _ := ctx.Deadline()
		err = conn.SetDeadline(deadline)
		if err != nil {
			return err
		}
		_, err = conn.Write([]byte(miningSubscribeBody))
		if err != nil {
			return err
		}
		_, err = bufio.NewReader(conn).ReadString('\n')
		conn.Close()
	}

	if isSoftError(err, "connect: connection refused") {
		return nil
	}
	return err
}
