package simulator

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/pkg/errors"
)

// Sinkhole simulator.
type Sinkhole struct {
	TCPConnectSimulator
}

// NewSinkhole creates sinkhole simulator.
func NewSinkhole() *Sinkhole {
	return &Sinkhole{}
}

// Hosts returns hosts marked as sinkhole threat.
func (t *Sinkhole) Hosts(scope string, size int) ([]string, error) {
	resp, err := http.Get("https://api.open.wisdom.alphasoc.net/v1/sinkhole")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "api.open.wisdom.alphasoc.net read body error")
	}

	response := &struct {
		Hosts []string `json:"hosts"`
	}{}

	if err := json.Unmarshal(b, response); err != nil {
		return nil, errors.Wrapf(err, "api.open.wisdom.alphasoc.net parse body error")
	}

	var (
		hosts []string
		idx   = rand.Perm(len(response.Hosts))
	)
	for n, i := 0, 0; n < len(response.Hosts) && i < size; n, i = n+1, i+1 {
		hosts = append(hosts, response.Hosts[idx[n]])
	}
	return hosts, nil
}
